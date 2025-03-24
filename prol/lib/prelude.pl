doc(prelude, extends_the_language_syntax_further, using_only_itself).


doc(line_comment, parses_a_line_comment).

line_comment(L0, L) :-
  \=(L0, \.(\%, L1)),
  line_comment_chars(L1, L).

line_comment_chars(L0, L) :-
  \=(L0, \.(Char, L1)),
  neq(Char, \
),
  line_comment_chars(L1, L).
line_comment_chars(L0, L) :-
  \=(L0, \.(\
, L)).


doc(registers_line_comment_as_whitespace).
doc(after_this, its_necessary_to_modify_the_order_of_clauses_within_the_ws_predicate).

ws(L0, L) :-
  line_comment(L0, L1),
  ws(L1, L).

directive() :-
  get_predicate(indicator(ws, 2), \.(C1, \.(C2, \.(C3, [])))),
  put_predicate(indicator(ws, 2), \.(C1, \.(C3, \.(C2, [])))).


% We can use line comments now!
% Now we can stop using these clumsy doc() facts to provide documentation.
% Let's delete them now.

directive() :-
  put_predicate(indicator(doc, 1), []),
  put_predicate(indicator(doc, 2), []),
  put_predicate(indicator(doc, 3), []).


% C-style comments.

c_comment(L0, L) :-
  \=(L0, \.(\/, \.(\*, L1))),
  c_comment_chars(L1, L).

c_comment_chars(L0, L) :-
  \=(L0, \.(Char, L1)),
  neq(Char, \*),
  c_comment_chars(L1, L).
c_comment_chars(L0, L) :-
  \=(L0, \.(\*, \.(Char, L1))),
  neq(Char, \/),
  c_comment_chars(L1, L).
c_comment_chars(L0, L) :-
  \=(L0, \.(\*, \.(\/, L))).

% Register C-style comments as whitespace in the penultimate predicate position.

ws(L0, L) :-
  c_comment(L0, L1),
  ws(L1, L).

directive() :-
  get_predicate(indicator(ws, 2), \.(C1, \.(C2, \.(C3, \.(C4, []))))),
  put_predicate(indicator(ws, 2), \.(C1, \.(C2, \.(C4, \.(C3, []))))).

/* WE CAN USE C-STYLE COMMENTS NOW, TOO! */


% Now, on to parsing lists!
% Lists are just linked list of dotted pairs, i.e., the struct .(Head, Tail).
% We don't need a special form to represent it.

parse_list(atom([]), L0, L) :- /* Empty list */
  \=(L0, \.(\[, L1)),
  ws(L1, L2),
  \=(L2, \.(\], L)).
parse_list(struct(\., \.(H, \.(T, []))), L0, L) :-
  \=(L0, \.(\[, L1)),
  ws(L1, L2),
  parse_term(H, L2, L3),
  ws(L3, L4),
  parse_tail(T, L4, L).

parse_tail(atom([]), L0, L) :-
  % Read ']'
  \=(L0, \.(\], L)).
parse_tail(Tail, L0, L) :-
  % Read '| Tail ]'
  \=(L0, \.(\|, L1)),
  ws(L1, L2),
  parse_term(Tail, L2, L3),
  ws(L3, L4),
  \=(L4, \.(\], L)).
parse_tail(struct(\., \.(X, \.(Xs, []))), L0, L) :- /* struct('.', [X, Xs]) */
  % Read ', Term ...'
  \=(L0, \.(\, , L1)),
  ws(L1, L2),
  parse_term(X, L2, L3),
  ws(L3, L4),
  parse_tail(Xs, L4, L).

% Register parse_list in parse_term.

parse_term(Term, L0, L) :-
  parse_list(Term, L0, L).

test_parse_list([ ], [/*comment*/], [1], [ 1 ], [1, 2], [1|2], [1, 2|X], [1|[2|[3|[]]]]).


% Now let's parse quoted atoms, and get rid of our non-standard escape
% for any char with backslash.
% We will also parse strings as lists of atomic chars.


% parse_quoted_chars//2 is a generic parsing utility for quoted atoms and strings.
% It supports a single escape for the quote itself, which must be duplicated.
% All other chars must be included as-is in the source file.

parse_quoted_chars(Quote, Chars, L0, L) :-
  \=(L0, [Quote|L1]),
  parse_quoted_chars0(Quote, Chars, L1, L).

parse_quoted_chars0(Quote, [Char|Chars], L0, L) :-
  \=(L0, [Char|L1]),
  neq(Char, Quote),
  parse_quoted_chars0(Quote, Chars, L1, L).
parse_quoted_chars0(Quote, [Quote|Chars], L0, L) :-
  \=(L0, [Quote, Quote|L1]),
  parse_quoted_chars0(Quote, Chars, L1, L).
parse_quoted_chars0(Quote, [], L0, L) :-
  \=(L0, [Quote|L]).


% parse_quoted_atom//1 parses a quoted char using single quotes.

parse_quoted_atom(atom(Name), L0, L) :-
  parse_quoted_chars(\', Chars, L0, L),
  atom_chars(Name, Chars).

parse_atom(Atom, L0, L) :-
  parse_quoted_atom(Atom, L0, L).

test_parse_quoted_atom('a', ' ', '''').


% parse_quoted_string//1 parses a list of atom chars.

parse_quoted_string(ListAST, L0, L) :-
  parse_quoted_chars('"', Chars, L0, L),
  atom_list_to_ast(Chars, ListAST).

atom_list_to_ast([Char|Chars], struct('.', [H, T])) :-
  '='(H, atom(Char)),
  atom_list_to_ast(Chars, T).
atom_list_to_ast([], atom([])).

parse_list(List, L0, L) :-
  parse_quoted_string(List, L0, L).

test_parse_quoted_string("", "a", "1 2 3").


% Let's remove the \ escape for single-char atoms, which is non-standard.

directive :-
  get_predicate(indicator(parse_atom, 3), [C1, C2, C3, C4]),
  put_predicate(indicator(parse_atom, 3), [C1, C3, C4]).


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
