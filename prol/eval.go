package prol

import (
	"fmt"
)

// Eval evaluates an arithmetic expression and returns its result.
func Eval(term Term) (Term, error) {
	term = Deref(term)
	switch t := term.(type) {
	case *Ref:
		return nil, fmt.Errorf("nonground term")
	case Struct:
		args := make([]Term, len(t.Args))
		for i, arg := range t.Args {
			var err error
			args[i], err = Eval(arg)
			if err != nil {
				return nil, fmt.Errorf("%v: arg #%d: %w", t.Indicator(), i+1, err)
			}
		}
		switch t.Indicator() {
		case Indicator{"-", 1}:
			arg1, ok := Deref(args[0]).(Int)
			if !ok {
				return nil, fmt.Errorf("-/1: want -Int, got -%T", args[0])
			}
			return Int(-int(arg1)), nil
		case Indicator{"+", 1}:
			arg1, ok := Deref(args[0]).(Int)
			if !ok {
				return nil, fmt.Errorf("+/1: want +Int, got +%T", args[0])
			}
			return Int(+int(arg1)), nil
		case Indicator{"-", 2}:
			arg1, ok1 := Deref(args[0]).(Int)
			arg2, ok2 := Deref(args[1]).(Int)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("-/2: want Int - Int, got %T - %T", args[0], args[1])
			}
			return Int(int(arg1) - int(arg2)), nil
		case Indicator{"+", 2}:
			arg1, ok1 := Deref(args[0]).(Int)
			arg2, ok2 := Deref(args[1]).(Int)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("+/2: want Int + Int, got %T + %T", args[0], args[1])
			}
			return Int(int(arg1) + int(arg2)), nil
		default:
			return nil, fmt.Errorf("unknown operator: %v", t.Indicator())
		}
	default:
		return term, nil
	}
}
