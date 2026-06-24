package transformer

import (
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// validate parses src and returns the result of Validate. Parse failures abort
// the test so cases can focus on the contextual check.
func validate(t *testing.T, src string) error {
	t.Helper()
	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return Validate(parsed)
}

// TestValidate_NilInput verifies a nil result is accepted (no panic, no error).
func TestValidate_NilInput(t *testing.T) {
	if err := Validate(nil); err != nil {
		t.Errorf("expected nil error for nil input, got %v", err)
	}
}

// TestValidate_ForInsideParallel verifies the canonical valid nesting: a bare
// for inside a parallel region passes.
func TestValidate_ForInsideParallel(t *testing.T) {
	src := `package main

func main() {
	const N = 10
	//gompher parallel
	{
		//gompher for
		for i := 0; i < N; i++ {
			work(i)
		}
	}
}

func work(i int) {}
`
	if err := validate(t, src); err != nil {
		t.Errorf("expected for-inside-parallel to validate, got: %v", err)
	}
}

// TestValidate_WorksharingOutsideParallel verifies that each team construct is
// rejected when it appears at top level, with a message naming the directive.
func TestValidate_WorksharingOutsideParallel(t *testing.T) {
	cases := []struct {
		name string
		src  string
		kind string
	}{
		{
			name: "for",
			src: `package main
func main() {
	const N = 10
	//gompher for
	for i := 0; i < N; i++ { work(i) }
}
func work(i int) {}
`,
			kind: "for",
		},
		{
			name: "single",
			src: `package main
func main() {
	//gompher single
	{ setup() }
}
func setup() {}
`,
			kind: "single",
		},
		{
			name: "master",
			src: `package main
func main() {
	//gompher master
	{ lead() }
}
func lead() {}
`,
			kind: "master",
		},
		{
			name: "barrier",
			src: `package main
func main() {
	a()
	//gompher barrier
	b()
}
func a() {}
func b() {}
`,
			kind: "barrier",
		},
		{
			name: "sections",
			src: `package main
func main() {
	//gompher sections
	{
		//gompher section
		{ a() }
	}
}
func a() {}
`,
			kind: "sections",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(t, tc.src)
			if err == nil {
				t.Fatalf("expected %s outside parallel to be rejected", tc.kind)
			}
			if !strings.Contains(err.Error(), tc.kind) || !strings.Contains(err.Error(), "parallel region") {
				t.Errorf("unexpected error message: %v", err)
			}
		})
	}
}

// TestValidate_CombinedConstructsExempt verifies that the self-contained
// combined constructs are allowed at top level: they provision their own team.
func TestValidate_CombinedConstructsExempt(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "parallel for",
			src: `package main
func main() {
	const N = 10
	//gompher parallel for
	for i := 0; i < N; i++ { work(i) }
}
func work(i int) {}
`,
		},
		{
			name: "parallel sections",
			src: `package main
func main() {
	//gompher parallel sections
	{
		//gompher section
		{ a() }
	}
}
func a() {}
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validate(t, tc.src); err != nil {
				t.Errorf("expected %s to be exempt from nesting validation, got: %v", tc.name, err)
			}
		})
	}
}

// TestValidate_ContextFreeConstructsExempt verifies that critical and atomic,
// which need no team, are allowed at top level.
func TestValidate_ContextFreeConstructsExempt(t *testing.T) {
	src := `package main

func main() {
	var counter int
	//gompher critical
	{
		counter++
	}
	//gompher atomic update
	counter++
	_ = counter
}
`
	if err := validate(t, src); err != nil {
		t.Errorf("expected critical/atomic to be exempt, got: %v", err)
	}
}

// TestValidate_NestedMasterAndBarrier verifies a realistic region with several
// constructs nested in one parallel block all validate together.
func TestValidate_NestedMasterAndBarrier(t *testing.T) {
	src := `package main

func main() {
	const N = 10
	//gompher parallel
	{
		//gompher for
		for i := 0; i < N; i++ { work(i) }

		//gompher single
		{ checkpoint() }

		//gompher barrier

		//gompher master
		{ summary() }
	}
}

func work(i int)   {}
func checkpoint()  {}
func summary()     {}
`
	if err := validate(t, src); err != nil {
		t.Errorf("expected fully-nested region to validate, got: %v", err)
	}
}
