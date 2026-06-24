package transformer

import (
	"fmt"
	"go/ast"
	"go/token"
)

// loopCounterName is the synthetic 0-based iteration counter introduced when a
// loop is normalized onto the runtime's [0, N) space. It is reserved; loop
// bodies are not expected to reference it (the same assumption the codebase
// makes for threadID and the _fp_/_red_/_lp_ prefixes).
const loopCounterName = "_gompherIter"

// loopForm is a canonical for loop decomposed into the parts needed either to
// emit it directly (the simple form) or to normalize it onto [0, N).
//
// It models the canonical loop:
//
//	for v := lb; v relop b; <step>
//
// where relop is one of <, <=, >, >= and step is v++, v--, v += c or v -= c,
// with the direction of relop and step in agreement. lb, b and c are assumed
// loop-invariant.
type loopForm struct {
	loopVar    string      // induction variable
	lb         ast.Expr    // lower bound (init RHS)
	bound      ast.Expr    // comparison bound b in `v relop b`
	step       ast.Expr    // step magnitude; nil denotes a unit step (++ / --)
	descending bool        // true for the `>`/`>=` + `--`/`-=` direction
	relop      token.Token // LSS, LEQ, GTR or GEQ
}

// simple reports whether the loop is the narrow canonical form
// `for v := 0; v < N; v++`, which the runtime distributes directly over [0, N)
// with no remapping. Only this form takes the fast emission path, keeping the
// common case's generated code identical to a hand-written worksharing loop.
func (f loopForm) simple() bool {
	return f.relop == token.LSS && !f.descending && f.stepIsOne() && isIntLit(f.lb, "0")
}

// stepIsOne reports whether the loop advances by exactly one each iteration.
func (f loopForm) stepIsOne() bool {
	return f.step == nil || isIntLit(f.step, "1")
}

// inclusive reports whether the comparison includes the bound (<= or >=).
func (f loopForm) inclusive() bool {
	return f.relop == token.LEQ || f.relop == token.GEQ
}

// analyzeLoop decomposes forStmt into a loopForm, enforcing the canonical loop
// form GompherMP accepts. Loops outside this form (non-countable, an unexpected
// step, a condition that does not test the induction variable, or a step whose
// direction contradicts the condition) are rejected with a diagnostic.
func analyzeLoop(forStmt *ast.ForStmt) (loopForm, error) {
	var f loopForm

	// init: v := lb
	if forStmt.Init == nil {
		return f, fmt.Errorf("non-canonical loop: missing init statement (want `v := lb`)")
	}
	assign, ok := forStmt.Init.(*ast.AssignStmt)
	if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return f, fmt.Errorf("non-canonical loop: init must be `v := lb`")
	}
	ident, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return f, fmt.Errorf("non-canonical loop: init left side must be an identifier")
	}
	f.loopVar = ident.Name
	f.lb = assign.Rhs[0]

	// cond: v relop b, with the induction variable on the left.
	if forStmt.Cond == nil {
		return f, fmt.Errorf("non-canonical loop: missing condition (want `%s relop b`)", f.loopVar)
	}
	bin, ok := forStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return f, fmt.Errorf("non-canonical loop: condition must compare %q against a bound", f.loopVar)
	}
	if id, ok := bin.X.(*ast.Ident); !ok || id.Name != f.loopVar {
		return f, fmt.Errorf("non-canonical loop: condition must test %q on the left (`%s relop b`)", f.loopVar, f.loopVar)
	}
	switch bin.Op {
	case token.LSS, token.LEQ, token.GTR, token.GEQ:
		f.relop = bin.Op
	default:
		return f, fmt.Errorf("non-canonical loop: condition operator %q not supported (use <, <=, >, >=)", bin.Op)
	}
	f.bound = bin.Y

	// step: v++ | v-- | v += c | v -= c
	switch post := forStmt.Post.(type) {
	case *ast.IncDecStmt:
		if id, ok := post.X.(*ast.Ident); !ok || id.Name != f.loopVar {
			return f, fmt.Errorf("non-canonical loop: step must advance %q", f.loopVar)
		}
		f.descending = post.Tok == token.DEC
		f.step = nil
	case *ast.AssignStmt:
		if len(post.Lhs) != 1 || len(post.Rhs) != 1 {
			return f, fmt.Errorf("non-canonical loop: step must be `%s += c` or `%s -= c`", f.loopVar, f.loopVar)
		}
		if id, ok := post.Lhs[0].(*ast.Ident); !ok || id.Name != f.loopVar {
			return f, fmt.Errorf("non-canonical loop: step must advance %q", f.loopVar)
		}
		switch post.Tok {
		case token.ADD_ASSIGN:
			f.descending = false
		case token.SUB_ASSIGN:
			f.descending = true
		default:
			return f, fmt.Errorf("non-canonical loop: step must be `%s += c` or `%s -= c`", f.loopVar, f.loopVar)
		}
		f.step = post.Rhs[0]
	default:
		return f, fmt.Errorf("non-canonical loop: step must be `%s++`, `%s--`, `%s += c` or `%s -= c`", f.loopVar, f.loopVar, f.loopVar, f.loopVar)
	}

	// The condition and the step must move in the same direction; otherwise the
	// loop is empty or unbounded.
	condDescending := f.relop == token.GTR || f.relop == token.GEQ
	if f.descending != condDescending {
		return f, fmt.Errorf("non-canonical loop: condition %q and step direction disagree", f.relop)
	}
	return f, nil
}

