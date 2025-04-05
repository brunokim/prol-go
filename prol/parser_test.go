package prol_test

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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
	if err1 != nil || err2 != nil {
		t.Errorf("source error: %v, %v", err1, err2)
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
