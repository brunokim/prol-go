package prol_test

import (
	"slices"
	"testing"

	"github.com/brunokim/prol-go/prol"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
		// member(Elem, [H|T]) :- member_(T, Elem, H).
		// member_(_, Elem, Elem).
		// member_([H|T], Elem, _) :- member_(T, Elem, H).
		clause(s("member", v("Elem"), s(".", v("H"), v("T"))),
			s("member_", v("T"), v("Elem"), v("H"))),
		clause(s("member_", v("_"), v("Elem"), v("Elem"))),
		clause(s("member_", s(".", v("H"), v("T")), v("Elem"), v("_")),
			s("member_", v("T"), v("Elem"), v("H"))),
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
		{
			"First 3 lists with 'a'",
			clause(s("query"),
				s("member", a("a"), v("List"))),
			[]any{"limit", 3},
			[]prol.Solution{
				{"List": s(".", a("a"), ref("T"))},
				{"List": s(".", ref("H"), s(".", a("a"), ref("T")))},
				{"List": s(".", ref("H"), s(".", ref("H"), s(".", a("a"), ref("T"))))},
			},
		},
	}
	t.Log(prol.NewDatabase(rules...))

	opts := cmp.Options{
		cmp.AllowUnexported(prol.Ref{}),
		cmpopts.IgnoreFields(prol.Ref{}, "id"),
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.query)
			db := prol.NewDatabase(rules...)
			got := slices.Collect(db.Solve(test.query, test.opts...))
			if diff := cmp.Diff(test.want, got, opts...); diff != "" {
				t.Errorf("(-want, +got): %s", diff)
			}
		})
	}
}
