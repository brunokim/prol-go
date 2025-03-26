doc(prelude, extends_the_language_syntax_further, using_only_itself).


doc(at_end_of_phrase, dcg_succeeds_if_we_are_at_the_last_position).

at_end_of_phrase([], []).


doc(peek, dcg_extracts_next_char_without_advancing_tokens).

peek(Char, \.(Char, L), \.(Char, L)).


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
line_comment_chars(L0, L) :-
  at_end_of_phrase(L0, L).


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


% parse_directive//1 parses a directive, that is, a rule for immediate execution.

parse_directive(clause(struct(directive, []), Goals)) -->
    ":-",
    ws,
    parse_goals(Goals),
    ws,
    ".".

parse_rule(Rule) --> parse_directive(Rule).

:- print('directive parsed').


/**
 * Expressions
 *
 * It's possible to express a computer program solely as functors, but it is a bit
 * contrary to our mathematical education and how most human languages are structured.
 *
 * For example, now we write
 *
 *     '='(T1, T2)
 *
 * but speak
 *
 *     T1 unifies with T2.
 *
 * and would like to write
 *
 *     T1 = T2.
 *
 * We will add suport to create expressions with any number of operators, with several
 * levels of precedence. An expression like
 *
 *     1 / 2 * 3 + 4 - 5 ^ -6 < +7 % 8
 *
 * will be parsed as if with the following parenthesis
 *
 *     ((((1 / 2) * 3) + 4) - (5 ^ (- 6))) < ((+ 7) % 8)
 *
 * and eventually become this nested functor:
 *
 *     '<'('-'('+'('*'('/'(1, 2), 3), 4), '^'(5, '-'(6))), '%'('+'(7), 8))
 */

% Our first step is allowing symbolic atoms like mathematical operators.
% Not all operators need to be symbolic, but many are.

ascii_symbol('=').
ascii_symbol('<').
ascii_symbol('>').
ascii_symbol('+').
ascii_symbol('-').
ascii_symbol('*').
ascii_symbol('/').
ascii_symbol('^').
ascii_symbol('\').

parse_symbol(atom(Name)) -->
  symbol_chars(Chars),
  { atom_chars(Name, Chars) }.

symbol_chars([Char|Chars]) -->
  [Char],
  { ascii_symbol(Char) },
  symbol_chars(Chars).
symbol_chars([Char]) -->
  [Char],
  { ascii_symbol(Char) }.

parse_atom(Atom) -->
  parse_symbol(Atom).


test_parse_symbol(=, ==, =<, >=, ++, **, -*/*-).

% Prolog allows for dynamic and user-defined operators.
% They must be registered as a fact op/3 like op(600, xfy, +), where the args mean
% - the operator precedence.
% - the operator position and associativity type
% - the operator atom
%
% An operator position may be suffix, prefix (arity 1), or infix (arity 2).
% It represents the valid positions where it may appear next to its arguments.
%
% The operator precedence determines which should be evaluated first when there's no
% parenthesis.
% For example, "2+3*4" is read as "2+(3*4)" because '*' has lower precedence than '+'.
%
% An operator associativity may be left, right, or none. It is used to disambiguate
% how to combine operators with same precedence, whether they must be parsed
% left-to-right or right-to-left.
% An operator with no associativity can't be combined with other operators with the
% same precedence.
%
% Infix examples:
% - left associativity: "2-3-4" is read as "(2-3)-4" (result: -5) and not as
%   "2-(3-4)" (result: 3)
% - right associativity: "2^3^4" is read as "2(3^4)" (result: 2^81) and not as
%   "(2^3)^4" (result: 8^4)
% - no associativity: "2<3" is an expression that results in a boolean, but
%   "2<3<4" has no obvious meaning, since it would compare a boolean with an integer.
%
% | Position | Associativity | Type |
% |----------|---------------|------|
% |    infix |          left |  yfx |
% |    infix |         right |  xfy |
% |    infix |          none |  xfx |
%
% Prefix operators are less common, and suffix operators even less. They can also
% have different associativity, representing whether they can be combined.
% For example, unary minus is (right) associative and "- - X" is read as "-(-(X))".
%
% | Position | Associativity | Type |
% |----------|---------------|------|
% |   prefix |         right |   fy |
% |   prefix |          none |   fx |
% |   suffix |          left |   yf |
% |   suffix |          none |   xf |

op(4, xfx, <).
op(4, xfx, =).
op(4, xfx, =<).
op(4, xfx, >).
op(4, xfx, >=).
op(4, xfx, \=).
op(4, xfx, \==).
op(4, xfx, is).
op(3, yfx, +).
op(3, yfx, -).
op(2, yfx, *).
op(2, yfx, /).
op(2, yfx, mod).
op(1, xfy, ^).
op(1, fy, +).
op(1, fy, -).

parse_atomic_term(Term) --> parse_struct(Term).
parse_atomic_term(Term) --> parse_atom(Term).
parse_atomic_term(Term) --> parse_var(Term).
parse_atomic_term(Term) --> parse_int(Term).
parse_atomic_term(Term) --> parse_list(Term).

parse_expr(Term) -->
  parse_expr(4, Term).

parse_expr(0, Term) -->
  "(",
  ws,
  parse_expr(Term),
  ws,
  ")".
parse_expr(0, Term) -->
  parse_atomic_term(Term).
parse_expr(Prec, Term) -->
  % fy: prefix operator with left associativity
  parse_atom(atom(Op)),
  not_followed_by_open_paren,
  { op(Prec, fy, Op) },
  ws,
  parse_expr(Prec, Arg),
  { =(Term, struct(Op, [Arg])) }.
parse_expr(Prec, Term) -->
  % fx: prefix operator without associativity
  parse_atom(atom(Op)),
  not_followed_by_open_paren,
  { op(Prec, fy, Op),
    is(Prec1, -(Prec, 1)) },
  ws,
  parse_expr(Prec1, Arg),
  { =(Term, struct(Op, [Arg])) }.
parse_expr(Prec, Term) -->
  /* fallthrough: parse expression with lower precedence. */
  { >(Prec, 0),
    is(Prec1, -(Prec, 1)) },
  parse_expr(Prec1, Term).

not_followed_by_open_paren -->
  peek(Char),
  { \==(Char, '(') }.
not_followed_by_open_paren -->
  at_end_of_phrase.

:- put_predicate(indicator(parse_term, 3), [
     dcg(struct(parse_term, [var('Term')]), [struct(parse_expr, [var('Term')])])
   ]).

test_parse_expr(1, a, X, f(g, h), [c, d]).
test_parse_expr((1), ( 1 ), f((g))).
test_parse_expr(+ 2, - 1, +2, -1).
test_parse_expr(- -1, + -1, + +2).
