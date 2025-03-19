package prol

import (
	"fmt"
	"slices"
	"strings"
)

type Rule interface {
	isRule()
	Functor() Functor
	Unify(s Solver, goal Struct) (body []Struct, ok bool)
}

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
			if tail != Atom("[]") {
				panic(fmt.Sprintf("invalid DCG"))
			}
			curr, next := Var(fmt.Sprintf("L%d", i)), Var(fmt.Sprintf("L%d", i+1))
			ss = append(ss, Struct{"=", []Term{curr, FromImproperList(terms, next)}})
			i++
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
	functor Functor
	unify   func(Solver, Struct) ([]Struct, bool)
}

func (Builtin) isRule() {}

// --- Functor ---

func (c Clause) Functor() Functor {
	return c.Head().Functor()
}

func (c DCG) Functor() Functor {
	return Functor{c[0].Name, len(c[0].Args) + 2}
}

func (c Builtin) Functor() Functor {
	return c.functor
}

// --- Unify ---

func (c Clause) Unify(s Solver, goal Struct) ([]Struct, bool) {
	c = varToRef(c, map[Var]*Ref{}).(Clause)
	if !s.Unify(c.Head(), goal) {
		return nil, false
	}
	return c.Body(), true
}

func (c DCG) Unify(s Solver, goal Struct) ([]Struct, bool) {
	return c.toClause().Unify(s, goal)
}

func (c Builtin) Unify(s Solver, goal Struct) ([]Struct, bool) {
	return c.unify(s, goal)
}

// --- String ---

func (c Clause) String() string {
	if len(c) == 1 {
		return fmt.Sprintf("%s.", c.Head())
	}
	body := make([]string, len(c)-1)
	for i, goal := range c.Body() {
		body[i] = goal.String()
	}
	return fmt.Sprintf("%s :-\n  %s.", c.Head(), strings.Join(body, ",\n  "))
}

func (c DCG) String() string {
	if len(c) == 1 {
		return fmt.Sprintf("%v --> [].", c[0])
	}
	body := make([]string, len(c)-1)
	for i, s := range c[1:] {
		body[i] = s.String()
	}
	return fmt.Sprintf("%v -->\n  %s.", c[0], strings.Join(body, ",\n  "))
}

func (c Builtin) String() string {
	return fmt.Sprintf("%v: <builtin %p>", c.functor, c.unify)
}
