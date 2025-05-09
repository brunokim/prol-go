package prol

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

//go:embed lib
var lib embed.FS

func readLib(name string) string {
	bs, err := lib.ReadFile(name)
	if err != nil {
		panic(err.Error())
	}
	return string(bs)
}

func dirFiles(dirName string) []string {
	entries, err := lib.ReadDir(dirName)
	if err != nil {
		panic(fmt.Sprintf("prelude library error! %v", err))
	}
	names := make([]string, len(entries))
	slices.SortFunc(entries, func(a, b fs.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})
	for i, entry := range entries {
		names[i] = dirName + "/" + entry.Name()
	}
	return names
}

func Prelude(opts ...any) *Database {
	db := Bootstrap()
	for _, name := range dirFiles("lib/prelude") {
		content := readLib(name)
		if err := db.Interpret(content, opts...); err != nil {
			panic(fmt.Sprintf("prelude library error! %s: %v", name, err))
		}
	}
	return db
}

// --- Bootstrap parser ---

func Bootstrap() *Database {
	p := parser{readLib("lib/bootstrap.pl"), 0}
	rules, err := p.database()
	if err != nil {
		panic(fmt.Sprintf("bootstrap library error!\n%v", err))
	}
	return NewDatabase(rules...)
}

func Bootstrap2(opts ...any) *Database {
	var rules []Rule
	for _, name := range dirFiles("lib/bootstrap") {
		content := readLib(name)
		p := parser{content, 0}
		newRules, err := p.database()
		if err != nil {
			panic(fmt.Sprintf("bootstrap2 library error!\n%v", err))
		}
		rules = append(rules, newRules...)
	}
	return NewDatabase(rules...)
}

type parser struct {
	text string
	pos  int
}

func (p *parser) isAtEOF() bool {
	return p.pos >= len(p.text)
}

func (p *parser) match2(pattern string) []string {
	re := regexp.MustCompile("^" + pattern)
	m := re.FindStringSubmatch(p.text[p.pos:])
	// fmt.Printf("%25s  %40q  %v\n", pattern, p.text[p.pos:min(len(p.text), p.pos+40)], m)
	if m != nil {
		p.pos += len(m[0])
	}
	return m
}

func (p *parser) match(pattern string) bool {
	m := p.match2(pattern)
	return (m != nil)
}

func (p *parser) database() ([]Rule, error) {
	var clauses []Rule
	p.ws()
	for !p.isAtEOF() {
		clause, ok := p.clause()
		if !ok {
			break
		}
		clauses = append(clauses, clause)
		p.ws()
	}
	if !p.isAtEOF() {
		return clauses, fmt.Errorf(
			"trailing characters at position %d:\n----\n%s\n----",
			p.pos, p.text[p.pos:min(len(p.text), p.pos+50)])
	}
	return clauses, nil
}

func (p *parser) clause() (Clause, bool) {
	term, ok := p.term()
	if !ok {
		return Clause{}, false
	}
	head, ok := term.(Struct)
	if !ok {
		return Clause{}, false
	}
	clause := Clause{Goal{Term: head}}
	p.ws()
	if p.match(`\.`) {
		return clause, true
	}
	if !p.match(`:-`) {
		return clause, false
	}
	p.ws()
	body, ok := p.terms()
	if !ok {
		return Clause{}, false
	}
	for _, term := range body {
		clause = append(clause, Goal{Term: term.(Struct)})
	}
	if p.match(`\.`) {
		return clause, true
	}
	return Clause{}, false
}

func (p *parser) terms() ([]Term, bool) {
	var terms []Term
	for !p.isAtEOF() {
		term, ok := p.term()
		if !ok {
			break
		}
		terms = append(terms, term)
		p.ws()
		if !p.match(`,`) {
			break
		}
		p.ws()
	}
	return terms, len(terms) > 0
}

func (p *parser) term() (Term, bool) {
	atom, ok := p.atom()
	if ok {
		if !p.match(`\(`) {
			// Plain atom
			return atom, true
		}
		return p.structArgs(atom)
	}
	if x, ok := p.var_(); ok {
		// Var
		return x, true
	}
	// Int
	return p.int_()
}

func (p *parser) structArgs(atom Atom) (Struct, bool) {
	// Struct
	p.ws()
	if p.match(`\)`) {
		// Empty struct
		return Struct{atom, []Term{}}, true
	}
	args, ok := p.terms()
	if !ok {
		return Struct{}, false
	}
	if !p.match(`\)`) {
		return Struct{}, false
	}
	return Struct{atom, args}, true
}

func (p *parser) atom() (Atom, bool) {
	m := p.match2(`\\(.|\n)`)
	if m != nil {
		// Single character
		return Atom(m[1]), true
	}
	m = p.match2(`([a-z][a-z0-9A-Z_]*|\[\])`)
	if m != nil {
		// Symbol
		return Atom(m[0]), true
	}
	return Atom(""), false
}

func (p *parser) int_() (Int, bool) {
	m := p.match2(`[0-9]+`)
	if m == nil {
		return Int(0), false
	}
	// Int
	i, err := strconv.Atoi(m[0])
	if err != nil {
		return Int(0), false
	}
	return Int(i), true
}

func (p *parser) var_() (Var, bool) {
	m := p.match2(`[A-Z_][a-z0-9A-Z_]*`)
	if m == nil {
		return Var(""), false
	}
	return Var(m[0]), true
}

func (p *parser) ws() {
	p.match(`([ \n]|%[^\n]*)*`)
}
