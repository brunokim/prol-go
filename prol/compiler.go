package prol

import (
	"fmt"
)

func CompileRule(ast Term) (Rule, error) {
	ruleAST, err := checkStruct(ast)
	if err != nil {
		return nil, fmt.Errorf("CompileRule: %w", err)
	}
	switch ruleAST.Indicator() {
	case Indicator{"clause", 2}:
		return compileClause(ruleAST)
	case Indicator{"dcg", 2}:
		return compileDCG(ruleAST)
	default:
		return nil, fmt.Errorf("CompileRule: unimplemented rule type: %v", ruleAST.Indicator())
	}
}

func CompileIndicator(ast Term) (Indicator, error) {
	indAST, err := checkStruct(ast)
	if err != nil {
		return Indicator{}, fmt.Errorf("CompileIndicator: %w", err)
	}
	if err := checkIndicator(indAST, Indicator{"indicator", 2}); err != nil {
		return Indicator{}, fmt.Errorf("CompileIndicator: %w", err)
	}
	nameAST, arityAST := Deref(indAST.Args[0]), Deref(indAST.Args[1])
	name, err := checkAtom(nameAST)
	if err != nil {
		return Indicator{}, fmt.Errorf("CompileIndicator: arg #1: %w", err)
	}
	arity, err := checkInt(arityAST)
	if err != nil {
		return Indicator{}, fmt.Errorf("CompileIndicator: arg #2: %w", err)
	}
	return Indicator{name, int(arity)}, nil
}

func compileDCG(ast Struct) (Rule, error) {
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
	return NewDCG(append([]Struct{head}, body...)), nil
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
	switch ast.Indicator() {
	case Indicator{"atom", 1}:
		return compileAtom(ast)
	case Indicator{"int", 1}:
		return compileInt(ast)
	case Indicator{"var", 1}:
		return compileVar(ast)
	case Indicator{"struct", 2}:
		return compileStruct(ast)
	default:
		return nil, fmt.Errorf("compileTerm: unimplemented term type: %v", ast.Indicator())
	}
}

func compileAtom(ast Struct) (Atom, error) {
	if err := checkIndicator(ast, Indicator{"atom", 1}); err != nil {
		return Atom(""), err
	}
	arg1 := Deref(ast.Args[0])
	name, err := checkAtom(arg1)
	if err != nil {
		return Atom(""), fmt.Errorf("name: %w", err)
	}
	return name, nil
}

func compileInt(ast Struct) (Int, error) {
	if err := checkIndicator(ast, Indicator{"int", 1}); err != nil {
		return Int(0), err
	}
	arg1 := Deref(ast.Args[0])
	i, err := checkInt(arg1)
	if err != nil {
		return Int(0), fmt.Errorf("name: %w", err)
	}
	return i, nil
}

func compileVar(ast Struct) (Var, error) {
	if err := checkIndicator(ast, Indicator{"var", 1}); err != nil {
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
	if err := checkIndicator(ast, Indicator{"struct", 2}); err != nil {
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

func checkInt(term Term) (Int, error) {
	i, ok := term.(Int)
	if !ok {
		return Int(0), fmt.Errorf("not an int")
	}
	return i, nil
}

func checkStruct(term Term) (Struct, error) {
	s, ok := term.(Struct)
	if !ok {
		return Struct{}, fmt.Errorf("not a struct")
	}
	return s, nil
}

func checkIndicator(s Struct, f Indicator) error {
	if s.Indicator() != f {
		return fmt.Errorf("want indicator %v (!= %v)", f, s.Indicator())
	}
	return nil
}

func checkProperList(term Term) ([]Term, error) {
	xs, tail := ToList(term)
	if tail != Nil {
		return nil, fmt.Errorf("not a proper list")
	}
	return xs, nil
}
