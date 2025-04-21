package prol_test

import (
	_ "embed"
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/brunokim/prol-go/kif"
	"github.com/brunokim/prol-go/prol"
)

//go:embed lib/bootstrap.pl
var bootstrap string

func TestBootstrapParsesItself(t *testing.T) {
	db := prol.Bootstrap()
	query := clause(s("query"),
		s("atom_chars", a(bootstrap), v("_Chars")),
		s("parse_database", v("Rules"), v("_Chars"), v("_Rest0")),
		s("ws", v("_Rest0"), v("Rest")),
	)
	solution, err := db.FirstSolution(query, "max_depth", len(bootstrap)*10)
	if err != nil {
		t.Fatal(err)
	}
	rulesAST, _ := toList(solution["Rules"])
	var rules []prol.Rule
	for i, ruleAST := range rulesAST {
		rule, err := prol.CompileRule(ruleAST)
		if err != nil {
			t.Errorf("compilation failure in rule %d: %v", i, err)
		} else {
			rules = append(rules, rule)
			t.Log(rule)
		}
	}
	rest, _ := toString(solution["Rest"])
	if len(rest) > 0 {
		t.Errorf("trailing characters: %v", rest[:min(len(rest), 50)])
		return
	}
	compiledKB := prol.NewDatabase(rules...)
	exporter := func(typ reflect.Type) bool {
		return (typ == reflect.TypeOf(prol.Database{}))
	}
	diff := cmp.Diff(db, compiledKB,
		cmp.Exporter(exporter),
		cmpopts.IgnoreUnexported(prol.Builtin{}),
		cmpopts.IgnoreFields(prol.Database{}, "index1"))
	if diff != "" {
		t.Errorf("difference between compilers (-want, +got):\n%s", diff)
	}
}

var (
	//go:embed lib/prelude/01_comments.pl
	commentsFile string
	//go:embed lib/prelude/02_lists_and_quotes.pl
	listsFile string
	//go:embed lib/prelude/03_dcg.pl
	dcgFile string
	//go:embed lib/prelude/04_expressions.pl
	expressionsFile string
)

func TestPreludeComments(t *testing.T) {
	db := prol.Bootstrap()
	err := db.Interpret(commentsFile, "max_depth", len(commentsFile)*10)
	if err != nil {
		t.Errorf("comments error: %v", err)
	}
	tests := []struct {
		name    string
		content string
	}{
		{"Line comment", "% content \n"},
		{"C comment", "/* content */"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := db.Clone()
			err := db.Interpret(test.content)
			if err != nil {
				t.Errorf("test interpret err: %v", err)
			}
		})
	}
}

