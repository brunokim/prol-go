package prol

// --- Private functional constructors ---

func a(name string) Atom {
	return Atom(name)
}

func v(name string) Var {
	return MustVar(name)
}

func s(name Atom, args ...Term) Struct {
	return Struct{name, args}
}

func clause(head Struct, body ...Struct) Clause {
	return append(Clause{head}, body...)
}

func dcg(head Struct, body ...Struct) DCG {
	dcg, err := NewDCG(append([]Struct{head}, body...))
	if err != nil {
		panic(err.Error())
	}
	return dcg
}

func toList(terms ...Term) Term {
	return FromList(terms)
}

func fromString(s string) Term {
	return FromString(s)
}
