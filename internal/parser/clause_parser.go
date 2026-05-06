package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Package-level compiled regexes for clause parsing.
// Compiled once at startup for performance. Each regex anchors to the start (^).
var (
	// reVarList covers: private(x, y), firstprivate(x), lastprivate(x), shared(x, y)
	reVarList = regexp.MustCompile(`^(private|firstprivate|lastprivate|shared)\(([^)]+)\)`)

	// reReduction covers: reduction(+:suma), reduction(max:val)
	reReduction = regexp.MustCompile(`^reduction\(([+\-*]|&&|\|\||max|min):([^)]+)\)`)

	// reSchedule covers: schedule(static), schedule(dynamic, 10)
	reSchedule = regexp.MustCompile(`^schedule\((static|dynamic)(?:,\s*([^)]+))?\)`)

	// reDepend covers: depend(in:x, y), depend(out:buff)
	reDepend = regexp.MustCompile(`^depend\((in|out|inout):([^)]+)\)`)

	// reGrainsize covers: grainsize(5)
	reGrainsize = regexp.MustCompile(`^grainsize\(([^)]+)\)`)
)

// extractClauses parses all clauses from a directive string.
// It iteratively consumes the text from left to right until nothing remains.
func extractClauses(text string) ([]Clause, error) {
	if text == "" {
		return nil, nil
	}

	var clauses []Clause
	remaining := strings.TrimSpace(text)

	for remaining != "" {
		clause, rest, err := parseNextClause(remaining)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, clause)
		remaining = strings.TrimSpace(rest)
	}

	return clauses, nil
}

// parseNextClause parses the first clause found at the start of the text.
// Returns the parsed Clause, the remaining unparsed text, and any error.
func parseNextClause(text string) (Clause, string, error) {
	if m := reVarList.FindStringSubmatchIndex(text); m != nil {
		matched := text[m[0]:m[1]]
		rest := text[m[1]:]

		parts := reVarList.FindStringSubmatch(matched)
		kind := parts[1]
		vars := splitVars(parts[2])

		clause, err := makeVarListClause(kind, vars)
		if err != nil {
			return nil, "", err
		}
		return clause, rest, nil
	}

	if m := reReduction.FindStringSubmatchIndex(text); m != nil {
		matched := text[m[0]:m[1]]
		rest := text[m[1]:]

		parts := reReduction.FindStringSubmatch(matched)
		op := parts[1]
		vars := splitVars(parts[2])

		return ReductionClause{Operator: op, Vars: vars}, rest, nil
	}

	if m := reSchedule.FindStringSubmatchIndex(text); m != nil {
		matched := text[m[0]:m[1]]
		rest := text[m[1]:]

		parts := reSchedule.FindStringSubmatch(matched)
		kind := parts[1]
		chunk := strings.TrimSpace(parts[2])

		return ScheduleClause{Kind: kind, Chunk: chunk}, rest, nil
	}

	if m := reDepend.FindStringSubmatchIndex(text); m != nil {
		matched := text[m[0]:m[1]]
		rest := text[m[1]:]

		parts := reDepend.FindStringSubmatch(matched)
		depType := parts[1]
		vars := splitVars(parts[2])

		return DependClause{DepType: depType, Vars: vars}, rest, nil
	}

	if m := reGrainsize.FindStringSubmatchIndex(text); m != nil {
		matched := text[m[0]:m[1]]
		rest := text[m[1]:]

		parts := reGrainsize.FindStringSubmatch(matched)
		size := strings.TrimSpace(parts[1])

		return GrainsizeClause{Size: size}, rest, nil
	}

	return nil, "", fmt.Errorf("unknown clause: %q", text)
}

// makeVarListClause maps a string kind to its corresponding concrete Clause type.
func makeVarListClause(kind string, vars []string) (Clause, error) {
	switch kind {
	case "private":
		return PrivateClause{Vars: vars}, nil
	case "firstprivate":
		return FirstPrivateClause{Vars: vars}, nil
	case "lastprivate":
		return LastPrivateClause{Vars: vars}, nil
	case "shared":
		return SharedClause{Vars: vars}, nil
	default:
		return nil, fmt.Errorf("unknown clause: %q", kind)
	}
}

// splitVars splits a comma-separated variable string into a trimmed slice.
// Example: "x, y,  z" -> []string{"x", "y", "z"}
func splitVars(s string) []string {
	parts := strings.Split(s, ",")
	var vars []string
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			vars = append(vars, v)
		}
	}
	return vars
}
