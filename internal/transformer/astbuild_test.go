package transformer

import (
	"go/ast"
	"go/token"
	"testing"
)

// TestBuildClosure_NoParams verifies that buildClosure emits a FuncLit whose
// parameter list is empty, matching the signature of every runtime entry
// point that takes a `func()` callback (Critical, Single, sections elements).
func TestBuildClosure_NoParams(t *testing.T) {
	body := &ast.BlockStmt{}
	fl := buildClosure(body)

	if fl.Type.Params == nil {
		t.Fatal("expected non-nil Params field")
	}
	if got := len(fl.Type.Params.List); got != 0 {
		t.Errorf("expected zero params, got %d", got)
	}
	if fl.Body != body {
		t.Error("body was not preserved by pointer identity")
	}
}

// TestBuildClosureWithIntParam_NamedInt verifies the helper used by Parallel
// (threadID) and For (iteration index). The single parameter must be named
// as requested and typed as int.
func TestBuildClosureWithIntParam_NamedInt(t *testing.T) {
	body := &ast.BlockStmt{}
	fl := buildClosureWithIntParam(body, "threadID")

	if got := len(fl.Type.Params.List); got != 1 {
		t.Fatalf("expected 1 param, got %d", got)
	}
	param := fl.Type.Params.List[0]
	if len(param.Names) != 1 || param.Names[0].Name != "threadID" {
		t.Errorf("expected param name threadID, got %v", param.Names)
	}
	ident, ok := param.Type.(*ast.Ident)
	if !ok || ident.Name != "int" {
		t.Errorf("expected param type int, got %T %v", param.Type, param.Type)
	}
}

// TestBuildRuntimeCall_StructureAndArgs verifies that the helper produces a
// SelectorExpr of the form `runtime.FuncName(args...)` wrapped in an
// ExprStmt. The selector identifier must match the runtimePkg constant so it
// stays in sync with ensureRuntimeImport.
func TestBuildRuntimeCall_StructureAndArgs(t *testing.T) {
	arg := &ast.Ident{Name: "x"}
	stmt := buildRuntimeCall("Critical", arg)

	call, ok := stmt.X.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", stmt.X)
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr, got %T", call.Fun)
	}
	if id, ok := sel.X.(*ast.Ident); !ok || id.Name != runtimePkg {
		t.Errorf("expected receiver %q, got %v", runtimePkg, sel.X)
	}
	if sel.Sel.Name != "Critical" {
		t.Errorf("expected selector Critical, got %s", sel.Sel.Name)
	}
	if len(call.Args) != 1 || call.Args[0] != arg {
		t.Errorf("expected single arg passed through, got %v", call.Args)
	}
}

// TestBuildStringLit_QuotesValue verifies that the helper produces a STRING
// BasicLit whose value contains the input properly quoted, including escape
// handling for embedded quotes. This is what makes Critical("mylock", ...)
// generate the right token sequence.
func TestBuildStringLit_QuotesValue(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"mylock", `"mylock"`},
		{"", `""`},
		{`he said "hi"`, `"he said \"hi\""`},
	}
	for _, tc := range cases {
		lit := buildStringLit(tc.in)
		if lit.Kind != token.STRING {
			t.Errorf("input %q: expected STRING kind, got %v", tc.in, lit.Kind)
		}
		if lit.Value != tc.want {
			t.Errorf("input %q: expected %q, got %q", tc.in, tc.want, lit.Value)
		}
	}
}
