doc(this_is_the_bootstrap_for_prolog).
doc(we_dont_have_comments_yet, so_its_necessary_to_use_plain_atoms_for_documentation).


doc(we_call_queries_as_clauses_with_a, query(), head).
doc(we_need_to_have_this_head_defined_in_the_bootstrap).

query().


doc(atom_chars, converts_between_an_atom_and_a_list_of_chars).

atom_chars(Atom, Chars) :-
  var(Atom),
  is_char_list(Chars),
  chars_to_atom(Chars, Atom).
atom_chars(Atom, Chars) :-
  atom(Atom),
  atom_to_chars(Atom, Chars).


doc(int_chars, converts_between_an_int_and_a_list_of_chars).

int_chars(Int, Chars) :-
  var(Int),
  is_char_list(Chars),
  chars_to_int(Chars, Int).
int_chars(Int, Chars) :-
  int(Int),
  int_to_chars(Int, Chars).


doc(is_char_list, tests_whether_list_is_composed_only_of_one_char_atoms).

is_char_list(\.(Char, Chars)) :-
  atom_length(Char, 1),
  is_char_list(Chars).
is_char_list([]).


doc(parse_database, reads_a_sequence_of_rules_from_the_difference_list).

parse_database(\.(Rule, Rules), L0, L) :-
  parse_rule(Rule, L0, L1),
  ws(L1, L2),
  parse_database(Rules, L2, L).
parse_database(\.(Rule, []), L0, L) :-
  parse_rule(Rule, L0, L).


doc(parse_rule, parses_a_rule).
doc(for_now_we_only_have_one_rule_type,
    but_we_will_use_this_as_a_hook_for_other_types_later).

parse_rule(Rule, L0, L) :-
  parse_clause(Rule, L0, L).


doc(parse_clause, parses_a_fact_or_clause).

parse_clause(clause(Head, Body), L0, L) :-
  parse_goal(Head, L0, L1),
  ws(L1, L2),
  \=(L2, \.(\:, \.(\-, L3))),
  ws(L3, L4),
  parse_goals(Body, L4, L5),
  ws(L5, L6),
  \=(L6, \.(\., L)).
parse_clause(clause(Head, []), L0, L) :-
  parse_goal(Head, L0, L1),
  ws(L1, L2),
  \=(L2, \.(\., L)).


doc(parse_goals, parses_a_sequence_of_comma_separated_clause_goals).

parse_goals(\.(Goal, Goals), L0, L) :-
  parse_goal(Goal, L0, L1),
  ws(L1, L2),
  \=(L2, \.(\, , L3)),
  ws(L3, L4),
  parse_goals(Goals, L4, L).
parse_goals(\.(Goal, []), L0, L) :-
  parse_goal(Goal, L0, L).


doc(parse_goal, parses_a_goal_term_and_convert_to_a_struct).

parse_goal(Struct, L0, L) :-
  parse_struct(Struct, L0, L).
parse_goal(struct(Atom, []), L0, L) :-
  parse_atom(atom(Atom), L0, L).
parse_goal(struct(call, \.(X, [])), L0, L) :-
  parse_var(X, L0, L).


doc(parse_terms, parses_a_sequence_of_comma_separated_terms).

parse_terms(\.(Term, Terms), L0, L) :-
  parse_term(Term, L0, L1),
  ws(L1, L2),
  \=(L2, \.(\, , L3)),
  ws(L3, L4),
  parse_terms(Terms, L4, L).
parse_terms(\.(Term, []), L0, L) :-
  parse_term(Term, L0, L).


doc(parse_term, parses_a_term).

parse_term(Term, L0, L) :-
  parse_struct(Term, L0, L).
parse_term(Term, L0, L) :-
  parse_atom(Term, L0, L).
parse_term(Term, L0, L) :-
  parse_var(Term, L0, L).
parse_term(Term, L0, L) :-
  parse_int(Term, L0, L).


doc(parse_struct, parses_a_struct).

parse_struct(struct(Name, Args), L0, L) :-
  parse_atom(atom(Name), L0, L1),
  \=(L1, \.(\(, L2)),
  ws(L2, L3),
  parse_terms(Args, L3, L4),
  ws(L4, L5),
  \=(L5, \.(\), L)).
