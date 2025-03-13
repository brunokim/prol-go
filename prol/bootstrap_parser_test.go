package prol_test

import (
	_ "embed"
	"iter"
	"testing"

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
	rules, _ := prol.TermToList(solution["Rules"])
	for _, rule := range rules {
		t.Log(rule)
	}
	rest, _ := prol.TermToList(solution["Rest"])
	t.Log(rest[:min(len(rest), 50)])
}
