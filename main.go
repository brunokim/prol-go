package main

import (
	"fmt"

	"github.com/brunokim/prol-go/prol"
)

// --- Test ---

func atom(name string) prol.Atom {
	return prol.Atom(name)
}

func var_(name string) prol.Var {
	return prol.MustVar(name)
}

func st(name string, terms ...prol.Term) prol.Struct {
	return prol.Struct{prol.Atom(name), terms}
}

var kb = prol.NewKnowledgeBase(
	// Basic clauses
	prol.Clause{st("query")},
	prol.Clause{st("nat", atom("0"))},
	prol.Clause{
		st("nat", st("s", var_("X"))),
		st("nat", var_("X")),
	},
	prol.Clause{st("add", atom("0"), var_("X"), var_("X"))},
	prol.Clause{
		st("add", st("s", var_("A")), var_("B"), st("s", var_("C"))),
		st("add", var_("A"), var_("B"), var_("C")),
	},
	prol.Clause{
		st("phrase", var_("Goal"), var_("List")),
		st("phrase", var_("Goal"), var_("List"), atom("[]")),
	},
	prol.Clause{
		st("phrase", var_("Goal"), var_("List"), var_("Rest")),
		st("call", var_("Goal"), var_("List"), var_("Rest")),
	},
	prol.Clause{st("=", var_("X"), var_("X"))},
	prol.Clause{
		st("member", var_("Elem"), st(".", var_("H"), var_("T"))),
		st("member_", var_("T"), var_("Elem"), var_("H")),
	},
	prol.Clause{st("member_", var_("_"), var_("Elem"), var_("Elem"))},
	prol.Clause{
		st("member_", st(".", var_("H"), var_("T")), var_("Elem"), var_("_")),
		st("member_", var_("T"), var_("Elem"), var_("H")),
	},
	// atom_chars
	prol.Clause{
		st("atom_chars", var_("Atom"), var_("Chars")),
		st("var", var_("Atom")),
		st("is_char_list", var_("Chars")),
		st("chars->atom", var_("Chars"), var_("Atom")),
	},
	prol.Clause{
		st("atom_chars", var_("Atom"), var_("Chars")),
		st("atom", var_("Atom")),
		st("atom->chars", var_("Atom"), var_("Chars")),
	},
	prol.Clause{st("is_char_list", atom("[]"))},
	prol.Clause{
		st("is_char_list", st(".", var_("Char"), var_("Chars"))),
		st("atom_length", var_("Char"), atom("1")),
		st("is_char_list", var_("Chars")),
	},
	// Prolog parser.
	prol.DCG{st("term", var_("Term")), st("struct", var_("Term"))},
	prol.DCG{st("term", var_("Term")), st("atom", var_("Term"))},
	prol.DCG{st("term", var_("Term")), st("var", var_("Term"))},
	prol.DCG{
		st("struct", st("struct", var_("Name"), var_("Args"))),
		st("atom", st("atom", var_("Name"))),
		st("ws"),
		st(".", atom("'('"), atom("[]")),
		st("ws"),
		st("terms", var_("Args")),
		st("ws"),
		st(".", atom("')'"), atom("[]")),
	},
	prol.DCG{
		st("struct", st("struct", var_("Name"), atom("[]"))),
		st("atom", st("atom", var_("Name"))),
		st("ws"),
		st(".", atom("("), atom("[]")),
		st("ws"),
		st(".", atom(")"), atom("[]")),
	},
	prol.DCG{
		st("terms", st(".", var_("Term"), var_("Terms"))),
		st("term", var_("Term")),
		st("ws"),
		st(".", atom(","), atom("[]")),
		st("ws"),
		st("terms", var_("Terms")),
	},
	prol.DCG{
		st("terms", st(".", var_("Term"), atom("[]"))),
		st("term", var_("Term")),
	},
	prol.DCG{
		st("atom", st("atom", var_("Atom"))),
		st(".", var_("Char"), atom("[]")),
		st("{}", st("atom_start", var_("Char"))),
		st("ident_chars", var_("Chars")),
		st("{}", st("atom_chars", var_("Atom"), st(".", var_("Char"), var_("Chars")))),
	},
	prol.DCG{
		st("var", st("var", var_("Var"))),
		st(".", var_("Char"), atom("[]")),
		st("{}", st("var_start", var_("Char"))),
		st("ident_chars", var_("Chars")),
		st("{}", st("atom_chars", var_("Var"), st(".", var_("Char"), var_("Chars")))),
	},
	prol.DCG{
		st("ident_chars", st(".", var_("Char"), var_("Chars"))),
		st(".", var_("Char"), atom("[]")),
		st("{}", st("ident", var_("Char"))),
		st("ident_chars", var_("Chars")),
	},
	prol.DCG{st("ident_chars", atom("[]"))},
	prol.DCG{
		st("ws"),
		st(".", var_("Char"), atom("[]")),
		st("{}", st("space", var_("Char"))),
		st("ws"),
	},
	prol.DCG{st("ws")},
	prol.Clause{st("atom_start", var_("Char")), st("lower", var_("Char"))},
	prol.Clause{st("atom_start", var_("Char")), st("digit", var_("Char"))},
	prol.Clause{st("var_start", atom("_"))},
	prol.Clause{st("var_start", var_("Char")), st("upper", var_("Char"))},
	prol.Clause{st("ident", atom("_"))},
	prol.Clause{st("ident", var_("Char")), st("lower", var_("Char"))},
	prol.Clause{st("ident", var_("Char")), st("upper", var_("Char"))},
	prol.Clause{st("ident", var_("Char")), st("digit", var_("Char"))},
	prol.Clause{st("space", atom(" "))},
	prol.Clause{st("space", atom("\n"))},
	prol.Clause{st("space", atom("\t"))},
)

func init() {
	for ch := 'A'; ch <= 'Z'; ch++ {
		kb.Assert(prol.Clause{st("upper", atom(fmt.Sprintf("'%c'", ch)))})
	}
	for ch := 'a'; ch <= 'z'; ch++ {
		kb.Assert(prol.Clause{st("lower", atom(string(ch)))})
	}
	for ch := '0'; ch <= '9'; ch++ {
		kb.Assert(prol.Clause{st("digit", atom(string(ch)))})
	}
}

func runQuery(title string, limit int, query ...prol.Struct) {
	fmt.Println("%", title)
	cnt := 1
	q := prol.Clause(append([]prol.Struct{st("query")}, query...))
	fmt.Println(q)
	for solution := range kb.Solve(q) {
		fmt.Println(solution)
		if cnt >= limit && limit >= 0 {
			break
		}
		cnt++
	}
	fmt.Println()
}

func main() {
	fmt.Println(kb)
	fmt.Println()
	runQuery("First 5 natural numbers", 5,
		st("nat", var_("X")))
	runQuery("All combinations of two numbers that sum to 3", -1,
		st("add", var_("X"), var_("Y"), st("s", st("s", st("s", atom("0"))))))
	runQuery("All combinations of three numbers that sum to 5", -1,
		st("add", var_("_Tmp"), var_("Z"), st("s", st("s", st("s", st("s", st("s", atom("0"))))))),
		st("add", var_("X"), var_("Y"), var_("_Tmp")),
	)
	runQuery("First 5 lists with 'a'", 5,
		st("member", atom("a"), var_("List")))
	runQuery("Parsing term", -1,
		st("struct", var_("Struct"), prol.ListToTerm([]prol.Term{atom("f"), atom("("), atom("a"), atom(","), atom("X"), atom(")")}, atom("[]")), atom("[]")))
}
