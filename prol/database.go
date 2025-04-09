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
	index1     map[Indicator][]*ruleIndex
}

// f(1). f(s(a, b)). f(X). f(Y). f(p). f(Z).
// ----------------- ----------- ----- -----
//    constant         variable  const  var

// ?- f(p) => [f(X), f(Z), f(p), f(Y)]
// ?- f(s(A, B)) => [f(s(a, b)), f(X), f(Y), f(Z)]
// ?- f(W) => [f(1), f(s(a, b)), f(X), f(Y), f(p), f(Z)]

type ruleIndex struct {
	isVar    bool
	byVar    []Rule
	byAtom   map[Atom][]Rule
	byInt    map[Int][]Rule
	byStruct map[Indicator][]Rule
}

type keyType int

const (
	atomKey keyType = iota
	intKey
	structKey
)

func newRuleIndex(isVar bool) *ruleIndex {
	if isVar {
		return &ruleIndex{isVar: true}
	}
	return &ruleIndex{
		isVar:    false,
		byAtom:   make(map[Atom][]Rule),
		byInt:    make(map[Int][]Rule),
		byStruct: make(map[Indicator][]Rule),
	}
}

func NewDatabase(rules ...Rule) *Database {
	db := &Database{
		index0: make(map[Indicator][]Rule),
		index1: make(map[Indicator][]*ruleIndex),
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
		index1:     maps.Clone(db.index1),
	}
}

func (db *Database) Assert(rule Rule) {
	f := rule.Indicator()
	if _, ok := db.index0[f]; !ok {
		db.indicators = append(db.indicators, f)
	}
	db.index0[f] = append(db.index0[f], rule)
	// Populate index1 from first arg type.
	var firstArg Term
	switch c := rule.(type) {
	case Clause:
		if f.Arity == 0 {
			return
		}
		firstArg = c[0].Args[0]
	case DCG:
		if f.Arity == 2 {
			return
		}
		firstArg = c.dcgGoals[0].Args[0]
	case Builtin:
		return
	default:
		panic(fmt.Sprintf("unhandled rule type %T", rule))
	}
	// Add new index to list if we're starting a different block.
	_, isVar := firstArg.(Var)
	indices, ok := db.index1[f]
	if !ok || indices[len(indices)-1].isVar != isVar {
		db.index1[f] = append(indices, newRuleIndex(isVar))
	}
	n := len(db.index1[f])
	lastIndex := db.index1[f][n-1]
	// Append rule to index.
	lastIndex.byVar = append(lastIndex.byVar, rule)
	switch t := firstArg.(type) {
	case Atom:
		lastIndex.byAtom[t] = append(lastIndex.byAtom[t], rule)
	case Int:
		lastIndex.byInt[t] = append(lastIndex.byInt[t], rule)
	case Struct:
		lastIndex.byStruct[t.Indicator()] = append(lastIndex.byStruct[t.Indicator()], rule)
	case Var:
		// Do nothing
	default:
		panic(fmt.Sprintf("unhandled term type %T", firstArg))
	}
}

func (db *Database) PredicateExists(goal Struct) bool {
	_, ok := db.index0[goal.Indicator()]
	return ok
}

func (db *Database) Matching(goal Struct) []Rule {
	f := goal.Indicator()
	indices, ok := db.index1[f]
	if !ok {
		return db.index0[f]
	}
	firstArg := Deref(goal.Args[0])
	if _, ok := firstArg.(*Ref); ok {
		return db.index0[f]
	}
	var rules []Rule
	for _, index := range indices {
		if index.isVar {
			rules = append(rules, index.byVar...)
			continue
		}
		switch t := firstArg.(type) {
		case Atom:
			rules = append(rules, index.byAtom[t]...)
		case Int:
			rules = append(rules, index.byInt[t]...)
		case Struct:
			rules = append(rules, index.byStruct[t.Indicator()]...)
		default:
			panic(fmt.Sprintf("unhandled term type %T", t))
		}
	}
	return rules
}

func (db *Database) Solve(query Clause, opts ...any) (iter.Seq[Solution], func() error) {
	env := make(map[Var]*Ref)
	query = varToRef(query, env).(Clause)
	s := &solver{db: db, env: env}
	s.readOpts(opts)
	var err error
	seq := func(yield func(Solution) bool) {
		s.yield = yield
		err = s.dfs(newEnvironment(query))
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
	for {
		query := Clause{
			Struct{"query", nil},
			Struct{"ws", []Term{chars, v("_Rest0")}},
			Struct{"parse_rule", []Term{v("Rule"), v("_Rest0"), v("Rest")}},
			Struct{"assertz", []Term{v("Rule")}},
		}
		solution, err := db.FirstSolution(query, opts...)
		if err != nil {
			break
		}
		chars = solution[v("Rest")]
	}
	log.Println("--- finished asserts ---")
	solution, err := db.FirstSolution(Clause{
		Struct{"query", nil},
		Struct{"ws", []Term{chars, v("Rest")}}}, opts...)
	if err != nil {
		return err
	}
	trailing, err := ToString(solution[v("Rest")])
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
	if _, ok := s.db.index0[ind]; ok {
		// Clear existing predicate.
		delete(s.db.index0, ind)
		if ind.Arity > 0 {
			delete(s.db.index1, ind)
		}
	}
	// Otherwise, assert all other rules.
	for _, rule := range rules {
		s.db.Assert(rule)
	}
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

// --- Environment

type environment struct {
	goals  []Struct
	parent *environment
}

func newEnvironment(goals []Struct) *environment {
	return &environment{goals: goals}
}

func (env *environment) isDone() bool {
	return env == nil
}

func (env *environment) next() (Struct, *environment) {
	goal, rest := env.goals[0], env.goals[1:]
	if len(rest) > 0 {
		return goal, &environment{goals: rest, parent: env.parent}
	}
	return goal, env.parent
}

func (env *environment) push(goals []Struct) *environment {
	if len(goals) == 0 {
		return env
	}
	return &environment{goals: goals, parent: env}
}

// ---

func (s *solver) dfs(env *environment) error {
	if env.isDone() {
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
	goal, env := env.next()
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
	for _, rule := range s.db.Matching(goal) {
		unwind()
		body, ok, err := rule.Unify(s, goal)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if err := s.dfs(env.push(body)); err != nil {
			return err
		}
	}
	//s.log("<<< backtrack")
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
	//s.log("=== unify", t1, t2)
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
	//s.log("::: bind ", ref, t)
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
		rules, ok := db.index0[ind]
		if !ok {
			// Indicator was deleted.
			continue
		}
		if len(rules) == 1 {
			// Don't print builtin rule.
			rule := rules[0]
			if _, ok := rule.(Builtin); ok {
				continue
			}
		}
		if cnt > 0 {
			b.WriteString("\n\n")
		}
		cnt++
		fmt.Fprintf(&b, "%% %v\n", ind)
		for j, rule := range rules {
			if j > 0 {
				b.WriteRune('\n')
			}
			fmt.Fprintf(&b, "%v", rule)
			if dcg, ok := rule.(DCG); ok && printDCGExpansion {
				fmt.Fprintf(&b, "\n/*")
				fmt.Fprintf(&b, "%v", dcg.clause)
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
