package prol

import (
	_ "embed"
	"fmt"
	"regexp"
)

//go:embed lib/bootstrap.pl
var bootstrap string

func Bootstrap() (*KnowledgeBase, error) {
	p := parser{bootstrap, 0}
	rules := p.database()
	kb := NewKnowledgeBase(rules...)
	if !p.isAtEOF() {
		return kb, fmt.Errorf("trailing characters at position %d:\n----\n%s\n----", p.pos, p.text[p.pos:min(len(p.text), p.pos+50)])
	}
	return kb, nil
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

func (p *parser) database() []Rule {
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
	return clauses
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
	clause := Clause{head}
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
		clause = append(clause, term.(Struct))
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
	if !ok {
		// Var
		return p.var_()
	}
	if !p.match(`\(`) {
		// Plain atom
		return atom, true
	}
	// Struct
	p.ws()
	if p.match(`\)`) {
		// Empty struct
		return Struct{atom, nil}, true
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
	m = p.match2(`([a-z0-9][a-z0-9A-Z_]*|\[\])`)
	if m != nil {
		// Symbol
		return Atom(m[0]), true
	}
	return Atom(""), false
}

func (p *parser) var_() (Var, bool) {
	m := p.match2(`[A-Z_][a-z0-9A-Z_]*`)
	if m == nil {
		return Var(""), false
	}
	return Var(m[0]), true
}

func (p *parser) ws() {
	p.match(`[ \n]*`)
}
