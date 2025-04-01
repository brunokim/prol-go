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


op(700, xfx, <).   % Less than
op(700, xfx, =<).  % Less or equal to
op(700, xfx, >).   % Greater than
op(700, xfx, >=).  % Greater or equal to
op(700, xfx, =).   % Unifies
op(700, xfx, \=).  % Does not unify
op(700, xfx, ==).  % Equivalent to
op(700, xfx, \==). % Not equivalent to
op(700, xfx, is).  % Arithmetic evaluation
op(500, yfx, +).   % Addition
op(500, yfx, -).   % Subtraction
op(400, yfx, *).   % Multiplication
op(400, yfx, /).   % Division (quotient)
op(400, yfx, mod). % Remainder of division
op(200, xfy, ^).   % Power to
op(200, fy, +).    % Positive (unary)
op(200, fy, -).    % Negative (unary)

% op_type_position(Type, Position) relates an operator type with its position.
op_type_position(fx, prefix).
op_type_position(fy, prefix).
op_type_position(xf, suffix).
op_type_position(yf, suffix).
op_type_position(yfx, infix).
op_type_position(xfy, infix).
op_type_position(xfx, infix).

% op_type_associativity(Type, Associativity) relates an operator type with its associativity.
op_type_associativity(fx, none).
op_type_associativity(fy, right).
op_type_associativity(xf, none).
op_type_associativity(yf, left).
op_type_associativity(yfx, left).
op_type_associativity(xfy, right).
op_type_associativity(xfx, none).

% left_precedence(OpPrecedence, Type, LeftPrecedence) provides the maximum precedence for
% the left argument of an operator with given Type and OpPrecedence.
left_precedence(Prec, yf, Prec).
left_precedence(Prec, yfx, Prec).
left_precedence(Prec0, xf, Prec) :-
  is(Prec, -(Prec0, 1)).
left_precedence(Prec0, xfy, Prec) :-
  is(Prec, -(Prec0, 1)).
left_precedence(Prec0, xfx, Prec) :-
  is(Prec, -(Prec0, 1)).

% right_precedence(OpPrecedence, Type, RightPrecedence) provides the maximum precedence for
% the right argument of an operator with given Type and OpPrecedence.
right_precedence(Prec, fy, Prec).
right_precedence(Prec, xfy, Prec).
right_precedence(Prec0, fx, Prec) :-
  is(Prec, -(Prec0, 1)).
right_precedence(Prec0, yfx, Prec) :-
  is(Prec, -(Prec0, 1)).
right_precedence(Prec0, xfx, Prec) :-
  is(Prec, -(Prec0, 1)).

% parse_atomic_term//1 parses expression literals, or a parenthesized expression.
parse_atomic_term(Term) --> parse_struct(Term).
parse_atomic_term(Term) --> parse_atom(Term).
parse_atomic_term(Term) --> parse_var(Term).
parse_atomic_term(Term) --> parse_int(Term).
parse_atomic_term(Term) --> parse_list(Term).
parse_atomic_term(Term) -->
  "(",
  ws,
  parse_expr(Term),
  ws,
  ")".

% parse_expr//1 parses an expression with operators.
parse_expr(Term) -->
  parse_leaf(Leaf),
  parse_infix(Leaf, Term),
  print(expr),
  print(Term).

% parse_leaf//1 parses an expression like 'prefix_op* atomic_term suffix_op*'
parse_leaf(Term) -->
  parse_prefix(1200, Term0),
  parse_suffix(Term0, Term),
  print(leaf),
  print(Term).

% parse_prefix//2 parses a prefix operator with precedence at most Prec0.
% The base case is parsing an atomic term.
parse_prefix(Prec0, Term) -->
  parse_atom(atom(Token)),
  { op(Prec1, Type, Token),
    >=(Prec0, Prec1),
    op_type_position(Type, prefix),
    right_precedence(Prec1, Type, Prec2) },
  ws,
  parse_prefix(Prec2, Term0),
  { =(Term, expr(nil, op(Prec1, Type, Token), Term0)) },
  print(prefix),
  print(Term).
parse_prefix(_, Term) -->
  parse_atomic_term(Term),
  print(atomic),
  print(Term).

% parse_suffix//2 parses a suffix operator given a Left tree.
% The operator is inserted at the appropriate position to satisfy precedence.
% The base case is outputting the provided Left tree.
parse_suffix(Left, Term) -->
  ws,
  parse_atom(atom(Token)),
  { op(Prec, Type, Token),
    op_type_position(Type, suffix),
    insert_right(Left, op(Prec, Type, Token), nil, Term0) },
  parse_suffix(Term0, Term),
  print(suffix),
  print(Term).
parse_suffix(Term, Term) --> [].

% parse_infix//2 parses an infix operator with a Left tree followed by a leaf expression Right.
% The operator is inserted at the appropriate position to satisfy precedence.
% The base case is outputting the provided Left tree.
parse_infix(Left, Term) -->
  ws,
  parse_atom(atom(Op)),
  { op(Prec, Type, Op),
    op_type_position(Type, infix) },
  ws,
  parse_leaf(Right),
  { insert_right(Left, op(Prec, Type, Op), Right, Term0) },
  parse_infix(Term0, Term),
  print(infix),
  print(Term).
parse_infix(Term, Term).

% insert_right(Left, Op, Arg, Term) inserts the given Arg at the Left tree using the operator Op.
% It outputs the result in Term.
%
% For example, given Left = (1 + 2), Op = +, Arg = 3, then Term = ((1 + 2) + 3)
% Likewise, given Left = (1 + 2), Op = *, Arg = 3, then Term = (1 + (2 * 3))
insert_right(Expr, Op1, Arg, Term) :-
  % Inserting operator with higher precedence than left tree.
  =(Expr, expr(Left, Op2, Right)),
  check_precedence(Op1, Op2, left),
  =(Term, expr(Expr, Op1, Arg)).
insert_right(Expr, Op1, Arg, Term) :-
  % Inserting operator with lower precedence than left tree.
  =(Expr, expr(Left, Op2, Right)),
  check_precedence(Op2, Op1, left),
  =(Term, expr(Left, Op2, expr(Right, Op1, Arg))).
insert_right(Expr, Op1, Arg, Term) :-
  % Inserting operator with conflicting precedence with left tree: go down one level.
  =(Expr, expr(Left, Op2, Right)),
  insert_right(Right, Op1, Arg, Right0),
  =(Term, expr(Left, Op2, Right0)).

% check_precedence(Op1, Op2, Pos) checks that Op1 can have Op2 as child in position Pos.
% If Pos corresponds to Op1's associativity, Op2's precedence may be greater than or equal to Op1's precedence.
% Otherwise, Op2's precedence needs to be strictly greather than Op1's precedence.
check_precedence(op(Prec1, Type1, _), op(Prec2, _, _), Pos) :-
  op_type_associativity(Type1, Pos),
  >=(Prec1, Prec2).
check_precedence(op(Prec1, _, _), op(Prec2, _, _), _) :-
  >(Prec1, Prec2).

% Makes parse_term an alias to parse_expr.
:- put_predicate(indicator(parse_term, 3), [
     dcg(struct(parse_term, [var('Term')]), [struct(parse_expr, [var('Term')])])
   ]).

test_parse_expr.
test_parse_expr(1).
test_parse_expr(1, a, X, f(g, h), [c, d]).
test_parse_expr((1), ( 1 ), f((g)), +(1,2)).
test_parse_expr(+ 2, - 1, +2, -1).
test_parse_expr(- -1, + -1, + +2).
