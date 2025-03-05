package main

import (
	"fmt"
	"iter"
	"sort"
	"strings"
)

// --- Terms ---

type Term interface {
	isTerm()
	fmt.Stringer
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

func (f Functor) String() string {
	return fmt.Sprintf("%v/%d", f.Name, f.Arity)
}

func termToList(t Term) (terms []Term, tail Term) {
	s, ok := t.(Struct)
	for ok && s.Name == "." && len(s.Args) == 2 {
		terms = append(terms, s.Args[0])
		t = s.Args[1]
		s, ok = t.(Struct)
	}
	tail = t
	return
}

type Ref struct {
	Name  Var
	ID    int
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

func (t Atom) String() string {
	return string(t)
}

func (t Var) String() string {
	return string(t)
}

func (t Struct) String() string {
	terms, tail := termToList(t)
	if len(terms) > 0 {
		strs := make([]string, len(terms))
		for i, term := range terms {
			strs[i] = term.String()
		}
		if tail == Atom("[]") {
			return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
		return fmt.Sprintf("[%s|%v]", strings.Join(strs, ", "), tail)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%v(", t.Name)
	for i, arg := range t.Args {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", deref(arg))
	}
	b.WriteRune(')')
	return b.String()
}

func (t *Ref) String() string {
	return fmt.Sprintf("%s@%d", t.Name, t.ID)
}

// --- Clause ---

type Clause []Struct

func (c Clause) Head() Struct {
	return c[0]
}

func (c Clause) Body() []Struct {
	return c[1:]
}

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

// --- Solution and KB ---

type Solution map[Var]Term

func (x Solution) String() string {
	keyvals := make([]string, len(x))
	var i int
	for k, v := range x {
		keyvals[i] = fmt.Sprintf("%v: %v", k, v)
		i++
	}
	sort.Strings(keyvals)
	return fmt.Sprintf("{%s}", strings.Join(keyvals, ", "))
}

type KnowledgeBase struct {
	functors []Functor
	index0   map[Functor][]Clause
}

func NewKnowledgeBase(clauses ...Clause) *KnowledgeBase {
	kb := &KnowledgeBase{
		index0: make(map[Functor][]Clause),
	}
	for _, clause := range clauses {
		kb.Assert(clause)
	}
	return kb
}

func (kb *KnowledgeBase) Assert(clause Clause) {
	f := clause.Head().Functor()
	if _, ok := kb.index0[f]; !ok {
		kb.functors = append(kb.functors, f)
	}
	kb.index0[f] = append(kb.index0[f], clause)
}

func (kb *KnowledgeBase) Matching(goal Struct) iter.Seq[Clause] {
	return func(yield func(Clause) bool) {
		f := goal.Functor()
		for _, clause := range kb.index0[f] {
			if !yield(clause) {
				break
			}
		}
	}
}

func (kb *KnowledgeBase) String() string {
	var b strings.Builder
	for i, f := range kb.functors {
		if i > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "%% %v\n", f)
		for j, clause := range kb.index0[f] {
			if j > 0 {
				b.WriteRune('\n')
			}
			fmt.Fprintf(&b, "%v", clause)
		}
	}
	return b.String()
}

// --- Solver ---

type solver struct {
	kb    *KnowledgeBase
	env   map[Var]*Ref
	trail []*Ref
	yield func(Solution) bool
}

func (kb *KnowledgeBase) Solve(query Clause) iter.Seq[Solution] {
	env := make(map[Var]*Ref)
	query = varToRef(query, env).(Clause)
	s := solver{kb: kb, env: env}
	return func(yield func(Solution) bool) {
		s.yield = yield
		s.dfs(query)
	}
}

func (s *solver) dfs(goals []Struct) bool {
	if len(goals) == 0 {
		m := make(Solution)
		for x, ref := range s.env {
			if x[0] == '_' {
				continue
			}
			m[x] = refToTerm(ref)
		}
		return s.yield(m)
	}
	goal, rest := goals[0], goals[1:]
	n := len(s.trail)
	for clause := range s.kb.Matching(goal) {
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

func (s *solver) unify(t1, t2 Term) bool {
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

// --- Test ---

func st(name string, terms ...Term) Struct {
	return Struct{Atom(name), terms}
}

func main() {
	kb := NewKnowledgeBase(
		Clause{st("query")},
		Clause{st("nat", Atom("0"))},
		Clause{
			st("nat", st("s", Var("X"))),
			st("nat", Var("X")),
		},
		Clause{st("add", Atom("0"), Var("X"), Var("X"))},
		Clause{
			st("add", st("s", Var("A")), Var("B"), st("s", Var("C"))),
			st("add", Var("A"), Var("B"), Var("C")),
		},
		Clause{
			st("phrase", Var("Goal"), Var("List")),
			st("phrase", Var("Goal"), Var("List"), Atom("[]")),
		},
		Clause{
			st("phrase", Var("Goal"), Var("List"), Var("Rest")),
			st("call", Var("Goal"), Var("List"), Var("Rest")),
		},
	)
	fmt.Println(kb)
	fmt.Println()
	var query Clause
	fmt.Println("% First 5 natural numbers")
	cnt := 1
	query = Clause{st("query"), st("nat", Var("X"))}
	fmt.Println(query)
	for solution := range kb.Solve(query) {
		fmt.Println(solution)
		if cnt >= 5 {
			break
		}
		cnt++
	}
	fmt.Println()
	fmt.Println("% All combinations of two numbers that sum to 3")
	query = Clause{st("query"), st("add", Var("X"), Var("Y"), st("s", st("s", st("s", Atom("0")))))}
	fmt.Println(query)
	for solution := range kb.Solve(query) {
		fmt.Println(solution)
	}
	fmt.Println()
	fmt.Println("% All combinations of three numbers that sum to 5")
	query = Clause{
		st("query"),
		st("add", Var("_Tmp"), Var("Z"), st("s", st("s", st("s", st("s", st("s", Atom("0"))))))),
		st("add", Var("X"), Var("Y"), Var("_Tmp")),
	}
	fmt.Println(query)
	for solution := range kb.Solve(query) {
		fmt.Println(solution)
	}
}
