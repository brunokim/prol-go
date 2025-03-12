package prol

import (
	"fmt"
)

func compileRule(ast Term) (Rule, error) {
	ruleAST, err := checkStruct(ast)
	if err != nil {
		return nil, fmt.Errorf("compileRule: %w", err)
	}
	switch ruleAST.Functor() {
	case Functor{"clause", 2}:
		return compileClause(ruleAST)
	default:
		return nil, fmt.Errorf("compileRule: unimplemented rule type: %v", ruleAST.Functor())
	}
}

func compileClause(ast Struct) (Rule, error) {
	arg1, arg2 := Deref(ast.Args[0]), Deref(ast.Args[1])
	headAST, err := checkStruct(arg1)
	if err != nil {
		return nil, fmt.Errorf("clause arg #1: %w", err)
	}
	bodyAST, err := checkProperList(arg2)
	if err != nil {
		return nil, fmt.Errorf("clause arg #2: %w", err)
	}
	head, err := compileStruct(headAST)
	if err != nil {
		return nil, fmt.Errorf("head: %w", err)
	}
	body, err := compileStructs(bodyAST)
	if err != nil {
		return nil, fmt.Errorf("body: %w", err)
	}
	return Clause(append([]Struct{head}, body...)), nil
}

func compileTerm(ast Struct) (Term, error) {
	switch ast.Functor() {
	case Functor{"atom", 1}:
		return compileAtom(ast)
	case Functor{"var", 1}:
		return compileVar(ast)
	case Functor{"struct", 2}:
		return compileStruct(ast)
	default:
		return nil, fmt.Errorf("compileTerm: unimplemented term type: %v", ast.Functor())
	}
}

func compileAtom(ast Struct) (Atom, error) {
	if err := checkFunctor(ast, Functor{"atom", 1}); err != nil {
		return Atom(""), err
	}
	arg1 := Deref(ast.Args[0])
	name, err := checkAtom(arg1)
	if err != nil {
		return Atom(""), fmt.Errorf("name: %w", err)
	}
	return name, nil
}

func compileVar(ast Struct) (Var, error) {
	if err := checkFunctor(ast, Functor{"var", 1}); err != nil {
		return Var(""), err
	}
	arg1 := Deref(ast.Args[0])
	name, err := checkAtom(arg1)
	if err != nil {
		return Var(""), fmt.Errorf("name: %w", err)
	}
	return NewVar(string(name))
}

func compileStruct(ast Struct) (Struct, error) {
	if err := checkFunctor(ast, Functor{"struct", 2}); err != nil {
		return Struct{}, err
	}
	arg1, arg2 := Deref(ast.Args[0]), Deref(ast.Args[1])
	name, err := checkAtom(arg1)
	if err != nil {
		return Struct{}, fmt.Errorf("name: %w", err)
	}
	argsAST, err := checkProperList(arg2)
	if err != nil {
		return Struct{}, fmt.Errorf("args: %w", err)
	}
	args, err := compileTerms(argsAST)
	if err != nil {
		return Struct{}, fmt.Errorf("args: %w", err)
	}
	return Struct{name, args}, nil
}

func compileTerms(ast []Term) ([]Term, error) {
	terms := make([]Term, len(ast))
	for i, termAST := range ast {
		structAST, err := checkStruct(Deref(termAST))
		if err != nil {
			return nil, fmt.Errorf("at #%d: %w", i+1, err)
		}
		term, err := compileTerm(structAST)
		if err != nil {
			return nil, fmt.Errorf("at #%d: %w", i+1, err)
		}
		terms[i] = term
	}
	return terms, nil
}

func compileStructs(ast []Term) ([]Struct, error) {
	structs := make([]Struct, len(ast))
	for i, termAST := range ast {
		structAST, err := checkStruct(Deref(termAST))
		if err != nil {
			return nil, fmt.Errorf("at #%d: %w", i+1, err)
		}
		s, err := compileStruct(structAST)
		if err != nil {
			return nil, fmt.Errorf("at #%d: %w", i+1, err)
		}
		structs[i] = s
	}
	return structs, nil
}

// --- Checks ---

func checkAtom(term Term) (Atom, error) {
	a, ok := term.(Atom)
	if !ok {
		return Atom(""), fmt.Errorf("not an atom")
	}
	return a, nil
}

func checkStruct(term Term) (Struct, error) {
	s, ok := term.(Struct)
	if !ok {
		return Struct{}, fmt.Errorf("not a struct")
	}
	return s, nil
}

func checkFunctor(s Struct, f Functor) error {
	if s.Functor() != f {
		return fmt.Errorf("want functor %v (!= %v)", f, s.Functor())
	}
	return nil
}

func checkProperList(term Term) ([]Term, error) {
	xs, tail := TermToList(term)
	if tail != Atom("[]") {
		return nil, fmt.Errorf("not a proper list")
	}
	return xs, nil
}
