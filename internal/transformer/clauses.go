package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// dataClausePrefixes turns the private/firstprivate clauses of a structural
// directive into the statements that implement them:
//
//   - closurePrefix runs inside the closure, before the user body:
//       private(x):      var x T      (a fresh per-goroutine copy, zero value)
//       firstprivate(x): x := _fp_x   (shadow x with the captured outer value)
//   - capturePrefix runs in the enclosing block, just before the runtime call:
//       firstprivate(x): _fp_x := x   (snapshot the outer value once)
//
// shared(x) is intentionally ignored: a Go closure already captures by
// reference, which is exactly the shared semantics.
func dataClausePrefixes(file *parser.ParseResult, clauses []parser.Clause, beforePos token.Pos) (closurePrefix, capturePrefix []ast.Stmt, err error) {
	var pvVars, fpVars []string
	for _, c := range clauses {
		switch cl := c.(type) {
		case parser.PrivateClause:
			pvVars = append(pvVars, cl.Vars...)
		case parser.FirstPrivateClause:
			fpVars = append(fpVars, cl.Vars...)
		}
	}

	for _, v := range pvVars {
		typeExpr, terr := resolveVarType(file.FileSet, file.File, v, beforePos)
		if terr != nil {
			return nil, nil, fmt.Errorf("private(%s): %w", v, terr)
		}
		closurePrefix = append(closurePrefix, buildPrivateVarDecl(v, typeExpr))
	}
	if len(fpVars) > 0 {
		closurePrefix = append(closurePrefix, buildFirstprivateShadow(fpVars))
		capturePrefix = append(capturePrefix, buildFirstprivateCapture(fpVars))
	}
	return closurePrefix, capturePrefix, nil
}

// wrapBodyWithPrefix returns the block to use as a closure body: the original
// body when there is no prefix, or a new block with the prefix statements
// prepended. The original body's statements are reused either way.
func wrapBodyWithPrefix(body *ast.BlockStmt, closurePrefix []ast.Stmt) *ast.BlockStmt {
	if len(closurePrefix) == 0 {
		return body
	}
	return &ast.BlockStmt{List: append(closurePrefix, body.List...)}
}

// lastprivateVar holds the data to implement one lastprivate variable: the
// variable is private inside the region (each goroutine has its own copy), and
// the copy produced by the sequentially-last iteration is written back to the
// original through a captured pointer.
type lastprivateVar struct {
	name     string   // the lastprivate variable
	ptrName  string   // the captured pointer (_lp_x)
	typeExpr ast.Expr // resolved type, for the private declaration
}

// lastprivateVars extracts the lastprivate clauses and resolves each variable's
// type.
func lastprivateVars(file *parser.ParseResult, clauses []parser.Clause, beforePos token.Pos) ([]lastprivateVar, error) {
	var out []lastprivateVar
	for _, c := range clauses {
		lc, ok := c.(parser.LastPrivateClause)
		if !ok {
			continue
		}
		for _, v := range lc.Vars {
			typeExpr, terr := resolveVarType(file.FileSet, file.File, v, beforePos)
			if terr != nil {
				return nil, fmt.Errorf("lastprivate(%s): %w", v, terr)
			}
			out = append(out, lastprivateVar{name: v, ptrName: "_lp_" + v, typeExpr: typeExpr})
		}
	}
	return out, nil
}

// buildLastprivateCaptures builds `_lp_x := &x` for each lastprivate variable,
// run before the region so the last iteration can write back through the pointer.
func buildLastprivateCaptures(lpvars []lastprivateVar) []ast.Stmt {
	stmts := make([]ast.Stmt, len(lpvars))
	for i, lp := range lpvars {
		stmts[i] = &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: lp.ptrName}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{buildAddrOf(&ast.Ident{Name: lp.name})},
		}
	}
	return stmts
}

// buildLastprivateDecls builds `var x T` for each lastprivate variable - the
// per-goroutine private copy, shadowing the outer variable.
func buildLastprivateDecls(lpvars []lastprivateVar) []ast.Stmt {
	stmts := make([]ast.Stmt, len(lpvars))
	for i, lp := range lpvars {
		stmts[i] = buildPrivateVarDecl(lp.name, lp.typeExpr)
	}
	return stmts
}

