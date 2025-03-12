query().

nat(0).
nat(s(X)) :-
  nat(X).

add(0, X, X).
add(s(A), B, s(C)) :-
  add(A, B, C).

member(Elem, [H|T]) :-
  member_(T, Elem, H).

member_(_, Elem, Elem).
member_([H|T], Elem, _) :-
  member_(T, Elem, H).

atom_chars(Atom, Chars) :-
  var(Atom),
  is_char_list(Chars),
  'chars->atom'(Chars, Atom).
atom_chars(Atom, Chars) :-
  atom(Atom),
  'chars->atom'(Atom, Chars).

clause(clause(Head, Body)) -->
  struct(Head),
  ws,
  ":-",
  ws,
  structs(Body),
  ws,
  ".".
clause(clause(Head, [])) -->
  struct(Head),
  ws,
  ".".

structs([Struct|Structs]) -->
  struct(Struct),
  ws,
  ",",
  ws,
  structs(Structs).
structs([Struct]) -->
  struct(Struct).

term(Term) --> struct(Term).
term(Term) --> atom(Term).
term(Term) --> var(Term).

struct(Name, Args) -->
  atom(atom(Name)),
  "(",
  ws,
  terms(Args),
  ws,
  ")".
struct(Name, []) -->
  atom(atom(Name)),
  "(",
  ws,
  ")".

terms([Struct|Structs]) -->
  term(Struct),
  ws,
  ",",
  ws,
  terms(Structs).
terms([Struct]) -->
  term(Struct).

atom(atom(Name)) -->
  [Char],
  '{}'(atom_start(Char)),
  ident_chars(Chars),
  '{}'(atom_chars(Atom, [Char|Chars])).

var(var(Var)) -->
  [Char],
  '{}'(var_start(Char)),
  ident_chars(Chars),
  '{}'(atom_chars(Var, [Char|Chars])).

ident_chars([Char|Chars]) -->
  [Char],
  '{}'(ident(Char)),
  ident_chars(Chars).
ident_chars([]) --> [].

ws -->
  [Char],
  '{}'(space(Char)),
  ws.
ws --> [].

atom_start(Char) :-
  lower(Char).
atom_start(Char) :-
  digit(Char).
var_start('_').
var_start(Char) :-
  upper(Char).

ident('_').
ident(Char) :-
  lower(Char).
ident(Char) :-
  upper(Char).
ident(Char) :-
  digit(Char).

space(' ').
space('\n').
space('\t').
