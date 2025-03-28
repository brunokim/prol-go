package prol

import (
	"fmt"
	"slices"
	"strings"
)

type Rule interface {
	isRule()
	Indicator() Indicator
	ToAST() Term
	Unify(s Solver, goal Struct) (body []Struct, ok bool, err error)
}

//   cont  | success |  error  |       description        |
// --------|---------|---------|--------------------------|
//     nil |   false | non-nil |          error condition |
//     nil |   false |     nil |         failure to unify |
//     nil |    true |     nil |   successful termination |
// non-nil |    true |     nil |  successful continuation |

// --- Clause ---

type Clause []Struct

func (Clause) isRule() {}

func (c Clause) Head() Struct {
	return c[0]
}

func (c Clause) Body() []Struct {
	return c[1:]
}

// --- DCG ---

type DCG []Struct

func (DCG) isRule() {}

func (c DCG) toClause() Clause {
	var ss Clause
	ss = append(ss, Struct{c[0].Name, slices.Concat(c[0].Args, []Term{Var("L0"), Var("L")})})
	var i int
	for _, s := range c[1:] {
		// List
		terms, tail := ToList(s)
		if len(terms) > 0 {
			if tail != Nil {
				panic(fmt.Sprintf("invalid DCG"))
			}
			curr, next := Var(fmt.Sprintf("L%d", i)), Var(fmt.Sprintf("L%d", i+1))
			ss = append(ss, Struct{"=", []Term{curr, FromImproperList(terms, next)}})
			i++
			continue
		}
		// Empty list
		if tail == Nil {
			continue
		}
		// Embedded code
		if s.Name == "{}" {
			for _, arg := range s.Args {
				ss = append(ss, arg.(Struct))
			}
			continue
		}
		// Other grammar rules
		curr, next := Var(fmt.Sprintf("L%d", i)), Var(fmt.Sprintf("L%d", i+1))
		ss = append(ss, Struct{s.Name, slices.Concat(s.Args, []Term{curr, next})})
		i++
	}
	curr := Var(fmt.Sprintf("L%d", i))
	ss = append(ss, Struct{"=", []Term{curr, Var("L")}})
	return ss
}

// --- Builtins ---

type Builtin struct {
	indicator Indicator
	unify     func(Solver, Struct) ([]Struct, bool, error)
}

func (Builtin) isRule() {}

// --- Indicator ---

func (c Clause) Indicator() Indicator {
	return c.Head().Indicator()
}

func (c DCG) Indicator() Indicator {
	return Indicator{c[0].Name, len(c[0].Args) + 2}
}

func (c Builtin) Indicator() Indicator {
	return c.indicator
}

// --- ToAST ---

func (c Clause) ToAST() Term {
	bodyAST := make([]Term, len(c)-1)
	for i, goal := range c[1:] {
		bodyAST[i] = goal.ToAST()
	}
	return Struct{"clause", []Term{c[0].ToAST(), FromList(bodyAST)}}
}

func (c DCG) ToAST() Term {
	bodyAST := make([]Term, len(c)-1)
	for i, goal := range c[1:] {
		bodyAST[i] = goal.ToAST()
	}
	return Struct{"clause", []Term{c[0].ToAST(), FromList(bodyAST)}}
}

func (c Builtin) ToAST() Term {
	return Struct{"builtin", []Term{c.indicator.ToAST()}}
}

// --- Unify ---

func hasContinuation(cont []Struct) ([]Struct, bool, error) {
	return cont, true, nil
}

func isSuccess(ok bool) ([]Struct, bool, error) {
	return nil, ok, nil
}

func isError(err error) ([]Struct, bool, error) {
	return nil, false, err
}

func (c Clause) Unify(s Solver, goal Struct) ([]Struct, bool, error) {
	c = varToRef(c, map[Var]*Ref{}).(Clause)
	if ok := s.Unify(c.Head(), goal); !ok {
		return isSuccess(false)
	}
	return hasContinuation(c.Body())
}

func (c DCG) Unify(s Solver, goal Struct) ([]Struct, bool, error) {
	return c.toClause().Unify(s, goal)
}

func (c Builtin) Unify(s Solver, goal Struct) ([]Struct, bool, error) {
	return c.unify(s, goal)
}

// --- String ---

func goalString(goal Struct, isDCG bool) string {
	// Atom.
	if len(goal.Args) == 0 {
		return goal.Name.String()
	}
	// DCG list.
	if isDCG && goal.Indicator() == (Indicator{".", 2}) {
		terms, tail := ToList(goal)
		if tail == Nil {
			return listToString(terms, tail)
		}
	}
	// DCG embedded.
	if isDCG && goal.Name == "{}" {
		goals := make([]string, len(goal.Args))
		for i, goal := range goal.Args {
			goals[i] = goalString(goal.(Struct) /*isDCG*/, false)
		}
		return fmt.Sprintf("{ %s }", strings.Join(goals, ",\n    "))
	}
	return structToString(goal)
}

func (c Clause) String() string {
	head := goalString(c.Head() /*isDCG*/, false)
	if len(c) == 1 {
		return fmt.Sprintf("%s.", head)
	}
	body := make([]string, len(c)-1)
	for i, goal := range c.Body() {
		body[i] = goalString(goal /*isDCG*/, false)
	}
	return fmt.Sprintf("%s :-\n  %s.", head, strings.Join(body, ",\n  "))
}

func (c DCG) String() string {
	head := goalString(c[0] /*isDCG*/, true)
	if len(c) == 1 {
		return fmt.Sprintf("%s --> [].", head)
	}
	body := make([]string, len(c)-1)
	for i, s := range c[1:] {
		body[i] = goalString(s /*isDCG*/, true)
	}
	return fmt.Sprintf("%s -->\n  %s.", head, strings.Join(body, ",\n  "))
}

func (c Builtin) String() string {
	return fmt.Sprintf("%v: <builtin %p>", c.indicator, c.unify)
}