// buildLastprivateWriteback builds `if counter == count-1 { *_lp_x = x }`, the
// statement appended to the loop body so the goroutine that runs the
// sequentially-last iteration copies its private value back to the original.
func buildLastprivateWriteback(lp lastprivateVar, counter string, count ast.Expr) ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: counter},
			Op: token.EQL,
			Y:  &ast.BinaryExpr{X: cloneExpr(count), Op: token.SUB, Y: &ast.BasicLit{Kind: token.INT, Value: "1"}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{buildLastprivateAssign(lp)}},
	}
}

// buildLastprivateAssign builds the bare `*_lp_x = x` write-back, without the
// guard. The loop directive wraps it in an `if counter == count-1`; the sections
// directive uses it directly in the lexically last section (which always runs).
func buildLastprivateAssign(lp lastprivateVar) ast.Stmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.StarExpr{X: &ast.Ident{Name: lp.ptrName}}},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{&ast.Ident{Name: lp.name}},
	}
}

// reductionVar holds everything needed to implement one reduction variable
// using the "pointer capture" technique: take the address of the outer variable
// before the region, shadow it inside with a per-goroutine accumulator
// initialised to the operator's identity, let the body accumulate into the
// shadow unchanged, and fold each goroutine's partial back through the pointer
// under a lock at the end.
type reductionVar struct {
	name     string   // the reduction variable (e.g. "sum")
	ptrName  string   // the captured pointer (_red_sum)
	initName string   // the captured initial value (_init_sum); only for max/min
	typeExpr ast.Expr // resolved type of the variable, for the accumulator
	op       string   // + - * && || max min
}

// reductionVars extracts the reduction clauses and resolves each variable's
// type. All operators in the specification (+ - * && || max min) are supported.
func reductionVars(file *parser.ParseResult, clauses []parser.Clause, beforePos token.Pos) ([]reductionVar, error) {
	var out []reductionVar
	for _, c := range clauses {
		rc, ok := c.(parser.ReductionClause)
		if !ok {
			continue
		}
		if !validReductionOp(rc.Operator) {
			return nil, fmt.Errorf("reduction operator %q not supported (supported: + - * && || max min)", rc.Operator)
		}
		for _, v := range rc.Vars {
			typeExpr, terr := resolveVarType(file.FileSet, file.File, v, beforePos)
			if terr != nil {
				return nil, fmt.Errorf("reduction(%s): %w", v, terr)
			}
			r := reductionVar{name: v, ptrName: "_red_" + v, typeExpr: typeExpr, op: rc.Operator}
			// max/min have no fixed identity; they start from the variable's
			// initial value (captured before the region), which yields the same
			// result as OpenMP's type-extreme identity without needing math
			// constants or type-specific handling.
			if rc.Operator == "max" || rc.Operator == "min" {
				r.initName = "_init_" + v
			}
			out = append(out, r)
		}
	}
	return out, nil
}

func validReductionOp(op string) bool {
	switch op {
	case "+", "-", "*", "&&", "||", "max", "min":
		return true
	}
	return false
}

// buildReductionCaptures builds the statements that run before the region: the
// pointer capture `_red_x := &x` for every reduction, plus `_init_x := x` for
// max/min so their accumulators can start from the original value.
func buildReductionCaptures(rvars []reductionVar) []ast.Stmt {
	var stmts []ast.Stmt
	for _, r := range rvars {
		stmts = append(stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: r.ptrName}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{buildAddrOf(&ast.Ident{Name: r.name})},
		})
		if r.initName != "" {
			stmts = append(stmts, &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: r.initName}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.Ident{Name: r.name}},
			})
		}
	}
	return stmts
}

// buildReductionAccumulators builds `var x T = <identity>` for each reduction,
// shadowing the outer variable with a per-goroutine accumulator initialised to
// the operator's identity.
func buildReductionAccumulators(rvars []reductionVar) []ast.Stmt {
	stmts := make([]ast.Stmt, len(rvars))
	for i, r := range rvars {
		stmts[i] = &ast.DeclStmt{Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{&ast.ValueSpec{
				Names:  []*ast.Ident{{Name: r.name}},
				Type:   r.typeExpr,
				Values: []ast.Expr{reductionIdentity(r)},
			}},
		}}
	}
	return stmts
}

