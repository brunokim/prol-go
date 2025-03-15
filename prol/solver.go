package prol

import (
	"fmt"
	"iter"
	"log"
	"slices"
	"sort"
	"strings"
)

type Solver interface {
	Assert(rule Rule)
	RetractIndex(f Functor, i int) bool
	MoveClauseInPredicate(f Functor, from int, to int) bool
	Unify(t1, t2 Term) bool
	Unwind() func() bool
}

type Solution map[Var]Term

type solver struct {
	kb    *KnowledgeBase
	env   map[Var]*Ref
	trail []*Ref
	yield func(Solution) bool
	// Opts
	trace     bool
	depth     int
	max_depth int
}

func (s *solver) Assert(rule Rule) {
	s.kb.Assert(rule)
}

func (s *solver) RetractIndex(f Functor, i int) bool {
	return s.kb.RetractIndex(f, i)
}

func (s *solver) MoveClauseInPredicate(f Functor, from int, to int) bool {
	return s.kb.MoveClauseInPredicate(f, from, to)
}

func (kb *KnowledgeBase) Solve(query Clause, opts ...any) iter.Seq[Solution] {
	env := make(map[Var]*Ref)
	query = varToRef(query, env).(Clause)
	s := &solver{kb: kb, env: env}
	for i := 0; i < len(opts); {
		switch opts[i] {
		case "trace":
			s.trace = true
			i += 1
		case "max_depth":
			s.max_depth = opts[i+1].(int)
			i += 2
		default:
			log.Printf("KnowledgeBase.Solve: unknown option at %d: %v\n", i, opts[i])
			i += 1
		}
	}
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
			m[x] = RefToTerm(ref)
		}
		return s.yield(m)
	}
	goal, rest := goals[0], goals[1:]
	// log.Println(">>> goal:", goal.Functor())
	if s.trace {
		log.Println(">>> goal:", goal.Functor())
	}
	s.depth++
	defer func() { s.depth-- }()
	if s.max_depth > 0 && s.depth > s.max_depth {
		log.Println("max depth reached")
		return false
	}
	if !s.kb.PredicateExists(goal) {
		log.Printf("predicate does not exist for goal: %v", goal.Functor())
		return false
	}
	unwind := s.Unwind()
	for rule := range s.kb.Matching(goal) {
		body, ok := rule.Unify(s, goal)
		if ok {
			if !s.dfs(slices.Concat(body, rest)) {
				return false
			}
		}
		unwind()
	}
	if s.trace {
		log.Println("<<< backtrack")
	}
	return true
}

func (s *solver) Unwind() func() bool {
	n := len(s.trail)
	return func() bool {
		if len(s.trail) == n {
			return false
		}
		for _, ref := range s.trail[n:] {
			ref.Value = nil
		}
		s.trail = s.trail[:n]
		return true
	}
}

func (s *solver) Unify(t1, t2 Term) bool {
	t1, t2 = Deref(t1), Deref(t2)
	if s.trace {
		log.Println("=== unify", t1, t2)
	}
	s1, isStruct1 := t1.(Struct)
	s2, isStruct2 := t2.(Struct)
	if isStruct1 && isStruct2 {
		if s1.Name != s2.Name || len(s1.Args) != len(s2.Args) {
			return false
		}
		for i := 0; i < len(s1.Args); i++ {
			if !s.Unify(s1.Args[i], s2.Args[i]) {
				return false
			}
		}
		return true
	}
	if t1 == t2 {
		return true
	}
	if ref1, ok := t1.(*Ref); ok {
		return s.bind(ref1, t2)
	}
	if ref2, ok := t2.(*Ref); ok {
		return s.bind(ref2, t1)
	}
	return false
}

func (s *solver) bind(ref *Ref, t Term) bool {
	if s.trace {
		log.Println("::: bind ", ref, t)
	}
	ref.Value = t
	s.trail = append(s.trail, ref)
	return true
}

// --- String ---

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
