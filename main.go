package main

import (
	"fmt"
)

type Term interface {
	isTerm()
}

type Atom string

type Var string

type Struct struct {
	Name Atom
	Args []Term
}

func (s Struct) Functor() Functor {
	return Functor{s.Name, len(s.Args)}
}

type Functor struct {
	Name  Atom
	Arity int
}

type Ref struct {
	Name  Var
	Id    int
	Value Term
}

var (
	refID = 0
)

func (Atom) isTerm()   {}
func (Var) isTerm()    {}
func (Struct) isTerm() {}
func (*Ref) isTerm()   {}

func deref(t Term) Term {
	if ref, ok := t.(*Ref); ok && ref.Value != nil {
		t = ref.Value
	}
	return t
}

type Clause []Struct

func (c Clause) Head() Struct {
	return c[0]
}

func (c Clause) Body() []Struct {
	return c[1:]
}

func varToRef(x any, env map[Var]*Ref) any {
	switch v := x.(type) {
	case Clause:
		y := make(Clause, len(v))
		for i, goal := range v {
			y[i] = varToRef(goal, env).(Struct)
		}
		return y
	case Struct:
		y := Struct{v.Name, make([]Term, len(v.Args))}
		for i, arg := range v.Args {
			y.Args[i] = varToRef(arg, env).(Term)
		}
		return y
	case Var:
		if v == "_" {
			refID++
			return &Ref{v, refID, nil}
		}
		if _, ok := env[v]; !ok {
			refID++
			env[v] = &Ref{v, refID, nil}
		}
		return env[v]
	default:
		return x
	}
}

func refToTerm(x Term) Term {
	x = deref(x)
	if s, ok := x.(Struct); ok {
		args := make([]Term, len(s.Args))
		for i, arg := range s.Args {
			args[i] = refToTerm(arg)
		}
		return Struct{s.Name, args}
	}
	return x
}

type Solver struct {
	clauses    map[Functor][]Clause
	trail      []*Ref
	onSolution func() bool
}

func (s *Solver) Solve(query []Struct, onSolution func(map[Var]Term) bool) {
	env := make(map[Var]*Ref)
	query = []Struct(varToRef(Clause(query), env).(Clause))
	s.trail = nil
	s.onSolution = func() bool {
		m := make(map[Var]Term)
		for x, ref := range env {
			if x[0] == '_' {
				continue
			}
			m[x] = refToTerm(ref)
		}
		return onSolution(m)
	}
	s.dfs(query)
	s.trail = nil
}

func (s *Solver) dfs(goals []Struct) bool {
	if len(goals) == 0 {
		return s.onSolution()
	}
	goal, rest := goals[0], goals[1:]
	n := len(s.trail)
	for _, clause := range s.clauses[goal.Functor()] {
		clause = varToRef(clause, map[Var]*Ref{}).(Clause)
		if s.unify(clause.Head(), goal) {
			if !s.dfs(append(clause.Body(), rest...)) {
				return false
			}
		}
		for _, ref := range s.trail[n:] {
			ref.Value = nil
		}
		s.trail = s.trail[:n]
	}
	return true
}

func (s *Solver) unify(t1, t2 Term) bool {
	t1, t2 = deref(t1), deref(t2)
	s1, isStruct1 := t1.(Struct)
	s2, isStruct2 := t2.(Struct)
	if isStruct1 && isStruct2 {
		if s1.Name != s2.Name || len(s1.Args) != len(s2.Args) {
			return false
		}
		for i := 0; i < len(s1.Args); i++ {
			if !s.unify(s1.Args[i], s2.Args[i]) {
				return false
			}
		}
		return true
	}
	if t1 == t2 {
		return true
	}
	if ref1, ok := t1.(*Ref); ok {
		ref1.Value = t2
		s.trail = append(s.trail, ref1)
		return true
	}
	if ref2, ok := t2.(*Ref); ok {
		ref2.Value = t1
		s.trail = append(s.trail, ref2)
		return true
	}
	return false
}

func st(name string, terms ...Term) Struct {
	return Struct{Atom(name), terms}
}

func main() {
	solver := Solver{
		clauses: map[Functor][]Clause{
			Functor{"nat", 1}: []Clause{
				Clause{st("nat", Atom("0"))},
				Clause{
					st("nat", st("s", Var("X"))),
					st("nat", Var("X")),
				},
			},
			Functor{"add", 3}: []Clause{
				Clause{st("add", Atom("0"), Var("X"), Var("X"))},
				Clause{
					st("add", st("s", Var("A")), Var("B"), st("s", Var("C"))),
					st("add", Var("A"), Var("B"), Var("C")),
				},
			},
		},
	}
	fmt.Println("First 5 natural numbers:")
	cnt := 1
	solver.Solve([]Struct{st("nat", Var("X"))}, func(m map[Var]Term) bool {
		fmt.Println(m)
		if cnt >= 5 {
			return false
		}
		cnt++
		return true
	})
	fmt.Println()
	fmt.Println("All sums of two numbers that lead to 3:")
	solver.Solve([]Struct{st("add", Var("X"), Var("Y"), st("s", st("s", st("s", Atom("0")))))}, func(m map[Var]Term) bool {
		fmt.Println(m)
		return true
	})
	fmt.Println()
	fmt.Println("All sums of three numbers that lead to 5:")
	solver.Solve([]Struct{
		st("add", Var("X"), Var("Y"), Var("_Tmp")),
		st("add", Var("_Tmp"), Var("Z"), st("s", st("s", st("s", st("s", st("s", Atom("0"))))))),
	}, func(m map[Var]Term) bool {
		fmt.Println(m)
		return true
	})
}
