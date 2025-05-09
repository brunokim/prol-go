% We run queries like '?- arg1, arg2, arg3.' by running as a clause like
% 'query :- arg1, arg2, arg3.'. For this to run, there needs to exist a
% query/0 clause to match.

query().


parse_query(Chars0, Rule, Rest) :-
  initial_lex_state(State0),
  lex_tokens(Tokens0, State0, State, Chars0, Chars),
  parse_goals(Goals, Tokens0, Tokens),
  \=(Rule, clause(goal(query, [], State0), Goals, State0)),
  \=(Rest, result(State, Chars, Tokens)).


interpret(Chars0) :-
  initial_lex_state(State0),
  lex_tokens(Tokens, State0, State, Chars0, Chars),
  parse_database(Rules, Tokens0, Tokens),
  assert_all(Rules).


assert_all(\.(Rule, Rules)) :-
  assertz(Rule),
  assert_all(Rules).
assert_all([]).

