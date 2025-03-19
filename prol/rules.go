package prol

import (
	"fmt"
	"iter"
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

// --- Knowledge base ---

type KnowledgeBase struct {
	functors []Functor
	index0   map[Functor][]Rule
}

func NewKnowledgeBase(rules ...Rule) *KnowledgeBase {
	kb := &KnowledgeBase{
		index0: make(map[Functor][]Rule),
	}
	for _, rule := range builtins {
		kb.Assert(rule)
	}
	for _, rule := range rules {
		kb.Assert(rule)
	}
	return kb
}

func (kb *KnowledgeBase) Assert(rule Rule) {
	f := rule.Functor()
	if _, ok := kb.index0[f]; !ok {
		kb.functors = append(kb.functors, f)
	}
	kb.index0[f] = append(kb.index0[f], rule)
}

func (kb *KnowledgeBase) RetractIndex(f Functor, i int) bool {
	clauses, ok := kb.index0[f]
	if !ok || len(clauses) < i {
		return false
	}
	kb.index0[f] = append(clauses[:i-1], clauses[i:]...)
	fmt.Println(kb.index0[f])
	if len(kb.index0[f]) == 0 {
		delete(kb.index0, f)
	}
	return true
}

func (kb *KnowledgeBase) MoveClauseInPredicate(f Functor, from int, to int) bool {
	rules, ok := kb.index0[f]
	if !ok || from < 1 || to < 1 || len(rules) < from || len(rules) < to {
		return false
	}
	if from == to {
		return true
	}
	i, j := from-1, to-1
	rule := rules[i]
	if i < j {
		// [a b X c d e f] --> [a b c d X e f]
		// [a b] [c d] X [e f]
		newRules := make([]Rule, len(rules))
		copy(newRules[:i], rules[:i])
		copy(newRules[i:j], rules[i+1:j+1])
		newRules[j] = rule
		copy(newRules[j:], rules[j+1:])
		kb.index0[f] = newRules
	} else {
		// [a b c d X e f] --> [a b X c d e f]
		// [a b] X [c d] [e f]
		newRules := make([]Rule, len(rules))
		copy(newRules[:j], rules[:j])
		newRules[j] = rule
		copy(newRules[j+1:i+1], rules[j:i])
		copy(newRules[i+1:], rules[i:])
		kb.index0[f] = newRules
	}
	fmt.Println(kb.index0[f])
	return true
}

func (kb *KnowledgeBase) PredicateExists(goal Struct) bool {
	_, ok := kb.index0[goal.Functor()]
	return ok
}

func (kb *KnowledgeBase) Matching(goal Struct) iter.Seq[Rule] {
	return func(yield func(Rule) bool) {
		f := goal.Functor()
		for _, rule := range kb.index0[f] {
			if !yield(rule) {
				break
			}
		}
	}
}

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

const (
	printDCGExpansion = false
)

func (kb *KnowledgeBase) String() string {
	var b strings.Builder
	for i, f := range kb.functors {
		if i > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "%% %v\n", f)
		for j, rule := range kb.index0[f] {
			if j > 0 {
				b.WriteRune('\n')
			}
			fmt.Fprintf(&b, "%v", rule)
			if dcg, ok := rule.(DCG); ok && printDCGExpansion {
				fmt.Fprintf(&b, "\n/*")
				fmt.Fprintf(&b, "%v", dcg.toClause())
				fmt.Fprintf(&b, "*/")
			}
		}
	}
	return b.String()
}

// --- Replace vars with refs ---

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
