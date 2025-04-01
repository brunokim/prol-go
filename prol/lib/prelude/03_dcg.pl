% parse_dcg//1 parses a DCG rule.

parse_dcg(dcg(Head, Body), L0, L) :-
  parse_goal(Head, L0, L1),
  ws(L1, L2),
  '='(L2, ['-', '-', '>'|L3]),
  ws(L3, L4),
  parse_dcg_goals(Body, L4, L5),
  ws(L5, L6),
  '='(L6, ['.'|L]).

parse_dcg_goals([Goal|Goals], L0, L) :-
  parse_dcg_goal(Goal, L0, L1),
  ws(L1, L2),
  '='(L2, [','|L3]),
  ws(L3, L4),
  parse_dcg_goals(Goals, L4, L).
parse_dcg_goals([Goal], L0, L) :-
  parse_dcg_goal(Goal, L0, L).

parse_dcg_goal(Goal, L0, L) :-
  parse_goal(Goal, L0, L).
parse_dcg_goal(List, L0, L) :-
  parse_list(List, L0, L).
parse_dcg_goal(struct('{}', Goals), L0, L) :-
  '='(L0, ['{'|L1]),
  ws(L1, L2),
  parse_goals(Goals, L2, L3),
  ws(L3, L4),
  '='(L4, ['}'|L]).

parse_rule(Rule, L0, L) :-
  parse_dcg(Rule, L0, L).


test_dcg([]) --> [].
test_dcg(1) --> an_atom.
test_dcg(X) --> a_struct(X).
test_dcg(P, Q) --> [P], test_dcg(Q).
test_dcg(a(X), Y) --> X, ":", { test(X, _Z), foo(_Z, Y) }.


% parse_directive//1 parses a directive, that is, a rule for immediate execution.

parse_directive(clause(struct(directive, []), Goals)) -->
    ":-",
    ws,
    parse_goals(Goals),
    ws,
    ".".

parse_rule(Rule) --> parse_directive(Rule).

:- print('directive parsed').
