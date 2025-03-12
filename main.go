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
	prol.DCG{
		st("clause", st("clause", var_("Head"), var_("Body"))),
		st("struct", var_("Head")),
		st("ws"),
		prol.StringToTerm(":-").(prol.Struct),
		st("ws"),
		st("structs", var_("Body")),
		st("ws"),
		st(".", atom("."), atom("[]")),
	},
	prol.DCG{
		st("clause", st("clause", var_("Head"), atom("[]"))),
		st("struct", var_("Head")),
		st("ws"),
		st(".", atom("."), atom("[]")),
	},
	prol.DCG{
		st("structs", st(".", var_("Struct"), var_("Structs"))),
		st("struct", var_("Struct")),
		st("ws"),
		st(".", atom(","), atom("[]")),
		st("ws"),
		st("structs", var_("Structs")),
	},
	prol.DCG{
		st("structs", st(".", var_("Struct"), atom("[]"))),
		st("struct", var_("Struct")),
	},
	prol.DCG{st("term", var_("Term")), st("struct", var_("Term"))},
	prol.DCG{st("term", var_("Term")), st("atom", var_("Term"))},
	prol.DCG{st("term", var_("Term")), st("var", var_("Term"))},
	prol.DCG{
		st("struct", st("struct", var_("Name"), var_("Args"))),
		st("atom", st("atom", var_("Name"))),
		st(".", atom("("), atom("[]")),
		st("ws"),
		st("terms", var_("Args")),
		st("ws"),
		st(".", atom(")"), atom("[]")),
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
		kb.Assert(prol.Clause{st("upper", atom(string(ch)))})
	}
	for ch := 'a'; ch <= 'z'; ch++ {
		kb.Assert(prol.Clause{st("lower", atom(string(ch)))})
	}
	for ch := '0'; ch <= '9'; ch++ {
		kb.Assert(prol.Clause{st("digit", atom(string(ch)))})
	}
}

func runQuery(title string, query []prol.Struct, opts ...any) {
	limit := -1
	var solveOpts []any
	for i := 0; i < len(opts); {
		switch opts[i] {
		case "limit":
			limit = opts[i+1].(int)
			i += 2
		default:
			solveOpts = append(solveOpts, opts[i])
			i += 1
		}
	}
	fmt.Println("%", title)
	cnt := 1
	q := prol.Clause(append([]prol.Struct{st("query")}, query...))
	fmt.Println(q)
	for solution := range kb.Solve(q, solveOpts...) {
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
	runQuery("First 5 natural numbers",
		[]prol.Struct{st("nat", var_("X"))},
		"limit", 5)
	runQuery("All combinations of two numbers that sum to 3",
		[]prol.Struct{st("add", var_("X"), var_("Y"), st("s", st("s", st("s", atom("0")))))})
	runQuery("All combinations of three numbers that sum to 5", []prol.Struct{
		st("add", var_("_Tmp"), var_("Z"), st("s", st("s", st("s", st("s", st("s", atom("0"))))))),
		st("add", var_("X"), var_("Y"), var_("_Tmp")),
	})
	runQuery("First 5 lists with 'a'",
		[]prol.Struct{st("member", atom("a"), var_("List"))},
		"limit", 5)
	runQuery("Parsing atom", []prol.Struct{
		st("atom", var_("Atom1"), prol.StringToTerm("a"), atom("[]")),
		st("atom", var_("Atom2"), prol.StringToTerm("ab"), atom("[]")),
		st("atom", var_("Atom3"), prol.StringToTerm("a1"), atom("[]")),
		st("atom", var_("Atom4"), prol.StringToTerm("aX"), atom("[]")),
		st("atom", var_("Atom5"), prol.StringToTerm("a_"), atom("[]")),
		st("atom", var_("Atom6"), prol.StringToTerm("foo_123"), atom("[]")),
	})
	runQuery("Parsing var", []prol.Struct{
		st("var", var_("Var1"), prol.StringToTerm("X"), atom("[]")),
		st("var", var_("Var2"), prol.StringToTerm("Y"), atom("[]")),
		st("var", var_("Var3"), prol.StringToTerm("Xa"), atom("[]")),
		st("var", var_("Var4"), prol.StringToTerm("X1"), atom("[]")),
		st("var", var_("Var5"), prol.StringToTerm("Foo123"), atom("[]")),
	})
	runQuery("Parsing struct", []prol.Struct{
		st("struct", var_("Struct1"), prol.StringToTerm("f()"), atom("[]")),
		st("struct", var_("Struct2"), prol.StringToTerm("f( )"), atom("[]")),
		st("struct", var_("Struct3"), prol.StringToTerm("f(a)"), atom("[]")),
		st("struct", var_("Struct4"), prol.StringToTerm("f(Z)"), atom("[]")),
		st("struct", var_("Struct5"), prol.StringToTerm("f(g())"), atom("[]")),
		st("struct", var_("Struct6"), prol.StringToTerm("f(b,c)"), atom("[]")),
		st("struct", var_("Struct7"), prol.StringToTerm("f( b , c )"), atom("[]")),
		st("struct", var_("Struct8"), prol.StringToTerm("foo(atom, Var, bar(Y,c,d))"), atom("[]")),
	})
	runQuery("Parsing clause", []prol.Struct{
		st("clause", var_("Clause1"), prol.StringToTerm("f()."), atom("[]")),
		st("clause", var_("Clause2"), prol.StringToTerm("f(a)."), atom("[]")),
		st("clause", var_("Clause3"), prol.StringToTerm("f(a):-g()."), atom("[]")),
		st("clause", var_("Clause4"), prol.StringToTerm("f(a):-g(),h()."), atom("[]")),
		st("clause", var_("Clause4"), prol.StringToTerm("f(a) :-\n  g(),\n  h()."), atom("[]")),
	})
	runQuery("Assert clause", []prol.Struct{
		st("assertz", st("clause", st("struct", atom("dummy"), atom("[]")), atom("[]"))),
		st("dummy"),
		st("assertz", st("clause", st("struct", atom("f"), st(".", st("atom", atom("a123")), atom("[]"))), atom("[]"))),
		st("assertz", st("clause", st("struct", atom("f"), st(".", st("var", atom("X123")), atom("[]"))), atom("[]"))),
		st("assertz", st("clause", st("struct", atom("f"), st(".", st("struct", atom("g"), atom("[]")), atom("[]"))), atom("[]"))),
		st("f", var_("X")),
		st("assertz", st("clause", st("struct", atom("foo"), st(".", st("var", atom("X")), atom("[]"))), st(".", st("struct", atom("f"), st(".", st("var", atom("X")), atom("[]"))), atom("[]")))),
		st("f", var_("Y")),
	})
}
