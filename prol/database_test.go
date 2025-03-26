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
		// complete_me([X|L0], L) :- atom(X), complete_me(L0, L).
		clause(s("complete_me", s(".", v("X"), v("L0")), v("L")),
			s("atom", v("X")),
			s("complete_me", v("L0"), v("L"))),
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
		{
			"Dynamic asserts",
			clause(s("query"),
				s("assertz", s("clause", s("struct", a("bit"), fromList(s("int", int_(0)))), a("[]"))),
				s("assertz", s("clause", s("struct", a("bit"), fromList(s("int", int_(1)))), a("[]"))),
				s("bit", v("X"))),
			nil,
			[]prol.Solution{
				{"X": int_(0)},
				{"X": int_(1)},
			},
		},
		{
			"Clause reflection",
			clause(s("query"),
				s("get_predicate", s("indicator", a("add"), int_(3)), v("Clauses"))),
			nil,
			[]prol.Solution{
				{"Clauses": fromList(
					s("clause",
						s("struct", a("add"),
							fromList(
								s("atom", a("0")),
								s("var", a("X")),
								s("var", a("X")))),
						prol.Nil),
					s("clause",
						s("struct", a("add"),
							fromList(
								s("struct", a("s"), fromList(s("var", a("X")))),
								s("var", a("Y")),
								s("struct", a("s"), fromList(s("var", a("Z")))))),
						fromList(s("struct", a("add"),
							fromList(
								s("var", a("X")),
								s("var", a("Y")),
								s("var", a("Z")))))),
				)},
			},
		},
		{
			"Database manipulation",
			clause(s("query"),
				s("get_predicate", s("indicator", a("complete_me"), int_(2)), fromList(v("_C1"))),
				s("put_predicate", s("indicator", a("complete_me"), int_(2)), fromList(
					s("clause",
						s("struct", a("complete_me"),
							fromList(
								s("var", a("L")),
								s("var", a("L")))),
						prol.Nil),
					v("_C1"))),
				s("complete_me", fromList(a("1"), a("10"), a("100"), int_(1000)), v("Rest"))),
			nil,
			[]prol.Solution{
				{"Rest": fromList(a("1"), a("10"), a("100"), int_(1000))},
				{"Rest": fromList(a("10"), a("100"), int_(1000))},
				{"Rest": fromList(a("100"), int_(1000))},
				{"Rest": fromList(int_(1000))},
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
			seq, ferr := db.Solve(test.query, test.opts...)
			got := slices.Collect(seq)
			if err := ferr(); err != nil {
				t.Errorf("got err: %v", err)
				return
			}
			if diff := cmp.Diff(test.want, got, opts...); diff != "" {
				t.Errorf("(-want, +got): %s", diff)
			}
		})
	}
}