// reductionIdentity returns the per-goroutine accumulator's starting value:
// 0 for + and -, 1 for *, true/false for the boolean operators, and the
// captured initial value for max/min.
func reductionIdentity(r reductionVar) ast.Expr {
	switch r.op {
	case "+", "-":
		return &ast.BasicLit{Kind: token.INT, Value: "0"}
	case "*":
		return &ast.BasicLit{Kind: token.INT, Value: "1"}
	case "&&":
		return &ast.Ident{Name: "true"}
	case "||":
		return &ast.Ident{Name: "false"}
	default: // max, min
		return &ast.Ident{Name: r.initName}
	}
}

// buildReductionCombines builds `runtime.Critical("", func() { <combine> })` for
// each reduction, folding each goroutine's partial into the shared variable
// under mutual exclusion. The combine shape depends on the operator.
func buildReductionCombines(rvars []reductionVar) []ast.Stmt {
	stmts := make([]ast.Stmt, len(rvars))
	for i, r := range rvars {
		body := &ast.BlockStmt{List: []ast.Stmt{reductionCombineStmt(r)}}
		stmts[i] = buildRuntimeCall("Critical", buildStringLit(""), buildClosure(body))
	}
	return stmts
}

// reductionCombineStmt builds the per-operator fold of the per-goroutine
// accumulator (local) into the shared variable through its pointer:
//
//	+ - *   *_red_x op= local
//	&& ||   *_red_x = *_red_x op local
//	max     if local > *_red_x { *_red_x = local }
//	min     if local < *_red_x { *_red_x = local }
func reductionCombineStmt(r reductionVar) ast.Stmt {
	ptr := func() ast.Expr { return &ast.StarExpr{X: &ast.Ident{Name: r.ptrName}} }
	local := func() ast.Expr { return &ast.Ident{Name: r.name} }

	switch r.op {
	case "+":
		return &ast.AssignStmt{Lhs: []ast.Expr{ptr()}, Tok: token.ADD_ASSIGN, Rhs: []ast.Expr{local()}}
	case "-":
		return &ast.AssignStmt{Lhs: []ast.Expr{ptr()}, Tok: token.SUB_ASSIGN, Rhs: []ast.Expr{local()}}
	case "*":
		return &ast.AssignStmt{Lhs: []ast.Expr{ptr()}, Tok: token.MUL_ASSIGN, Rhs: []ast.Expr{local()}}
	case "&&", "||":
		op := token.LAND
		if r.op == "||" {
			op = token.LOR
		}
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ptr()},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{&ast.BinaryExpr{X: ptr(), Op: op, Y: local()}},
		}
	default: // max, min
		cmp := token.GTR
		if r.op == "min" {
			cmp = token.LSS
		}
		return &ast.IfStmt{
			Cond: &ast.BinaryExpr{X: local(), Op: cmp, Y: ptr()},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.AssignStmt{Lhs: []ast.Expr{ptr()}, Tok: token.ASSIGN, Rhs: []ast.Expr{local()}},
			}},
		}
	}
}

// sectionClauseData holds the resolved data-sharing clauses of a sections
// construct. Unlike a loop - which has a single per-goroutine wrapper - sections
// expands one closure per section, so each section needs its own fresh copies of
// the prefix/suffix statements (sharing AST nodes across closures would corrupt
// the printed output). The build* methods regenerate fresh nodes on every call.
type sectionClauseData struct {
	privateVars []privateVar
	fpVars      []string
	rvars       []reductionVar
	lpvars      []lastprivateVar
}

// privateVar pairs a private variable name with its resolved type.
type privateVar struct {
	name     string
	typeExpr ast.Expr
}

