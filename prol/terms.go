package prol

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// --- Terms ---

// Term represents a logic term in Prolog.
type Term interface {
	isTerm()
	fmt.Stringer
}

// Atom is an immutable symbol.
type Atom string

// Var is a static-time variable.
type Var string

// Struct is a compound of multiple terms with a name.
type Struct struct {
	Name Atom
	Args []Term
}

// Ref is a run-time variable.
type Ref struct {
	name  Var
	id    int
	Value Term
}

func (Atom) isTerm()   {}
func (Var) isTerm()    {}
func (Struct) isTerm() {}
func (*Ref) isTerm()   {}

// --- Constructors ---

// NewVar creates a new variable, checking that the provided name is valid.
func NewVar(name string) (Var, error) {
	r, size := utf8.DecodeRuneInString(name)
	if r == utf8.RuneError && size == 0 {
		return Var(""), fmt.Errorf("empty var name")
	}
	if r == utf8.RuneError && size == 1 {
		return Var(""), fmt.Errorf("invalid encoding")
	}
	if !(unicode.IsUpper(r) || r == '_') {
		return Var(""), fmt.Errorf("first char in var is not uppercase or underscore: %c", r)
	}
	return Var(name), nil
}

// MustVar is like NewVar, but panics if the name is invalid.
func MustVar(name string) Var {
	v, err := NewVar(name)
	if err != nil {
		panic(err.Error())
	}
	return v
}

// NewStruct creates a struct from the given parameters.
func NewStruct(name string, terms ...Term) Struct {
	return Struct{Atom(name), terms}
}

var (
	refID = 0
)

// NewRef creates a fresh reference from the provided var.
func NewRef(v Var) *Ref {
	refID++
	return &Ref{v, refID, nil}
}

// --- Functor ---

// Functor represents the basic shape of a struct, with its name and arity.
type Functor struct {
	Name  Atom
	Arity int
}

func (f Functor) String() string {
	return fmt.Sprintf("%v/%d", f.Name, f.Arity)
}

func (s Struct) Functor() Functor {
	return Functor{s.Name, len(s.Args)}
}

// --- Atom ---

// IsChar returns whether this atom has a single char.
func (a Atom) IsChar() bool {
	return utf8.RuneCountInString(string(a)) == 1
}

// --- Conversion between term and list ---

// TermToList unwraps a linked list of cons cells into a list of terms.
func TermToList(t Term) (terms []Term, tail Term) {
	s, ok := t.(Struct)
	for ok && s.Name == "." && len(s.Args) == 2 {
		terms = append(terms, s.Args[0])
		t = Deref(s.Args[1])
		s, ok = t.(Struct)
	}
	tail = t
	return
}

// ListToTerm wraps the given list of terms into a linked list.
func ListToTerm(terms []Term, tail Term) Term {
	for i := len(terms) - 1; i >= 0; i-- {
		t := terms[i]
		tail = Struct{".", []Term{t, tail}}
	}
	return tail
}

// StringToTerm converts a string to a linked list of single-char atoms.
func StringToTerm(s string) Term {
	runes := []rune(s)
	terms := make([]Term, len(runes))
	for i, r := range runes {
		terms[i] = Atom(string(r))
	}
	return ListToTerm(terms, Atom("[]"))
}

// --- Ref ---

// Deref walks the chain of references until finding a non-ref term, or unbound ref.
func Deref(t Term) Term {
	if ref, ok := t.(*Ref); ok && ref.Value != nil {
		t = ref.Value
	}
	return t
}

// RefToTerm resolves all nested refs into ground terms, if possible.
func RefToTerm(x Term) Term {
	x = Deref(x)
	if s, ok := x.(Struct); ok {
		args := make([]Term, len(s.Args))
		for i, arg := range s.Args {
			args[i] = RefToTerm(arg)
		}
		return Struct{s.Name, args}
	}
	return x
}

// --- String ---

var (
	atomRE = regexp.MustCompile(`^([\p{Ll}\pN][\pL\pN_]*|\[\])$`)
)

func (t Atom) String() string {
	if atomRE.MatchString(string(t)) {
		return string(t)
	}
	return fmt.Sprintf("'%s'", strings.Replace(string(t), "'", "''", -1))
}

func (t Var) String() string {
	return string(t)
}

func (t Struct) String() string {
	terms, tail := TermToList(t)
	if len(terms) > 0 {
		return listToString(terms, tail)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%v(", t.Name)
	commaSeparated(&b, t.Args)
	b.WriteRune(')')
	return b.String()
}

func (t *Ref) String() string {
	return fmt.Sprintf("%s@%d", t.name, t.id)
}

func commaSeparated(b *strings.Builder, terms []Term) {
	for i, term := range terms {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(Deref(term).String())
	}
}

func listToString(terms []Term, tail Term) string {
	var b strings.Builder
	// Not a proper list.
	if tail != Atom("[]") {
		b.WriteRune('[')
		commaSeparated(&b, terms)
		b.WriteRune('|')
		b.WriteString(tail.String())
		b.WriteRune(']')
		return b.String()
	}
	// Deref all terms and check if it is a char list.
	isCharList := true
	for i := range terms {
		terms[i] = Deref(terms[i])
		if atom, ok := terms[i].(Atom); !(ok && atom.IsChar()) {
			isCharList = false
		}
	}
	// t is an ordinary list.
	if !isCharList {
		b.WriteRune('[')
		commaSeparated(&b, terms)
		b.WriteRune(']')
		return b.String()
	}
	// t is a proper char list.
	b.WriteRune('"')
	for _, term := range terms {
		atom := string(term.(Atom))
		if atom[0] == '"' {
			b.WriteString(`""`)
		} else {
			b.WriteString(atom)
		}
	}
	b.WriteRune('"')
	return b.String()
}

