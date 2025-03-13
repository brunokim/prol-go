package prol_test

import (
	_ "embed"
	"iter"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/brunokim/prol-go/prol"
)

//go:embed lib/bootstrap.pl
var bootstrap string

func TestBootstrapParsesItself(t *testing.T) {
	kb, err := prol.Bootstrap()
	if err != nil {
		t.Fatal(err)
	}
	query := prol.Clause{
		prol.Struct{"query", nil},
		prol.Struct{"atom_chars_", []prol.Term{
			prol.Atom(bootstrap), prol.MustVar("_Chars")}},
		prol.Struct{"database_", []prol.Term{
			prol.MustVar("Rules"), prol.MustVar("_Chars"), prol.MustVar("Rest")}},
	}
	next, stop := iter.Pull(kb.Solve(query, "max_depth", len(bootstrap)*10))
	defer stop()
	solution, ok := next()
	if !ok {
		t.Errorf("Expecting at least one solution, found none")
		return
	}
	rulesAST, _ := prol.TermToList(solution["Rules"])
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
	rest, _ := prol.TermToList(solution["Rest"])
	t.Log(rest[:min(len(rest), 50)])
	compiledKB := prol.NewKnowledgeBase(rules...)
	exporter := func(typ reflect.Type) bool {
		return (typ == reflect.TypeOf(prol.KnowledgeBase{}))
	}
	diff := cmp.Diff(kb, compiledKB,
		cmp.Exporter(exporter),
		cmpopts.IgnoreUnexported(prol.Builtin{}))
	if diff != "" {
		t.Errorf("difference between compilers (-want, +got):\n%s", diff)
	}
}
