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
	Unify(t1, t2 Term) bool
}

type Solution map[Var]Term

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
			m[x] = RefToTerm(ref)
		}
		return s.yield(m)
	}
	goal, rest := goals[0], goals[1:]
	if !s.kb.PredicateExists(goal) {
		log.Printf("predicate does not exist for goal: %v", goal.Functor())
		return false
	}
	n := len(s.trail)
	for rule := range s.kb.Matching(goal) {
		body, ok := rule.Unify(s, goal)
		if ok {
			if !s.dfs(slices.Concat(body, rest)) {
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

func (s *solver) Unify(t1, t2 Term) bool {
	t1, t2 = Deref(t1), Deref(t2)
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
