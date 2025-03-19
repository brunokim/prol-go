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
	return append(DCG{head}, body...)
}

func toList(terms ...Term) Term {
	return FromList(terms)
}

func fromString(s string) Term {
	return FromString(s)
}
