package prol_test

import (
	"slices"
	"testing"

	"github.com/brunokim/prol-go/prol"
	"github.com/google/go-cmp/cmp"
)

var (
	rules = []prol.Rule{
		// query().
		clause(s("query")),
		// nat(0).
		// nat(s(X)) :- nat(X).
		clause(s("nat", a("0"))),
		clause(s("nat", s("s", v("X"))), s("nat", v("X"))),
	}
)

func TestSolve(t *testing.T) {
	tests := []struct {
		name  string
		query prol.Clause
		opts  []any
		want  []prol.Solution
	}{
		{
			"Empty query",
			clause(s("query")),
			nil,
			[]prol.Solution{{}},
		},
		{
			"First 5 natural numbers",
			clause(s("query"), s("nat", v("X"))),
			[]any{"limit", 5},
			[]prol.Solution{
				{"X": a("0")},
				{"X": s("s", a("0"))},
				{"X": s("s", s("s", a("0")))},
				{"X": s("s", s("s", s("s", a("0"))))},
				{"X": s("s", s("s", s("s", s("s", a("0")))))},
			},
		},
	}
	t.Log(prol.NewDatabase(rules...))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.query)
			db := prol.NewDatabase(rules...)
			got := slices.Collect(db.Solve(test.query, test.opts...))
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("(-want, +got): %s", diff)
			}
		})
	}
}