parse_struct(struct(Name, []), L0, L) :-
  parse_atom(atom(Name), L0, L1),
  \=(L1, \.(\(, L2)),
  ws(L2, L3),
  \=(L3, \.(\), L)).


doc(parse_atom, parses_an_atom).
doc(an_atom_may_be_a_symbol, escaped_char, or_nil_atom).

parse_atom(atom(Name), L0, L) :-
  \=(L0, \.(Char, L1)),
  atom_start(Char),
  ident_chars(Chars, L1, L),
  atom_chars(Name, \.(Char, Chars)).
parse_atom(atom(Name), L0, L) :-
  \=(L0, \.(\\, \.(Char, L))),
  atom_chars(Name, \.(Char, [])).
parse_atom(atom(Name), L0, L) :-
  \=(L0, \.(\[, \.(\], L))),
  atom_chars(Name, \.(\[, \.(\], []))).


doc(parse_var, parses_a_var).

parse_var(var(Name), L0, L) :-
  \=(L0, \.(Char, L1)),
  var_start(Char),
  ident_chars(Chars, L1, L),
  atom_chars(Name, \.(Char, Chars)).


doc(parse_int, parses_an_integer).

parse_int(int(Int), L0, L) :-
  \=(L0, \.(Char, L1)),
  ascii_digit(Char),
  digits(Chars, L1, L),
  int_chars(Int, \.(Char, Chars)).


doc(ident_chars, parses_a_sequence_of_identifier_chars).

ident_chars(\.(Char, Chars), L0, L) :-
  \=(L0, \.(Char, L1)),
  ident(Char),
  ident_chars(Chars, L1, L).
ident_chars([], L, L).


doc(digits, parses_a_sequence_of_digits).

digits(\.(Char, Chars), L0, L) :-
  \=(L0, \.(Char, L1)),
  ascii_digit(Char),
  digits(Chars, L1, L).
digits([], L, L).


doc(ws, parses_whitespace).

ws(L0, L) :-
  \=(L0, \.(Char, L1)),
  space(Char),
  ws(L1, L).
ws(L, L).


doc(facts_about_characters).

atom_start(Char) :-
  ascii_lower(Char).

var_start(\_).
var_start(Char) :-
  ascii_upper(Char).

ident(\_).
ident(Char) :-
  ascii_lower(Char).
ident(Char) :-
  ascii_upper(Char).
ident(Char) :-
  ascii_digit(Char).

ascii_digit(\0).
ascii_digit(\1).
ascii_digit(\2).
ascii_digit(\3).
ascii_digit(\4).
ascii_digit(\5).
ascii_digit(\6).
ascii_digit(\7).
ascii_digit(\8).
ascii_digit(\9).

ascii_lower(\a).
ascii_lower(\b).
ascii_lower(\c).
ascii_lower(\d).
ascii_lower(\e).
ascii_lower(\f).
ascii_lower(\g).
ascii_lower(\h).
ascii_lower(\i).
ascii_lower(\j).
ascii_lower(\k).
ascii_lower(\l).
ascii_lower(\m).
ascii_lower(\n).
ascii_lower(\o).
ascii_lower(\p).
ascii_lower(\q).
ascii_lower(\r).
ascii_lower(\s).
ascii_lower(\t).
ascii_lower(\u).
ascii_lower(\v).
ascii_lower(\w).
ascii_lower(\x).
ascii_lower(\y).
ascii_lower(\z).

ascii_upper(\A).
ascii_upper(\B).
ascii_upper(\C).
ascii_upper(\D).
ascii_upper(\E).
ascii_upper(\F).
ascii_upper(\G).
ascii_upper(\H).
ascii_upper(\I).
ascii_upper(\J).
ascii_upper(\K).
ascii_upper(\L).
ascii_upper(\M).
ascii_upper(\N).
ascii_upper(\O).
ascii_upper(\P).
ascii_upper(\Q).
ascii_upper(\R).
ascii_upper(\S).
ascii_upper(\T).
ascii_upper(\U).
ascii_upper(\V).
ascii_upper(\W).
ascii_upper(\X).
ascii_upper(\Y).
ascii_upper(\Z).

space(\ ).
space(\
).

