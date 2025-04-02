package prol

import (
	"fmt"
	"iter"
	"log"
	"maps"
	"slices"
	"sort"
	"strings"
)

// --- Database ---

type Database struct {
	indicators []Indicator
	index0     map[Indicator][]Rule
}

func NewDatabase(rules ...Rule) *Database {
	db := &Database{
		index0: make(map[Indicator][]Rule),
	}
	for _, rule := range builtins {
		db.Assert(rule)
	}
	for _, rule := range rules {
		db.Assert(rule)
	}
	return db
}

func (db *Database) Clone() *Database {
	return &Database{
		indicators: slices.Clone(db.indicators),
		index0:     maps.Clone(db.index0),
	}
}

func (db *Database) Assert(rule Rule) {
	f := rule.Indicator()
	if _, ok := db.index0[f]; !ok {
		db.indicators = append(db.indicators, f)
	}
	db.index0[f] = append(db.index0[f], rule)
}

func (db *Database) PredicateExists(goal Struct) bool {
	_, ok := db.index0[goal.Indicator()]
	return ok
}

func (db *Database) Matching(goal Struct) iter.Seq[Rule] {
	return func(yield func(Rule) bool) {
		f := goal.Indicator()
		for _, rule := range db.index0[f] {
			if !yield(rule) {
				break
			}
		}
	}
}

func (db *Database) Solve(query Clause, opts ...any) (iter.Seq[Solution], func() error) {
	env := make(map[Var]*Ref)
	query = varToRef(query, env).(Clause)
	s := &solver{db: db, env: env}
	s.readOpts(opts)
	var err error
	seq := func(yield func(Solution) bool) {
		s.yield = yield
		err = s.dfs(query)
	}
	errFn := func() error {
		return err
	}
	return seq, errFn
}

func (db *Database) FirstSolution(query Clause, opts ...any) (Solution, error) {
	seq, errFn := db.Solve(query, opts...)
	next, stop := iter.Pull(seq)
	defer stop()
	solution, ok := next()
	if !ok {
		return nil, fmt.Errorf("expecting at least one solution: %v", query)
	}
	return solution, errFn()
}

func (db *Database) Interpret(text string, opts ...any) error {
	chars := FromString(text)
	rest := MustVar("Rest")
	for {
		_rest0, rule := MustVar("_Rest0"), MustVar("Rule")
		query := Clause{
			Struct{"query", nil},
			Struct{"ws", []Term{chars, _rest0}},
			Struct{"parse_rule", []Term{rule, _rest0, rest}},
			Struct{"assertz", []Term{rule}},
		}
		solution, err := db.FirstSolution(query, opts...)
		if err != nil {
			break
		}
		chars = solution[rest]
	}
	log.Println("--- finished asserts ---")
	solution, err := db.FirstSolution(Clause{
		Struct{"query", nil},
		Struct{"ws", []Term{chars, rest}}}, opts...)
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
	GetPredicate(ind Indicator) []Rule
	PutPredicate(ind Indicator, rules []Rule) bool
	Assert(rule Rule)
	Unify(t1, t2 Term) bool
	Unwind() func() bool
}

type Solution map[Var]Term

type MaxSolutionsError struct{}
type MaxDepthError struct{}
type StopIterationError struct{}

func (MaxSolutionsError) Error() string  { return "max solutions reached" }
func (MaxDepthError) Error() string      { return "max depth reached" }
func (StopIterationError) Error() string { return "stop iteration" }

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

func (s *solver) GetPredicate(ind Indicator) []Rule {
	return s.db.index0[ind]
}

func (s *solver) PutPredicate(ind Indicator, rules []Rule) bool {
	for i, rule := range rules {
		if rule.Indicator() != ind {
			log.Printf("put_predicate: arg #%d: want %v, got %v", i+1, ind, rule.Indicator())
			return false
		}
	}
	if _, ok := s.db.index0[ind]; !ok {
		s.db.indicators = append(s.db.indicators, ind)
	}
	s.db.index0[ind] = rules
	return true
}

func (s *solver) Assert(rule Rule) {
	s.db.Assert(rule)
}

// --- Search ---

func (s *solver) solution() Solution {
	m := make(Solution)
	for x, ref := range s.env {
		if x[0] == '_' {
			continue
		}
		m[x] = RefToTerm(ref)
	}
	return m
}

func (s *solver) log(args ...any) {
	if s.trace {
		indent := strings.Repeat("  ", s.depth)
		args = append([]any{indent}, args...)
		log.Println(args...)
	}
}

func (s *solver) dfs(goals []Struct) error {
	if len(goals) == 0 {
		// Found a solution
		if !s.yield(s.solution()) {
			return StopIterationError{}
		}
		s.numSolutions++
		if s.limit > 0 && s.numSolutions >= s.limit {
			return MaxSolutionsError{}
		}
		return nil
	}
	goal, rest := goals[0], goals[1:]
	s.log(">>> goal:", goal.Indicator())
	// Check call depth.
	s.depth++
	defer func() { s.depth-- }()
	if s.maxDepth > 0 && s.depth > s.maxDepth {
		return MaxDepthError{}
	}
	// Check if predicate exists.
	if !s.db.PredicateExists(goal) {
		return fmt.Errorf("predicate does not exist for goal: %v", goal.Indicator())
	}
	unwind := s.Unwind()
	defer unwind()
	for rule := range s.db.Matching(goal) {
		unwind()
		body, ok, err := rule.Unify(s, goal)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if err := s.dfs(slices.Concat(body, rest)); err != nil {
			return err
		}
	}
	s.log("<<< backtrack")
	return nil
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
	for _, ind := range db.indicators {
		if len(db.index0[ind]) == 1 {
			rule := db.index0[ind][0]
			if _, ok := rule.(Builtin); ok {
				continue
			}
		}
		if cnt > 0 {
			b.WriteString("\n\n")
		}
		cnt++
		fmt.Fprintf(&b, "%% %v\n", ind)
		for j, rule := range db.index0[ind] {
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
