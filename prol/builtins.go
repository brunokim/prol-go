package prol

import (
	"fmt"
	"strconv"
)

func unifyBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1, arg2 := goal.Args[0], goal.Args[1]
	return isSuccess(s.Unify(arg1, arg2))
}

func notEqualsBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1, arg2 := goal.Args[0], goal.Args[1]
	unwind := s.Unwind()
	ok := s.Unify(arg1, arg2)
	didBind := unwind()
	return isSuccess(!ok && !didBind)
}

func atomBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	term := Deref(goal.Args[0])
	_, ok := term.(Atom)
	return isSuccess(ok)
}

func intBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	term := Deref(goal.Args[0])
	_, ok := term.(Int)
	return isSuccess(ok)
}

func varBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	term := Deref(goal.Args[0])
	_, ok := term.(*Ref)
	return isSuccess(ok)
}

func atomToCharsBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1 := Deref(goal.Args[0])
	atom, ok := arg1.(Atom)
	if !ok {
		return isError(fmt.Errorf("atom_to_chars/2: arg #1: not an atom: %v", arg1))
	}
	chars := make([]Term, len(atom))
	for i, ch := range atom {
		chars[i] = Atom(string(ch))
	}
	term := FromList(chars)
	return isSuccess(s.Unify(term, goal.Args[1]))
}

func charsToAtomBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1 := Deref(goal.Args[0])
	text, err := ToString(arg1)
	if err != nil {
		return isError(fmt.Errorf("chars_to_atom/2: arg #1: %w", err))
	}
	return isSuccess(s.Unify(Atom(text), goal.Args[1]))
}

func intToCharsBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1 := Deref(goal.Args[0])
	i, ok := arg1.(Int)
	if !ok {
		return isError(fmt.Errorf("int_to_chars/2: arg #1: not an int: %v", arg1))
	}
	var chars []Term
	if i < 0 {
		chars = append(chars, Atom("-"))
		i = -i
	}
	for i > 0 {
		a, b := i/10, i%10
		chars = append(chars, Atom(strconv.Itoa(int(b))))
		i = a
	}
	term := FromList(chars)
	return isSuccess(s.Unify(term, goal.Args[1]))
}

func charsToIntBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1 := Deref(goal.Args[0])
	text, err := ToString(arg1)
	if err != nil {
		return isError(fmt.Errorf("chars_to_int/2: arg #1: %w", err))
	}
	i, err := strconv.Atoi(text)
	if err != nil {
		return isError(fmt.Errorf("chars_to_int/2: arg #1: %w", err))
	}
	return isSuccess(s.Unify(Int(i), goal.Args[1]))
}

func atomLengthBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1 := Deref(goal.Args[0])
	atom, ok := arg1.(Atom)
	if !ok {
		return isError(fmt.Errorf("atom_length/2: arg #1: not an atom: %v", arg1))
	}
	length := Int(len(atom))
	return isSuccess(s.Unify(length, goal.Args[1]))
}

func getPredicateBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1 := Deref(goal.Args[0])
	ind, err := CompileIndicator(arg1)
	if err != nil {
		return isError(fmt.Errorf("get_predicate/2: %w", err))
	}
	rules := s.GetPredicate(ind)
	terms := make([]Term, len(rules))
	for i, rule := range rules {
		terms[i] = rule.ToAST()
	}
	x := FromList(terms)
	return isSuccess(s.Unify(x, goal.Args[1]))
}

func putPredicateBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1 := Deref(goal.Args[0])
	ind, err := CompileIndicator(arg1)
	if err != nil {
		return isError(fmt.Errorf("put_predicate/2: arg #1: %w", err))
	}
	rulesAST, tail := ToList(Deref(goal.Args[1]))
	if tail != Nil {
		return isError(fmt.Errorf("put_predicate/2: arg #2: not a proper list"))
	}
	rules := make([]Rule, len(rulesAST))
	for i, ruleAST := range rulesAST {
		rules[i], err = CompileRule(Deref(ruleAST))
		if err != nil {
			return isError(fmt.Errorf("put_predicate/2: rule #%d: %w", i+1, err))
		}
	}
	return isSuccess(s.PutPredicate(ind, rules))
}

func printBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg1 := Deref(goal.Args[0])
	fmt.Println(arg1)
	return isSuccess(true)
}

func isBuiltin(s Solver, goal Struct) ([]Struct, bool, error) {
	arg2, err := Eval(goal.Args[1])
	if err != nil {
		return isError(fmt.Errorf("is/2: arg #2: %w", err))
	}
	return isSuccess(s.Unify(goal.Args[0], arg2))
}

var builtins = []Builtin{
	Builtin{Indicator{"=", 2}, unifyBuiltin},
	Builtin{Indicator{"neq", 2}, notEqualsBuiltin},
	Builtin{Indicator{"\\==", 2}, notEqualsBuiltin},
	Builtin{Indicator{"atom", 1}, atomBuiltin},
	Builtin{Indicator{"int", 1}, intBuiltin},
	Builtin{Indicator{"var", 1}, varBuiltin},
	Builtin{Indicator{"atom_to_chars", 2}, atomToCharsBuiltin},
	Builtin{Indicator{"chars_to_atom", 2}, charsToAtomBuiltin},
	Builtin{Indicator{"int_to_chars", 2}, intToCharsBuiltin},
	Builtin{Indicator{"chars_to_int", 2}, charsToIntBuiltin},
	Builtin{Indicator{"atom_length", 2}, atomLengthBuiltin},
	Builtin{Indicator{"get_predicate", 2}, getPredicateBuiltin},
	Builtin{Indicator{"put_predicate", 2}, putPredicateBuiltin},
	Builtin{Indicator{"print", 1}, printBuiltin},
	Builtin{Indicator{"is", 2}, isBuiltin},
}
