package prol_test

import (
	"github.com/brunokim/prol-go/prol"
)

// --- Private functional constructors ---

func a(name string) prol.Atom {
	return prol.Atom(name)
}

func int_(i int) prol.Int {
	return prol.Int(i)
}

func v(name string) prol.Var {
	return prol.MustVar(name)
}

func s(name prol.Atom, args ...prol.Term) prol.Struct {
	return prol.Struct{name, args}
}

func ref(name string) *prol.Ref {
	return prol.NewRef(v(name))
}

func goal(t prol.Struct) prol.Goal {
	return prol.Goal{Term: t}
}

func goals(ts []prol.Struct) []prol.Goal {
	goals := make([]prol.Goal, len(ts))
	for i, t := range ts {
		goals[i] = goal(t)
	}
	return goals
}

func clause(head prol.Struct, body ...prol.Struct) prol.Clause {
	return append(prol.Clause{goal(head)}, goals(body)...)
}

func dcg(head prol.Struct, body ...prol.Struct) prol.DCG {
	dcg, err := prol.NewDCG(append([]prol.Goal{goal(head)}, goals(body)...))
	if err != nil {
		panic(err.Error())
	}
	return dcg
}

func toList(term prol.Term) ([]prol.Term, prol.Term) {
	return prol.ToList(term)
}

func fromList(terms ...prol.Term) prol.Term {
	return prol.FromList(terms)
}

func fromString(s string) prol.Term {
	return prol.FromString(s)
}

func toString(t prol.Term) (string, error) {
	return prol.ToString(t)
}
