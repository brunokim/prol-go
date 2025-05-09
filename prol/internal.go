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
	return append(Clause{goal(head)}, goals(body)...)
}

func goal(t Struct) Goal {
	return Goal{Term: t}
}

func goals(ts []Struct) []Goal {
	goals := make([]Goal, len(ts))
	for i, t := range ts {
		goals[i] = goal(t)
	}
	return goals
}

func dcg(head Struct, body ...Struct) DCG {
	dcg, err := NewDCG(append([]Goal{goal(head)}, goals(body)...))
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
