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
		// add(0, X, X).
		// add(s(X), Y, s(Z)) :- add(X, Y, Z).
		clause(s("add", a("0"), v("X"), v("X"))),
		clause(s("add", s("s", v("X")), v("Y"), s("s", v("Z"))),
			s("add", v("X"), v("Y"), v("Z"))),
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
		{
			"All combinations of three numbers that sum to 3",
			clause(s("query"),
				s("add", v("_Tmp"), v("Z"), s("s", s("s", s("s", a("0"))))),
				s("add", v("X"), v("Y"), v("_Tmp"))),
			nil,
			[]prol.Solution{
				{"X": a("0"), "Y": a("0"), "Z": s("s", s("s", s("s", a("0"))))},
				{"X": a("0"), "Y": s("s", a("0")), "Z": s("s", s("s", a("0")))},
				{"X": s("s", a("0")), "Y": a("0"), "Z": s("s", s("s", a("0")))},
				{"X": a("0"), "Y": s("s", s("s", a("0"))), "Z": s("s", a("0"))},
				{"X": s("s", a("0")), "Y": s("s", a("0")), "Z": s("s", a("0"))},
				{"X": s("s", s("s", a("0"))), "Y": a("0"), "Z": s("s", a("0"))},
				{"X": a("0"), "Y": s("s", s("s", s("s", a("0")))), "Z": a("0")},
				{"X": s("s", a("0")), "Y": s("s", s("s", a("0"))), "Z": a("0")},
				{"X": s("s", s("s", a("0"))), "Y": s("s", a("0")), "Z": a("0")},
				{"X": s("s", s("s", s("s", a("0")))), "Y": a("0"), "Z": a("0")},
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
