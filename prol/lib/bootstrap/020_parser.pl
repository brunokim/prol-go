% parse_database//1 reads a sequence of rules from the token difference list.

parse_database(\.(Rule, Rules), T0, T) :-
  parse_rule(Rule, T0, T1),
  ws(T1, T2),
  parse_database(Rules, T2, T).
parse_database(\.(Rule, []), T0, T) :-
  parse_rule(Rule, T0, T).


% parse_rule//1 parses a rule. For now, we only have one rule type (clause), but we will use this
% predicate as a hook for other types later.

parse_rule(Rule, T0, T) :-
  parse_clause(Rule, T0, T).


% parse_clause//1 parses a fact or clause.

parse_clause(clause(Head, Body, S0), T0, T) :-
  lexer_state(T0, S0),
  parse_goal(Head, T0, T1),
  ws(T1, T2),
  implied_by(T2, T3),
  ws(T3, T4),
  parse_goals(Body, T4, T5),
  ws(T5, T6),
  full_stop(T6, T).
parse_clause(clause(Head, [], S0), T0, T) :-
  lexer_state(T0, S0),
  parse_goal(Head, T0, T1),
  ws(T1, T2),
  implied_by(T2, T).


% parse_goals//1 parses a sequence of one or more comma-separated goals.

parse_goals(\.(Goal, Goals), T0, T) :-
  parse_goal(Goal, T0, T1),
  ws(T1, T2),
  comma(T2, T3),
  ws(T3, T4),
  parse_goals(Goals, T4, T).
parse_goals(\.(Goal, []), T0, T) :-
  parse_goal(Goal, T0, T).


% parse_goal//1 parses a goal.

parse_goal(goal(Name, Args, S0), T0, T) :-
  lexer_state(T0, S0),
  parse_struct(struct(Name, Args), T0, T).
parse_goal(goal(Atom, [], S0), T0, T) :-
  lexer_state(T0, S0),
  parse_atom(atom(Atom), T0, T).
parse_goal(goal(call, \.(X, []), S0), T0, T) :-
  lexer_state(T0, S0),
  parse_var(X, T0, T).


% parse_terms//1 parses a sequence of one or more comma-separated terms.

parse_terms(\.(Term, Terms), T0, T) :-
  parse_term(Term, T0, T1),
  ws(T1, T2),
  comma(T2, T3),
  ws(T3, T4),
  parse_terms(Terms, T4, T).
parse_terms(\.(Term, []), T0, T) :-
  parse_term(Term, T0, T).


% parse_term//1 parses a term.

parse_term(Term, T0, T) :-
  parse_struct(Term, T0, T).
parse_term(Term, T0, T) :-
  parse_atom(Term, T0, T).
parse_term(Term, T0, T) :-
  parse_var(Term, T0, T).
parse_term(Term, T0, T) :-
  parse_int(Term, T0, T).


% parse_struct//1 parses a structure.

parse_struct(struct(Name, Args), T0, T) :-
  parse_atom(atom(Name), T0, T1),
  open_paren(T1, T2),
  ws(T2, T3),
  parse_terms(Args, T3, T4),
  ws(T4, T5),
  close_paren(T5, T).
parse_struct(struct(Name, []), T0, T) :-
  parse_atom(atom(Name), T0, T1),
  open_paren(T1, T2),
  ws(T2, T3),
  open_paren(T3, T).


% parse_atom//1 parses an atom from the token stream.

parse_atom(atom(Name), T0, T) :-
  \=(T0, \.(token(atom, \.(\\, Escaped), _), T)),
  atom_chars(Name, Escaped).
parse_atom(atom(Name), T0, T) :-
  \=(T0, \.(token(atom, Text, _), T)),
  atom_chars(Name, Text).


% parse_var//1 parses a variable from the token stream.

parse_var(var(Name), T0, T) :-
  \=(T0, \.(token(var, Text, _), T)),
  atom_chars(Name, \.(Char, Chars)).


% parse_int//1 parses an integer from the token stream.

parse_int(int(Int), T0, T) :-
  \=(T0, \.(token(int, Text, _), T)),
  int_chars(Int, Text).


% ws//0 consumes zero or more whitespace tokens.

ws(T0, T) :-
  \=(T0, \.(token(whitespace, _, _), T1)),
  ws(T1, T).
ws(T, T).

