package prol

import (
	"fmt"
	"iter"
	"log"
	"slices"
	"sort"
	"strings"
)

// --- Database ---

type Database struct {
	functors []Functor
	index0   map[Functor][]Rule
}

func NewDatabase(rules ...Rule) *Database {
	db := &Database{
		index0: make(map[Functor][]Rule),
	}
	for _, rule := range builtins {
		db.Assert(rule)
	}
	for _, rule := range rules {
		db.Assert(rule)
	}
	return db
}

func (db *Database) Assert(rule Rule) {
	f := rule.Functor()
	if _, ok := db.index0[f]; !ok {
		db.functors = append(db.functors, f)
	}
	db.index0[f] = append(db.index0[f], rule)
}

func (db *Database) PredicateExists(goal Struct) bool {
	_, ok := db.index0[goal.Functor()]
	return ok
}

func (db *Database) Matching(goal Struct) iter.Seq[Rule] {
	return func(yield func(Rule) bool) {
		f := goal.Functor()
		for _, rule := range db.index0[f] {
			if !yield(rule) {
				break
			}
		}
	}
}

func (db *Database) Solve(query Clause, opts ...any) iter.Seq[Solution] {
	env := make(map[Var]*Ref)
	query = varToRef(query, env).(Clause)
	s := &solver{db: db, env: env}
	s.readOpts(opts)
	return func(yield func(Solution) bool) {
		s.yield = yield
		s.dfs(query)
	}
}

func (db *Database) FirstSolution(query Clause, opts ...any) (Solution, error) {
	next, stop := iter.Pull(db.Solve(query, opts...))
	defer stop()
	solution, ok := next()
	if !ok {
		return nil, fmt.Errorf("expecting at least one solution: %v", query)
	}
	return solution, nil
}

func (db *Database) Interpret(text string, opts ...any) error {
	chars := FromString(text)
	rest := MustVar("Rest")
	for {
		_rest0, rule := MustVar("_Rest0"), MustVar("Rule")
		query := Clause{
			Struct{"query", nil},
			Struct{"ws_", []Term{chars, _rest0}},
			Struct{"rule_", []Term{rule, _rest0, rest}},
			Struct{"assertz", []Term{rule}},
		}
		solution, err := db.FirstSolution(query, opts...)
		if err != nil {
			break
		}
		chars = solution[rest]
	}
	fmt.Println("--- finished asserts ---")
	solution, err := db.FirstSolution(Clause{
		Struct{"query", nil},
		Struct{"ws_", []Term{chars, rest}}}, opts...)
	trailing, err := ToString(solution[rest])
	if err != nil {
		return err
	}
	if trailing != "" {
		return fmt.Errorf("trailing characters: %q", trailing)
	}
	return nil
}

// --- Solver ---

type Solver interface {
	Assert(rule Rule)
	Unify(t1, t2 Term) bool
	Unwind() func() bool
}

type Solution map[Var]Term

type solver struct {
	db    *Database
	env   map[Var]*Ref
	trail []*Ref
	yield func(Solution) bool
	// Opts
	trace        bool
	depth        int
	maxDepth     int
	numSolutions int
	limit        int
}

func (s *solver) readOpts(opts []any) {
	for i := 0; i < len(opts); {
		switch opts[i] {
		case "trace":
			s.trace = true
			i += 1
		case "max_depth":
			s.maxDepth = opts[i+1].(int)
			i += 2
		case "limit":
			s.limit = opts[i+1].(int)
			i += 2
		default:
			log.Printf("unknown option at %d: %v\n", i, opts[i])
			i += 1
		}
	}
}

func (s *solver) Assert(rule Rule) {
	s.db.Assert(rule)
}

func (s *solver) dfs(goals []Struct) bool {
	if s.limit > 0 && s.numSolutions >= s.limit {
		return false
	}
	if len(goals) == 0 {
		m := make(Solution)
		for x, ref := range s.env {
			if x[0] == '_' {
				continue
			}
			m[x] = RefToTerm(ref)
		}
		s.numSolutions++
		return s.yield(m)
	}
	goal, rest := goals[0], goals[1:]
	if s.trace {
		log.Println(">>> goal:", goal.Functor())
	}
	s.depth++
	defer func() { s.depth-- }()
	if s.maxDepth > 0 && s.depth > s.maxDepth {
		log.Println("max depth reached")
		return false
	}
	if !s.db.PredicateExists(goal) {
		log.Printf("predicate does not exist for goal: %v", goal.Functor())
		return false
	}
	unwind := s.Unwind()
	for rule := range s.db.Matching(goal) {
		body, ok := rule.Unify(s, goal)
		if ok {
			if !s.dfs(slices.Concat(body, rest)) {
				unwind()
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

// --- String ---

const (
	printDCGExpansion = false
)

func (db *Database) String() string {
	var b strings.Builder
	var cnt int
	for _, f := range db.functors {
		if len(db.index0[f]) == 1 {
			rule := db.index0[f][0]
			if _, ok := rule.(Builtin); ok {
				continue
			}
		}
		if cnt > 0 {
			b.WriteString("\n\n")
		}
		cnt++
		fmt.Fprintf(&b, "%% %v\n", f)
		for j, rule := range db.index0[f] {
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
