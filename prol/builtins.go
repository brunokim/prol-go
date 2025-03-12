package prol

import (
	"log"
	"strings"
)

func equalsBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1, arg2 := goal.Args[0], goal.Args[1]
	return nil, s.Unify(arg1, arg2)
}

func atomBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	term := Deref(goal.Args[0])
	_, ok := term.(Atom)
	return nil, ok
}

func varBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	term := Deref(goal.Args[0])
	_, ok := term.(*Ref)
	return nil, ok
}

func atomToCharsBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1 := Deref(goal.Args[0])
	atom, ok := arg1.(Atom)
	if !ok {
		log.Printf("atom->chars/2: arg #1: not an atom: %v", arg1)
		return nil, false
	}
	chars := make([]Term, len(atom))
	for i, ch := range atom {
		chars[i] = Atom(string(ch))
	}
	term := ListToTerm(chars, Atom("[]"))
	return nil, s.Unify(term, goal.Args[1])
}

func charsToAtomBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1 := Deref(goal.Args[0])
	chars, tail := TermToList(arg1)
	if tail != Atom("[]") {
		log.Printf("chars->atom/2: arg #1: not a proper list: %v", arg1)
		return nil, false
	}
	var b strings.Builder
	for _, ch := range chars {
		b.WriteString(string(Deref(ch).(Atom)))
	}
	atom := Atom(b.String())
	return nil, s.Unify(atom, goal.Args[1])
}

func atomLengthBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1 := Deref(goal.Args[0])
	atom, ok := arg1.(Atom)
	if !ok {
		log.Printf("atom_length/2: arg #1: not an atom: %v", arg1)
		return nil, false
	}
	length := Atom(string(len(atom) + '0'))
	return nil, s.Unify(length, goal.Args[1])
}

func assertzBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1 := Deref(goal.Args[0])
	clause, err := compileRule(arg1)
	if err != nil {
		log.Printf("assertz/1: %v", err)
		return nil, false
	}
	s.Assert(clause)
	return nil, true
}

var builtins = []Builtin{
	Builtin{Functor{"=", 2}, equalsBuiltin},
	Builtin{Functor{"atom", 1}, atomBuiltin},
	Builtin{Functor{"var", 1}, varBuiltin},
	Builtin{Functor{"atom->chars", 2}, atomToCharsBuiltin},
	Builtin{Functor{"chars->atom", 2}, charsToAtomBuiltin},
	Builtin{Functor{"atom_length", 2}, atomLengthBuiltin},
	Builtin{Functor{"assertz", 1}, assertzBuiltin},
}