// gatherSectionClauses resolves the private/firstprivate/lastprivate/reduction
// clauses of a sections construct into a sectionClauseData.
func gatherSectionClauses(file *parser.ParseResult, clauses []parser.Clause, beforePos token.Pos) (*sectionClauseData, error) {
	d := &sectionClauseData{}
	for _, c := range clauses {
		switch cl := c.(type) {
		case parser.PrivateClause:
			for _, v := range cl.Vars {
				te, err := resolveVarType(file.FileSet, file.File, v, beforePos)
				if err != nil {
					return nil, fmt.Errorf("private(%s): %w", v, err)
				}
				d.privateVars = append(d.privateVars, privateVar{name: v, typeExpr: te})
			}
		case parser.FirstPrivateClause:
			d.fpVars = append(d.fpVars, cl.Vars...)
		}
	}
	rvars, err := reductionVars(file, clauses, beforePos)
	if err != nil {
		return nil, err
	}
	d.rvars = rvars
	lpvars, err := lastprivateVars(file, clauses, beforePos)
	if err != nil {
		return nil, err
	}
	d.lpvars = lpvars
	return d, nil
}

// empty reports whether there are no data-sharing clauses to apply.
func (d *sectionClauseData) empty() bool {
	return len(d.privateVars) == 0 && len(d.fpVars) == 0 && len(d.rvars) == 0 && len(d.lpvars) == 0
}

// captures returns the statements that run once before the Sections call: the
// firstprivate snapshot and the reduction/lastprivate pointer captures. These
// read the outer variables before any section shadows them.
func (d *sectionClauseData) captures() []ast.Stmt {
	var s []ast.Stmt
	if len(d.fpVars) > 0 {
		s = append(s, buildFirstprivateCapture(d.fpVars))
	}
	s = append(s, buildReductionCaptures(d.rvars)...)
	s = append(s, buildLastprivateCaptures(d.lpvars)...)
	return s
}

// sectionPrefix returns fresh statements to prepend inside one section closure:
// private declarations, the firstprivate shadow, lastprivate private copies and
// reduction accumulators. Types are cloned so no two sections share a node.
func (d *sectionClauseData) sectionPrefix() []ast.Stmt {
	var s []ast.Stmt
	for _, p := range d.privateVars {
		s = append(s, buildPrivateVarDecl(p.name, cloneExpr(p.typeExpr)))
	}
	if len(d.fpVars) > 0 {
		s = append(s, buildFirstprivateShadow(d.fpVars))
	}
	s = append(s, buildLastprivateDecls(cloneLastprivateTypes(d.lpvars))...)
	s = append(s, buildReductionAccumulators(cloneReductionTypes(d.rvars))...)
	return s
}

// sectionSuffix returns fresh statements to append inside one section closure:
// the reduction combines for every section, plus the lastprivate handling - the
// write-back in the lexically last section (which always executes), or a blank
// `_ = x` in the others. The blank keeps the private copy "used", since Go rejects
// a variable that is only assigned and never read, which would.
func (d *sectionClauseData) sectionSuffix(isLast bool) []ast.Stmt {
	var s []ast.Stmt
	s = append(s, buildReductionCombines(d.rvars)...)
	for _, lp := range d.lpvars {
		if isLast {
			s = append(s, buildLastprivateAssign(lp))
		} else {
			s = append(s, buildBlankUse(lp.name))
		}
	}
	return s
}

// buildBlankUse builds `_ = name`, used to mark a privatized variable as used so
// the generated code compiles even when the body only writes it.
func buildBlankUse(name string) ast.Stmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "_"}},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{&ast.Ident{Name: name}},
	}
}

// cloneReductionTypes returns a copy of rvars whose type expressions are fresh
// clones, so reduction accumulators built for different sections do not share
// type nodes.
func cloneReductionTypes(rvars []reductionVar) []reductionVar {
	out := make([]reductionVar, len(rvars))
	for i, rv := range rvars {
		out[i] = rv
		out[i].typeExpr = cloneExpr(rv.typeExpr)
	}
	return out
}

// cloneLastprivateTypes returns a copy of lpvars whose type expressions are
// fresh clones, for the same reason as cloneReductionTypes.
func cloneLastprivateTypes(lpvars []lastprivateVar) []lastprivateVar {
	out := make([]lastprivateVar, len(lpvars))
	for i, lp := range lpvars {
		out[i] = lp
		out[i].typeExpr = cloneExpr(lp.typeExpr)
	}
	return out
}
