package main

import (
	"fmt"
	"iter"
	"log"
	"slices"
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

func listToTerm(terms []Term, tail Term) Term {
	for i := len(terms) - 1; i >= 0; i-- {
		t := terms[i]
		tail = Struct{".", []Term{t, tail}}
	}
	return tail
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

type Rule interface {
	isRule()
	Functor() Functor
	Unify(s Solver, goal Struct) (body []Struct, ok bool)
}

type Clause []Struct

func (Clause) isRule() {}

func (c Clause) Head() Struct {
	return c[0]
}

func (c Clause) Body() []Struct {
	return c[1:]
}

func (c Clause) Functor() Functor {
	return c.Head().Functor()
}

func (c Clause) Unify(s Solver, goal Struct) ([]Struct, bool) {
	c = varToRef(c, map[Var]*Ref{}).(Clause)
	if !s.Unify(c.Head(), goal) {
		return nil, false
	}
	return c.Body(), true
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

// --- DCG ---

type DCG []Struct

func (DCG) isRule() {}

func (c DCG) Functor() Functor {
	return Functor{c[0].Name, len(c[0].Args) + 2}
}

func (c DCG) Unify(s Solver, goal Struct) ([]Struct, bool) {
	return c.toClause().Unify(s, goal)
}

func (c DCG) toClause() Clause {
	var ss Clause
	ss = append(ss, Struct{c[0].Name, slices.Concat(c[0].Args, []Term{Var("L0"), Var("L")})})
	var i int
	for _, s := range c[1:] {
		// List
		terms, tail := termToList(s)
		if len(terms) > 0 {
			if tail != Atom("[]") {
				panic(fmt.Sprintf("invalid DCG"))
			}
			curr, next := Var(fmt.Sprintf("L%d", i)), Var(fmt.Sprintf("L%d", i+1))
			ss = append(ss, Struct{"=", []Term{curr, listToTerm(terms, next)}})
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

// --- Builtins

type Builtin struct {
	functor Functor
	unify   func(Solver, Struct) ([]Struct, bool)
}

func (Builtin) isRule() {}

func (c Builtin) Functor() Functor {
	return c.functor
}

func (c Builtin) Unify(s Solver, goal Struct) ([]Struct, bool) {
	return c.unify(s, goal)
}

func atomBuiltin(s Solver, goal Struct) ([]Struct, bool)        { return nil, false }
func varBuiltin(s Solver, goal Struct) ([]Struct, bool)         { return nil, false }
func atomToCharsBuiltin(s Solver, goal Struct) ([]Struct, bool) { return nil, false }
func charsToAtomBuiltin(s Solver, goal Struct) ([]Struct, bool) { return nil, false }
func atomLengthBuiltin(s Solver, goal Struct) ([]Struct, bool)  { return nil, false }

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
	index0   map[Functor][]Rule
}

func NewKnowledgeBase(rules ...Rule) *KnowledgeBase {
	kb := &KnowledgeBase{
		index0: make(map[Functor][]Rule),
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

// --- Solver ---

type Solver interface {
	Unify(t1, t2 Term) bool
}

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
	t1, t2 = deref(t1), deref(t2)
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

// --- Test ---

func st(name string, terms ...Term) Struct {
	return Struct{Atom(name), terms}
}

func main() {
	kb := NewKnowledgeBase(
		// Builtins
		Builtin{Functor{"atom", 1}, atomBuiltin},
		Builtin{Functor{"var", 1}, varBuiltin},
		Builtin{Functor{"atom->chars", 2}, atomToCharsBuiltin},
		Builtin{Functor{"chars->atom", 2}, charsToAtomBuiltin},
		Builtin{Functor{"atom_length", 2}, atomLengthBuiltin},
		// Basic clauses
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
		Clause{st("=", Var("X"), Var("X"))},
		Clause{
			st("member", Var("Elem"), st(".", Var("H"), Var("T"))),
			st("member_", Var("T"), Var("Elem"), Var("H")),
		},
		Clause{st("member_", Var("_"), Var("Elem"), Var("Elem"))},
		Clause{
			st("member_", st(".", Var("H"), Var("T")), Var("Elem"), Var("_")),
			st("member_", Var("T"), Var("Elem"), Var("H")),
		},
		// atom_chars
		Clause{
			st("atom_chars", Var("Atom"), Var("Chars")),
			st("var", Var("Atom")),
			st("is_char_list", Var("Chars")),
			st("chars->atom", Var("Chars"), Var("Atom")),
		},
		Clause{
			st("atom_chars", Var("Atom"), Var("Chars")),
			st("atom", Var("Atom")),
			st("atom->chars", Var("Atom"), Var("Chars")),
		},
		Clause{st("is_char_list", Atom("[]"))},
		Clause{
			st("is_char_list", st(".", Var("Char"), Var("Chars"))),
			st("atom_length", Var("Char"), Atom("1")),
			st("is_char_list", Var("Chars")),
		},
		// Prolog parser.
		DCG{st("term", Var("Term")), st("struct", Var("Term"))},
		DCG{st("term", Var("Term")), st("atom", Var("Term"))},
		DCG{st("term", Var("Term")), st("var", Var("Term"))},
		DCG{
			st("struct", st("struct", Var("Name"), Var("Args"))),
			st("atom", st("atom", Var("Name"))),
			st("ws"),
			st(".", Atom("'('"), Atom("[]")),
			st("ws"),
			st("terms", Var("Args")),
			st("ws"),
			st(".", Atom("')'"), Atom("[]")),
		},
		DCG{
			st("struct", st("struct", Var("Name"), Atom("[]"))),
			st("atom", st("atom", Var("Name"))),
			st("ws"),
			st(".", Atom("'('"), Atom("[]")),
			st("ws"),
			st(".", Atom("')'"), Atom("[]")),
		},
		DCG{
			st("terms", st(".", Var("Term"), Var("Terms"))),
			st("term", Var("Term")),
			st("ws"),
			st(".", Atom("','"), Atom("[]")),
			st("ws"),
			st("terms", Var("Terms")),
		},
		DCG{
			st("terms", st(".", Var("Term"), Atom("[]"))),
			st("term", Var("Term")),
		},
		DCG{
			st("atom", st("atom", Var("Atom"))),
			st(".", Var("Char"), Atom("[]")),
			st("{}", st("atom_start", Var("Char"))),
			st("ident_chars", Var("Chars")),
			st("{}", st("atom_chars", Var("Atom"), st(".", Var("Char"), Var("Chars")))),
		},
		DCG{
			st("var", st("var", Var("Var"))),
			st(".", Var("Char"), Atom("[]")),
			st("{}", st("var_start", Var("Char"))),
			st("ident_chars", Var("Chars")),
			st("{}", st("atom_chars", Var("Var"), st(".", Var("Char"), Var("Chars")))),
		},
		DCG{
			st("ident_chars", st(".", Var("Char"), Var("Chars"))),
			st(".", Var("Char"), Atom("[]")),
			st("{}", st("ident", Var("Char"))),
			st("ident_chars", Var("Chars")),
		},
		DCG{st("ident_chars", Atom("[]"))},
		DCG{
			st("ws"),
			st(".", Var("Char"), Atom("[]")),
			st("{}", st("space", Var("Char"))),
			st("ws"),
		},
		DCG{st("ws")},
		Clause{st("atom_start", Var("Char")), st("lower", Var("Char"))},
		Clause{st("atom_start", Var("Char")), st("digit", Var("Char"))},
		Clause{st("var_start", Atom("'_'"))},
		Clause{st("var_start", Var("Char")), st("upper", Var("Char"))},
		Clause{st("ident", Atom("'_'"))},
		Clause{st("ident", Var("Char")), st("lower", Var("Char"))},
		Clause{st("ident", Var("Char")), st("upper", Var("Char"))},
		Clause{st("ident", Var("Char")), st("digit", Var("Char"))},
		Clause{st("space", Atom("' '"))},
		Clause{st("space", Atom("'\n'"))},
		Clause{st("space", Atom("'\t'"))},
	)
	for ch := 'A'; ch <= 'Z'; ch++ {
		kb.Assert(Clause{st("upper", Atom(fmt.Sprintf("'%c'", ch)))})
	}
	for ch := 'a'; ch <= 'z'; ch++ {
		kb.Assert(Clause{st("lower", Atom(string(ch)))})
	}
	for ch := '0'; ch <= '9'; ch++ {
		kb.Assert(Clause{st("digit", Atom(string(ch)))})
	}
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
	fmt.Println("% First 5 lists with 'a'")
	cnt = 1
	query = Clause{st("query"), st("member", Atom("a"), Var("List"))}
	fmt.Println(query)
	for solution := range kb.Solve(query) {
		fmt.Println(solution)
		if cnt >= 5 {
			break
		}
		cnt++
	}
	fmt.Println()
	fmt.Println("% Parsing term")
	query = Clause{
		st("query"),
		st("struct", Var("Struct"), listToTerm([]Term{Atom("f"), Atom("'('"), Atom("a"), Atom("','"), Atom("'X'"), Atom("')'")}, Atom("[]")), Atom("[]")),
	}
	fmt.Println(query)
	for solution := range kb.Solve(query) {
		fmt.Println(solution)
	}
}
