query().

atom_chars(Atom, Chars) :-
  var(Atom),
  is_char_list(Chars),
  chars_to_atom(Chars, Atom).
atom_chars(Atom, Chars) :-
  atom(Atom),
  atom_to_chars(Atom, Chars).

int_chars(Int, Chars) :-
  var(Int),
  is_char_list(Chars),
  chars_to_int(Chars, Int).
int_chars(Int, Chars) :-
  int(Int),
  int_to_chars(Int, Chars).

is_char_list(\.(Char, Chars)) :-
  atom_length(Char, 1),
  is_char_list(Chars).
is_char_list([]).

database(\.(Rule, Rules), L0, L) :-
  rule(Rule, L0, L1),
  ws(L1, L2),
  database(Rules, L2, L).
database(\.(Rule, []), L0, L) :-
  rule(Rule, L0, L).

rule(Rule, L0, L) :-
  clause(Rule, L0, L).

clause(clause(Head, Body), L0, L) :-
  struct(Head, L0, L1),
  ws(L1, L2),
  \=(L2, \.(\:, \.(\-, L3))),
  ws(L3, L4),
  terms(Body, L4, L5),
  ws(L5, L6),
  \=(L6, \.(\., L)).
clause(clause(Head, []), L0, L) :-
  struct(Head, L0, L1),
  ws(L1, L2),
  \=(L2, \.(\., L)).

terms(\.(Term, Terms), L0, L) :-
  term(Term, L0, L1),
  ws(L1, L2),
  \=(L2, \.(\, , L3)),
  ws(L3, L4),
  terms(Terms, L4, L).
terms(\.(Term, []), L0, L) :-
  term(Term, L0, L).

term(Term, L0, L) :-
  struct(Term, L0, L).
term(Term, L0, L) :-
  atom(Term, L0, L).
term(Term, L0, L) :-
  var(Term, L0, L).
term(Term, L0, L) :-
  int(Term, L0, L).

struct(struct(Name, Args), L0, L) :-
  atom(atom(Name), L0, L1),
  \=(L1, \.(\(, L2)),
  ws(L2, L3),
  terms(Args, L3, L4),
  ws(L4, L5),
  \=(L5, \.(\), L)).
struct(struct(Name, []), L0, L) :-
  atom(atom(Name), L0, L1),
  \=(L1, \.(\(, L2)),
  ws(L2, L3),
  \=(L3, \.(\), L)).

atom(atom(Name), L0, L) :-
  \=(L0, \.(Char, L1)),
  atom_start(Char),
  ident_chars(Chars, L1, L),
  atom_chars(Name, \.(Char, Chars)).
atom(atom(Name), L0, L) :-
  \=(L0, \.(\\, \.(Char, L))),
  atom_chars(Name, \.(Char, [])).
atom(atom(Name), L0, L) :-
  \=(L0, \.(\[, \.(\], L))),
  atom_chars(Name, \.(\[, \.(\], []))).

var(var(Name), L0, L) :-
  \=(L0, \.(Char, L1)),
  var_start(Char),
  ident_chars(Chars, L1, L),
  atom_chars(Name, \.(Char, Chars)).

int(int(Int), L0, L) :-
  \=(L0, \.(Char, L1)),
  ascii_digit(Char),
  print(int),
  print(Char),
  digits(Chars, L1, L),
  print(Chars),
  int_chars(Int, \.(Char, Chars)).

ident_chars(\.(Char, Chars), L0, L) :-
  \=(L0, \.(Char, L1)),
  ident(Char),
  ident_chars(Chars, L1, L).
ident_chars([], L, L).

digits(\.(Char, Chars), L0, L) :-
  \=(L0, \.(Char, L1)),
  ascii_digit(Char),
  digits(Chars, L1, L).
digits([], L, L).

ws(L0, L) :-
  \=(L0, \.(Char, L1)),
  space(Char),
  ws(L1, L).
ws(L, L).

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
