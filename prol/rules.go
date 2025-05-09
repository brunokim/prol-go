package prol

import (
	"fmt"
	"slices"
	"strings"
)

type Goal struct {
	Term       Struct
	LexerState Term
}

type Rule interface {
	isRule()
	Indicator() Indicator
	ToAST() Term
	Unify(s Solver, goal Goal) (body []Goal, ok bool, err error)
}

//   cont  | success |  error  |       description        |
// --------|---------|---------|--------------------------|
//     nil |   false | non-nil |          error condition |
//     nil |   false |     nil |         failure to unify |
//     nil |    true |     nil |   successful termination |
// non-nil |    true |     nil |  successful continuation |

// --- Clause ---

type Clause []Goal

func (Clause) isRule() {}

// --- DCG ---

type DCG struct {
	dcgGoals []Goal
	clause   Clause
}

func (DCG) isRule() {}

func NewDCG(dcgGoals []Goal) (DCG, error) {
	clause, err := toClause(dcgGoals)
	if err != nil {
		return DCG{}, err
	}
	return DCG{
		dcgGoals: dcgGoals,
		clause:   clause,
	}, nil
}

func toClause(dcgGoals []Goal) (Clause, error) {
	// TODO: use some gensym that prevents conflicts with user-defined variables.
	var c Clause
	head := dcgGoals[0].Term
	c = append(c, Goal{Struct{head.Name, slices.Concat(head.Args, []Term{Var("L0"), Var("L")})}, dcgGoals[0].LexerState})
	var k int
	for i, dcgGoal := range dcgGoals[1:] {
		// List
		s := dcgGoal.Term
		terms, tail := ToList(s)
		if len(terms) > 0 {
			if tail != Nil {
				return nil, fmt.Errorf("invalid DCG: goal #%d: improper list: %v", i+1, dcgGoal)
			}
			curr, next := Var(fmt.Sprintf("L%d", k)), Var(fmt.Sprintf("L%d", k+1))
			c = append(c, Goal{Struct{"=", []Term{curr, FromImproperList(terms, next)}}, dcgGoal.LexerState})
			k++
			continue
		}
		// Empty list
		if s.Name == "[]" && len(s.Args) == 0 {
			continue
		}
		// Embedded code
		if s.Name == "{}" {
			for j, arg := range s.Args {
				goal, ok := arg.(Struct)
				if !ok {
					return nil, fmt.Errorf("invalid DCG: goal #%d: subgoal #%d: not a struct: %v", i+1, j+1, arg)
				}
				// TODO: include lexer state at each embedded goal.
				c = append(c, Goal{goal, dcgGoal.LexerState})
			}
			continue
		}
		// Other grammar rules
		curr, next := Var(fmt.Sprintf("L%d", k)), Var(fmt.Sprintf("L%d", k+1))
		c = append(c, Goal{Struct{s.Name, slices.Concat(s.Args, []Term{curr, next})}, dcgGoal.LexerState})
		k++
	}
	curr := Var(fmt.Sprintf("L%d", k))
	c = append(c, Goal{Struct{"=", []Term{curr, Var("L")}}, dcgGoals[len(dcgGoals)-1].LexerState})
	return c, nil
}

// --- Builtins ---

type Builtin struct {
	indicator Indicator
	unify     func(Solver, Goal) ([]Goal, bool, error)
}

func (Builtin) isRule() {}

// --- Indicator ---

func (c Clause) Indicator() Indicator {
	return c[0].Term.Indicator()
}

func (c DCG) Indicator() Indicator {
	return Indicator{c.dcgGoals[0].Term.Name, len(c.dcgGoals[0].Term.Args) + 2}
}

func (c Builtin) Indicator() Indicator {
	return c.indicator
}

// --- ToAST ---

func (c Clause) ToAST() Term {
	bodyAST := make([]Term, len(c)-1)
	for i, goal := range c[1:] {
		bodyAST[i] = goal.Term.ToAST()
	}
	return Struct{"clause", []Term{c[0].Term.ToAST(), FromList(bodyAST)}}
}

func (c DCG) ToAST() Term {
	bodyAST := make([]Term, len(c.dcgGoals)-1)
	for i, goal := range c.dcgGoals[1:] {
		bodyAST[i] = goal.Term.ToAST()
	}
	return Struct{"clause", []Term{c.dcgGoals[0].Term.ToAST(), FromList(bodyAST)}}
}

func (c Builtin) ToAST() Term {
	return Struct{"builtin", []Term{c.indicator.ToAST()}}
}

// --- Unify ---

func hasContinuation(cont []Goal) ([]Goal, bool, error) {
	return cont, true, nil
}

func isSuccess(ok bool) ([]Goal, bool, error) {
	return nil, ok, nil
}

func isError(err error) ([]Goal, bool, error) {
	return nil, false, err
}

func (c Clause) Unify(s Solver, goal Goal) ([]Goal, bool, error) {
	c = varToRef(c, map[Var]*Ref{}).(Clause)
	if ok := s.Unify(c[0].Term, goal.Term); !ok {
		return isSuccess(false)
	}
	return hasContinuation(c[1:])
}

func (c DCG) Unify(s Solver, goal Goal) ([]Goal, bool, error) {
	return c.clause.Unify(s, goal)
}

func (c Builtin) Unify(s Solver, goal Goal) ([]Goal, bool, error) {
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
		for i, arg := range goal.Args {
			goals[i] = goalString(arg.(Struct), false)
		}
		return fmt.Sprintf("{ %s }", strings.Join(goals, ",\n    "))
	}
	return structToString(goal)
}

func (c Clause) String() string {
	isDCG := false
	head := goalString(c[0].Term, isDCG)
	if len(c) == 1 {
		return fmt.Sprintf("%s.", head)
	}
	body := make([]string, len(c)-1)
	for i, goal := range c[1:] {
		body[i] = goalString(goal.Term, isDCG)
	}
	return fmt.Sprintf("%s :-\n  %s.", head, strings.Join(body, ",\n  "))
}

func (c DCG) String() string {
	isDCG := true
	head := goalString(c.dcgGoals[0].Term, isDCG)
	if len(c.dcgGoals) == 1 {
		return fmt.Sprintf("%s --> [].", head)
	}
	body := make([]string, len(c.dcgGoals)-1)
	for i, goal := range c.dcgGoals[1:] {
		body[i] = goalString(goal.Term, isDCG)
	}
	return fmt.Sprintf("%s -->\n  %s.", head, strings.Join(body, ",\n  "))
}

func (c Builtin) String() string {
	return fmt.Sprintf("%v: <builtin %p>", c.indicator, c.unify)
}
