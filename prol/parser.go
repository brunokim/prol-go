package prol

import (
	_ "embed"
	"fmt"
	"iter"
)

//go:embed lib/prelude.pl
var prelude string

// 2025/03/15 02:26:25 clause(struct(quoted_atom_chars, [var('Chars'), var('L1'), var('L')]), [])
// 2025/03/15 02:26:25 clause(struct(quoted_atom_chars, [atom([]), var('L0'), var('L')]), [struct('=', [var('L0'), struct('.', [atom(''''), var('L')])])])
// 2025/03/15 02:26:25 clause(struct(quoted_atom_chars, [struct('.', [var('Char'), var('Chars')]), var('L0'), var('L')]), [struct('=', [var('L0'), struct('.', [var('Char'), var('L1')])])

func Prelude(opts ...any) (*KnowledgeBase, error) {
	kb, err := Bootstrap()
	if err != nil {
		// Should never happen, represents a library error.
		return nil, fmt.Errorf("bootstrap library error: %w", err)
	}
	if err := kb.Interpret(prelude, opts...); err != nil {
		// Should never happen, represents a library error.
		return nil, fmt.Errorf("prelude library error: %w", err)
	}
	return kb, nil
}

func (kb *KnowledgeBase) FirstSolution(query Clause, opts ...any) (Solution, error) {
	next, stop := iter.Pull(kb.Solve(query, opts...))
	defer stop()
	solution, ok := next()
	if !ok {
		return nil, fmt.Errorf("expecting at least one solution: %v", query)
	}
	return solution, nil
}

func (kb *KnowledgeBase) Interpret(text string, opts ...any) error {
	chars := StringToTerm(text)
	rest := MustVar("Rest")
	for {
		_rest0, rule := MustVar("_Rest0"), MustVar("Rule")
		query := Clause{
			Struct{"query", nil},
			Struct{"ws_", []Term{chars, _rest0}},
			Struct{"rule_", []Term{rule, _rest0, rest}},
			Struct{"assertz", []Term{rule}},
		}
		solution, err := kb.FirstSolution(query, opts...)
		if err != nil {
			break
		}
		chars = solution[rest]
	}
	fmt.Println("--- finished asserts ---")
	solution, err := kb.FirstSolution(Clause{
		Struct{"query", nil},
		Struct{"ws_", []Term{chars, rest}}}, opts...)
	trailing, err := TermToString(solution[rest])
	if err != nil {
		return err
	}
	if trailing != "" {
		return fmt.Errorf("trailing characters: %q", trailing)
	}
	return nil
}
