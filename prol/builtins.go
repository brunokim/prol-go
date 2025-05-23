package prol

import (
	"fmt"
	"os"
	"strconv"
)

func unifyBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1, arg2 := goal.Term.Args[0], goal.Term.Args[1]
	return isSuccess(s.Unify(arg1, arg2))
}

func notEqualsBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1, arg2 := goal.Term.Args[0], goal.Term.Args[1]
	unwind := s.Unwind()
	ok := s.Unify(arg1, arg2)
	didBind := unwind()
	return isSuccess(!ok && !didBind)
}

func gtBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1, arg2 := Deref(goal.Term.Args[0]), Deref(goal.Term.Args[1])
	i1, ok1 := arg1.(Int)
	i2, ok2 := arg2.(Int)
	if !ok1 || !ok2 {
		return isError(fmt.Errorf(">/2: want Int < Int, got %T < %T", arg1, arg2))
	}
	return isSuccess(int(i1) > int(i2))
}

func gteBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1, arg2 := Deref(goal.Term.Args[0]), Deref(goal.Term.Args[1])
	i1, ok1 := arg1.(Int)
	i2, ok2 := arg2.(Int)
	if !ok1 || !ok2 {
		return isError(fmt.Errorf(">=/2: want Int >= Int, got %T >= %T", arg1, arg2))
	}
	return isSuccess(int(i1) >= int(i2))
}

func ltBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1, arg2 := Deref(goal.Term.Args[0]), Deref(goal.Term.Args[1])
	i1, ok1 := arg1.(Int)
	i2, ok2 := arg2.(Int)
	if !ok1 || !ok2 {
		return isError(fmt.Errorf("</2: want Int < Int, got %T < %T", arg1, arg2))
	}
	return isSuccess(int(i1) < int(i2))
}

func lteBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1, arg2 := Deref(goal.Term.Args[0]), Deref(goal.Term.Args[1])
	i1, ok1 := arg1.(Int)
	i2, ok2 := arg2.(Int)
	if !ok1 || !ok2 {
		return isError(fmt.Errorf("=</2: want Int =< Int, got %T =< %T", arg1, arg2))
	}
	return isSuccess(int(i1) <= int(i2))
}

func atomBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	term := Deref(goal.Term.Args[0])
	_, ok := term.(Atom)
	return isSuccess(ok)
}

func intBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	term := Deref(goal.Term.Args[0])
	_, ok := term.(Int)
	return isSuccess(ok)
}

func varBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	term := Deref(goal.Term.Args[0])
	_, ok := term.(*Ref)
	return isSuccess(ok)
}

func atomToCharsBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	atom, ok := arg1.(Atom)
	if !ok {
		return isError(fmt.Errorf("atom_to_chars/2: arg #1: not an atom: %v", arg1))
	}
	chars := make([]Term, len(atom))
	for i, ch := range atom {
		chars[i] = Atom(string(ch))
	}
	term := FromList(chars)
	return isSuccess(s.Unify(term, goal.Term.Args[1]))
}

func charsToAtomBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	text, err := ToString(arg1)
	if err != nil {
		return isError(fmt.Errorf("chars_to_atom/2: arg #1: %w", err))
	}
	return isSuccess(s.Unify(Atom(text), goal.Term.Args[1]))
}

func intToCharsBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
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
	return isSuccess(s.Unify(term, goal.Term.Args[1]))
}

func charsToIntBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	text, err := ToString(arg1)
	if err != nil {
		return isError(fmt.Errorf("chars_to_int/2: arg #1: %w", err))
	}
	i, err := strconv.Atoi(text)
	if err != nil {
		return isError(fmt.Errorf("chars_to_int/2: arg #1: %w", err))
	}
	return isSuccess(s.Unify(Int(i), goal.Term.Args[1]))
}

func atomLengthBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	atom, ok := arg1.(Atom)
	if !ok {
		return isError(fmt.Errorf("atom_length/2: arg #1: not an atom: %v", arg1))
	}
	length := Int(len(atom))
	return isSuccess(s.Unify(length, goal.Term.Args[1]))
}

func getPredicateBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
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
	return isSuccess(s.Unify(x, goal.Term.Args[1]))
}

func putPredicateBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	ind, err := CompileIndicator(arg1)
	if err != nil {
		return isError(fmt.Errorf("put_predicate/2: arg #1: %w", err))
	}
	rulesAST, tail := ToList(Deref(goal.Term.Args[1]))
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

func assertzBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	rule, err := CompileRule(arg1)
	if err != nil {
		return isError(fmt.Errorf("assertz/1: %w", err))
	}
	if rule.Indicator() == (Indicator{"directive", 0}) {
		// Execute directive immediately.
		// TODO: consider other rule types.
		clause := varToRef(rule, map[Var]*Ref{}).(Clause)
		return hasContinuation(clause[1:])
	}
	s.Assert(rule)
	return isSuccess(true)
}

func printBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	fmt.Println(arg1)
	return isSuccess(true)
}

func isBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg2, err := Eval(goal.Term.Args[1])
	if err != nil {
		return isError(fmt.Errorf("is/2: arg #2: %w", err))
	}
	return isSuccess(s.Unify(goal.Term.Args[0], arg2))
}

func consultBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	bs, err := os.ReadFile(string(arg1.(Atom)))
	if err != nil {
		return isError(fmt.Errorf("consult/1: arg #1: %w", err))
	}
	err = s.Interpret(string(bs))
	if err != nil {
		return isError(fmt.Errorf("consult/1: arg #1: %w", err))
	}
	return isSuccess(true)
}

func putBreakpointBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	ind, err := CompileIndicator(arg1)
	if err != nil {
		return isError(fmt.Errorf("put_breakpoint/2: arg #1: %w", err))
	}
	return isSuccess(s.PutBreakpoint(ind))
}

func clearBreakpointBuiltin(s Solver, goal Goal) ([]Goal, bool, error) {
	arg1 := Deref(goal.Term.Args[0])
	ind, err := CompileIndicator(arg1)
	if err != nil {
		return isError(fmt.Errorf("clear_breakpoint/2: arg #1: %w", err))
	}
	return isSuccess(s.ClearBreakpoint(ind))
}

var builtins = []Builtin{
	Builtin{Indicator{"=", 2}, unifyBuiltin},
	Builtin{Indicator{"neq", 2}, notEqualsBuiltin},
	Builtin{Indicator{"\\==", 2}, notEqualsBuiltin},
	Builtin{Indicator{">", 2}, gtBuiltin},
	Builtin{Indicator{">=", 2}, gteBuiltin},
	Builtin{Indicator{"<", 2}, ltBuiltin},
	Builtin{Indicator{"=<", 2}, lteBuiltin},
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
	Builtin{Indicator{"assertz", 1}, assertzBuiltin},
	Builtin{Indicator{"print", 1}, printBuiltin},
	Builtin{Indicator{"is", 2}, isBuiltin},
	Builtin{Indicator{"consult", 1}, consultBuiltin},
	Builtin{Indicator{"put_breakpoint", 1}, putBreakpointBuiltin},
	Builtin{Indicator{"clear_breakpoint", 1}, clearBreakpointBuiltin},
}
