package prol_test

import (
	"github.com/brunokim/prol-go/prol"
)

// --- Private functional constructors ---

func a(name string) prol.Atom {
	return prol.Atom(name)
}

func v(name string) prol.Var {
	return prol.MustVar(name)
}

func s(name prol.Atom, args ...prol.Term) prol.Struct {
	return prol.Struct{name, args}
}

func clause(head prol.Struct, body ...prol.Struct) prol.Clause {
	return append(prol.Clause{head}, body...)
}

func dcg(head prol.Struct, body ...prol.Struct) prol.DCG {
	return append(prol.DCG{head}, body...)
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
