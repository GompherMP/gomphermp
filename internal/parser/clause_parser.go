package parser

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reVarList   = regexp.MustCompile(`^(private|firstprivate|lastprivate|shared)\(([^)]+)\)`)
	reReduction = regexp.MustCompile(`^reduction\(([+\-*]|&&|\|\||max|min):([^)]+)\)`)
	reSchedule  = regexp.MustCompile(`^schedule\((static|dynamic)(?:,\s*([^)]+))?\)`)
)

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

	return nil, "", fmt.Errorf("unknown clause: %q", text)
}

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