func TestPreludeLists(t *testing.T) {
	db := prol.Bootstrap()
	err1 := db.Interpret(commentsFile)
	err2 := db.Interpret(listsFile)
	err := errors.Join(err1, err2)
	if err != nil {
		t.Errorf("source error: %v", err)
	}
	tests := []struct {
		name    string
		content string
		query   prol.Clause
		want    prol.Solution
	}{
		{
			"Lists",
			"test_list([ ], [/*comment*/], [1], [ 1 ], [1, 2], [1|2], [1, 2|X], [1|[2|[3|[]]]]).",
			clause(
				s("query"),
				s("test_list", v("T1"), v("T2"), v("T3"), v("T4"), v("T5"), v("T6"), v("T7"), v("T8"))),
			prol.Solution{
				v("T1"): a("[]"),
				v("T2"): a("[]"),
				v("T3"): fromList(int_(1)),
				v("T4"): fromList(int_(1)),
				v("T5"): fromList(int_(1), int_(2)),
				v("T6"): s(".", int_(1), int_(2)),
				v("T7"): s(".", int_(1), s(".", int_(2), ref("X"))),
				v("T8"): fromList(int_(1), int_(2), int_(3)),
			},
		},
		{
			"Quoted atom",
			"test_quoted_atom('a', ' ', '''', '%-()[]:/**/').",
			clause(
				s("query"),
				s("test_quoted_atom", v("T1"), v("T2"), v("T3"), v("T4"))),
			prol.Solution{
				v("T1"): a("a"),
				v("T2"): a(" "),
				v("T3"): a("'"),
				v("T4"): a("%-()[]:/**/"),
			},
		},
		{
			"Quoted string",
			`test_quoted_string("a", "", "double->""<-quote", "1 2 3").`,
			clause(
				s("query"),
				s("test_quoted_string", v("T1"), v("T2"), v("T3"), v("T4"))),
			prol.Solution{
				v("T1"): fromString("a"),
				v("T2"): fromString(""),
				v("T3"): fromString(`double->"<-quote`),
				v("T4"): fromString("1 2 3"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := cmp.Options{
				cmp.AllowUnexported(prol.Ref{}),
				cmpopts.IgnoreFields(prol.Ref{}, "id"),
			}
			db := db.Clone()
			err := db.Interpret(test.content)
			if err != nil {
				t.Errorf("test interpret err: %v", err)
			}
			got, err := db.FirstSolution(test.query)
			if err != nil {
				t.Fatalf("want solution, got: %v", err)
			}
			if diff := cmp.Diff(test.want, got, opts...); diff != "" {
				t.Errorf("(-want, +got):\n%s", diff)
			}
		})
	}
}

func TestPreludeDCG(t *testing.T) {
	db := prol.Bootstrap()
	err1 := db.Interpret(commentsFile)
	err2 := db.Interpret(listsFile)
	err3 := db.Interpret(dcgFile)
	err := errors.Join(err1, err2, err3)
	if err != nil {
		t.Errorf("source error: %v", err)
	}
	tests := []struct {
		name    string
		content string
		query   prol.Clause
		want    prol.Solution
	}{
		{
			"Empty DCG",
			"test_dcg([]) --> [].",
			clause(
				s("query"),
				s("test_dcg", v("T"), fromString(""), v("Rest"))),
			prol.Solution{
				v("T"):    a("[]"),
				v("Rest"): a("[]"),
			},
		},
		{
			"Atom goal",
			`one_atom --> "a". test_dcg(1) --> one_atom.`,
			clause(
				s("query"),
				s("test_dcg", v("T"), fromString("a"), v("Rest"))),
			prol.Solution{
				v("T"):    int_(1),
				v("Rest"): a("[]"),
			},
		},
		{
			"A var",
			`a_struct(p(0)) --> []. test_dcg(X) --> a_struct(X).`,
			clause(
				s("query"),
				s("test_dcg", s("p", v("T")), fromString(""), v("Rest"))),
			prol.Solution{
				v("T"):    int_(0),
				v("Rest"): a("[]"),
			},
		},
		{
			"Two vars",
			`foo(y) --> [y]. test_dcg(P, Q) --> [P], foo(Q).`,
			clause(
				s("query"),
				s("test_dcg", v("A"), v("B"), fromString("xy"), v("Rest"))),
			prol.Solution{
				v("A"):    a("x"),
				v("B"):    a("y"),
				v("Rest"): a("[]"),
			},
		},
		{
			"Multiple DCG goals",
			`test(x, z).
             foo(z, w).
             test_dcg(a(X), Y) --> [X], ":", { test(X, _Z), foo(_Z, Y) }.`,
			clause(
				s("query"),
				s("test_dcg", v("A"), v("B"), fromString("x:"), v("Rest"))),
			prol.Solution{
				v("A"):    s("a", a("x")),
				v("B"):    a("w"),
				v("Rest"): a("[]"),
			},
		},
		{
			"Directive",
			`:- put_predicate(
                indicator(test_foo, 1),
                [clause(struct(test_foo, [int(1)]), [])]).`,
			clause(s("query"),
				s("test_foo", v("X"))),
			prol.Solution{
				v("X"): int_(1),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := cmp.Options{
				cmp.AllowUnexported(prol.Ref{}),
				cmpopts.IgnoreFields(prol.Ref{}, "id"),
			}
			db := db.Clone()
			err := db.Interpret(test.content)
			if err != nil {
				t.Errorf("test interpret err: %v", err)
			}
			got, err := db.FirstSolution(test.query)
			if err != nil {
				t.Fatalf("want solution, got: %v", err)
			}
			if diff := cmp.Diff(test.want, got, opts...); diff != "" {
				t.Errorf("(-want, +got):\n%s", diff)
			}
		})
	}
}

func TestPreludeExpressions(t *testing.T) {
	db := prol.Bootstrap()
	err1 := db.Interpret(commentsFile)
	err2 := db.Interpret(listsFile)
	err3 := db.Interpret(dcgFile)
	err4 := db.Interpret(expressionsFile)
	err := errors.Join(err1, err2, err3, err4)
	if err != nil {
		t.Errorf("source error: %v", err)
	}
	tests := []struct {
		name    string
		content string
		query   prol.Clause
		want    prol.Solution
	}{
		{
			"Parse atomic expr",
			`test_parse_expr.`,
			clause(s("query"), s("test_parse_expr")),
			prol.Solution{},
		},
		{
			"Parse atomic exprs",
			`test_parse_expr(1, a, X, f(g, h), [c, d]).`,
			clause(s("query"), s("test_parse_expr", v("T1"), v("T2"), v("T3"), v("T4"), v("T5"))),
			prol.Solution{
				v("T1"): int_(1),
				v("T2"): a("a"),
				v("T3"): ref("T3"),
				v("T4"): s("f", a("g"), a("h")),
				v("T5"): fromList(a("c"), a("d")),
			},
		},
		{
			"Parse parens",
			`test_parse_expr((1), ( 1 ), f((g)), +(1,2)).`,
			clause(s("query"), s("test_parse_expr", v("T1"), v("T2"), v("T3"), v("T4"))),
			prol.Solution{
				v("T1"): int_(1),
				v("T2"): int_(1),
				v("T3"): s("f", a("g")),
				v("T4"): s("+", int_(1), int_(2)),
			},
		},
		{
			"Parse prefix",
			`f(+ 2)`,
			clause(s("query"), s("f", v("T1") /*, v("T2"), v("T3"), v("T4")*/)),
			prol.Solution{
				v("T1"): s("+", int_(2)),
				v("T2"): s("-", int_(1)),
				v("T3"): s("+", int_(2)),
				v("T4"): s("-", int_(1)),
			},
		},
		// test_parse_expr(- -1, + -1, + +2).
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := cmp.Options{
				cmp.AllowUnexported(prol.Ref{}),
				cmpopts.IgnoreFields(prol.Ref{}, "id"),
			}
			db := db.Clone()
			var err error
			db.Logger, err = kif.NewFileLogger("testoutput/" + test.name + ".log")
			if err != nil {
				t.Fatalf("error opening log: %v", err)
			}
			defer db.Logger.Close()
			db.Logger.DisableCaller = true
            db.Logger.LogLevel = kif.DEBUG
			err = db.Interpret(test.content)
			if err != nil {
				t.Errorf("test interpret err: %v", err)
			}
			got, err := db.FirstSolution(test.query)
			if err != nil {
				t.Fatalf("want solution, got: %v", err)
			}
			if diff := cmp.Diff(test.want, got, opts...); diff != "" {
				t.Errorf("(-want, +got):\n%s", diff)
			}
		})
	}
}