// iterationCount returns the value passed to the runtime as the iteration space
// size: the bare bound for the simple form, or the normalized trip count.
func (f loopForm) iterationCount() ast.Expr {
	if f.simple() {
		return f.bound
	}
	return f.tripCount()
}

// tripCount builds the number of iterations of a normalized loop:
//
//	ascending  (<,  <=): (b - lb)  [+ s] [- 1] [/ s]
//	descending (>,  >=): (lb - b)  [+ s] [- 1] [/ s]
//
// For a unit step the `/ s` is dropped and the count is just (base) or
// (base + 1) for an inclusive bound. Fresh nodes are used throughout so the
// expression can also be reused (e.g. for the lastprivate guard).
func (f loopForm) tripCount() ast.Expr {
	var base ast.Expr // b - lb (ascending) or lb - b (descending)
	if f.descending {
		base = binExpr(cloneExpr(f.lb), token.SUB, cloneExpr(f.bound))
	} else {
		base = binExpr(cloneExpr(f.bound), token.SUB, cloneExpr(f.lb))
	}

	if f.stepIsOne() {
		if f.inclusive() {
			return binExpr(base, token.ADD, intLit("1"))
		}
		return base
	}

	s := atomize(cloneExpr(f.step))
	numer := binExpr(base, token.ADD, s) // base + s
	if !f.inclusive() {
		numer = binExpr(numer, token.SUB, intLit("1")) // base + s - 1
	}
	return binExpr(paren(numer), token.QUO, atomize(cloneExpr(f.step)))
}

// inductionRecovery builds `v := lb ± _gompherIter*step`, the statement
// prepended to the body of a normalized loop so the body sees the user's
// induction variable rather than the 0-based counter.
func (f loopForm) inductionRecovery() ast.Stmt {
	var term ast.Expr = &ast.Ident{Name: loopCounterName}
	if !f.stepIsOne() {
		term = binExpr(&ast.Ident{Name: loopCounterName}, token.MUL, atomize(cloneExpr(f.step)))
	}

	var rhs ast.Expr
	switch {
	case isIntLit(f.lb, "0") && !f.descending:
		rhs = term // lb == 0: i := _gompherIter[*step]
	case f.descending:
		rhs = binExpr(cloneExpr(f.lb), token.SUB, term)
	default:
		rhs = binExpr(cloneExpr(f.lb), token.ADD, term)
	}

	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: f.loopVar}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{rhs},
	}
}

// isIntLit reports whether e is the integer literal with the given value. The
// small expression constructors it pairs with (binExpr, paren, intLit, atomize)
// live with the other AST constructors in astbuild.go.
func isIntLit(e ast.Expr, val string) bool {
	lit, ok := e.(*ast.BasicLit)
	return ok && lit.Kind == token.INT && lit.Value == val
}
