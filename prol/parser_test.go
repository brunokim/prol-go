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
	db, err := prol.Bootstrap()
	if err != nil {
		t.Fatal(err)
	}
	query := clause(s("query"),
		s("atom_chars", a(bootstrap), v("_Chars")),
		s("database", v("Rules"), v("_Chars"), v("_Rest0")),
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
		cmpopts.IgnoreUnexported(prol.Builtin{}))
	if diff != "" {
		t.Errorf("difference between compilers (-want, +got):\n%s", diff)
	}
}
