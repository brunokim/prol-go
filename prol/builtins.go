package prol

import (
	"fmt"
	"log"
	"strconv"
)

func unifyBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1, arg2 := goal.Args[0], goal.Args[1]
	return nil, s.Unify(arg1, arg2)
}

func notEqualsBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1, arg2 := goal.Args[0], goal.Args[1]
	unwind := s.Unwind()
	ok := s.Unify(arg1, arg2)
	didBind := unwind()
	return nil, !ok && !didBind
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
	text, err := TermToString(arg1)
	if err != nil {
		log.Printf("chars->atom/2: arg #1: %v", err)
		return nil, false
	}
	return nil, s.Unify(Atom(text), goal.Args[1])
}

func atomLengthBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1 := Deref(goal.Args[0])
	atom, ok := arg1.(Atom)
	if !ok {
		log.Printf("atom_length/2: arg #1: not an atom: %v", arg1)
		return nil, false
	}
	length := Atom(string([]rune{rune(len(atom) + '0')}))
	return nil, s.Unify(length, goal.Args[1])
}

func assertzBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1 := Deref(goal.Args[0])
	clause, err := CompileRule(arg1)
	if err != nil {
		log.Printf("assertz/1: %v", err)
		return nil, false
	}
	log.Println("asserting\n", clause)
	if clause.Functor() == (Functor{"directive", 0}) {
		// Execute directive immediately.
		return clause.(Clause).Body(), true
	}
	s.Assert(clause)
	return nil, true
}

func retractIndexBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1, arg2 := Deref(goal.Args[0]), Deref(goal.Args[1])
	f, ok := arg1.(Struct)
	if !ok || len(f.Args) != 1 {
		log.Println("retract_index_/2: arg #1: not a functor:", arg1)
		return nil, false
	}
	arityAtom, ok := Deref(f.Args[0]).(Atom)
	arity, err := strconv.Atoi(string(arityAtom))
	if !ok || err != nil || arity < 0 {
		log.Println("retract_index_/2: arg #1: not a functor:", arg1)
		return nil, false
	}
	index, ok := arg2.(Atom)
	if !ok {
		log.Println("retract_index_/2: arg #2: not an atom:", arg2)
		return nil, false
	}
	i, err := strconv.Atoi(string(index))
	if err != nil || i <= 0 {
		log.Println("retract_index_/2: arg #2: not a valid number:", index)
		return nil, false
	}
	return nil, s.RetractIndex(Functor{f.Name, arity}, i)
}

func moveClauseInPredicateBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	log.Println("move_clause_in_predicate builtin")
	arg1, arg2, arg3 := Deref(goal.Args[0]), Deref(goal.Args[1]), Deref(goal.Args[2])
	f, ok := arg1.(Struct)
	if !ok || len(f.Args) != 1 {
		log.Println("retract_index_/2: arg #1: not a functor:", arg1)
		return nil, false
	}
	arityAtom, ok := Deref(f.Args[0]).(Atom)
	arity, err := strconv.Atoi(string(arityAtom))
	if !ok || err != nil || arity < 0 {
		log.Println("retract_index_/2: arg #1: not a functor:", arg1)
		return nil, false
	}
	fromAtom, ok := arg2.(Atom)
	if !ok {
		log.Println("retract_index_/2: arg #2: not an atom:", arg2)
		return nil, false
	}
	from, err := strconv.Atoi(string(fromAtom))
	if err != nil {
		log.Println("retract_index_/2: arg #2: not a number:", from)
		return nil, false
	}
	toAtom, ok := arg3.(Atom)
	if !ok {
		log.Println("retract_index_/2: arg #3: not an atom:", arg3)
		return nil, false
	}
	to, err := strconv.Atoi(string(toAtom))
	if err != nil {
		log.Println("retract_index_/2: arg #3: not a number:", to)
		return nil, false
	}
	return nil, s.MoveClauseInPredicate(Functor{f.Name, arity}, from, to)
}

func printBuiltin(s Solver, goal Struct) ([]Struct, bool) {
	arg1 := Deref(goal.Args[0])
	fmt.Println(arg1)
	return nil, true
}

var builtins = []Builtin{
	Builtin{Functor{"=", 2}, unifyBuiltin},
	Builtin{Functor{"neq", 2}, notEqualsBuiltin},
	Builtin{Functor{"atom", 1}, atomBuiltin},
	Builtin{Functor{"var", 1}, varBuiltin},
	Builtin{Functor{"atom->chars", 2}, atomToCharsBuiltin},
	Builtin{Functor{"chars->atom", 2}, charsToAtomBuiltin},
	Builtin{Functor{"atom_to_chars", 2}, atomToCharsBuiltin},
	Builtin{Functor{"chars_to_atom", 2}, charsToAtomBuiltin},
	Builtin{Functor{"atom_length", 2}, atomLengthBuiltin},
	Builtin{Functor{"assertz", 1}, assertzBuiltin},
	Builtin{Functor{"retract_index_", 2}, retractIndexBuiltin},
	Builtin{Functor{"move_clause_in_predicate", 3}, moveClauseInPredicateBuiltin},
	Builtin{Functor{"print", 1}, printBuiltin},
}
