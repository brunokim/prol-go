package prol

import (
	"log"
	"strings"
)

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
	arg1, arg2 := Deref(goal.Args[0]), Deref(goal.Args[1])
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
	return nil, s.Unify(term, arg2)
}

func charsToAtomBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1, arg2 := Deref(goal.Args[0]), Deref(goal.Args[1])
	chars, tail := TermToList(arg2)
	if tail != Atom("[]") {
		log.Printf("atom->chars/2: arg #2: not a proper list: %v", arg2)
		return nil, false
	}
	var b strings.Builder
	for _, ch := range chars {
		b.WriteString(string(ch.(Atom)))
	}
	atom := Atom(b.String())
	return nil, s.Unify(arg1, atom)
}

func atomLengthBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1, arg2 := Deref(goal.Args[0]), Deref(goal.Args[1])
	atom, ok := arg1.(Atom)
	if !ok {
		log.Printf("atom_length/2: arg #1: not an atom: %v", arg1)
		return nil, false
	}
	length := Atom(string(len(atom)))
	return nil, s.Unify(length, arg2)
}

var builtins = []Builtin{
	Builtin{Functor{"atom", 1}, atomBuiltin},
	Builtin{Functor{"var", 1}, varBuiltin},
	Builtin{Functor{"atom->chars", 2}, atomToCharsBuiltin},
	Builtin{Functor{"chars->atom", 2}, charsToAtomBuiltin},
	Builtin{Functor{"atom_length", 2}, atomLengthBuiltin},
}
