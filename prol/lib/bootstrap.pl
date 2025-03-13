query().

atom_chars_(Atom, Chars) :-
  var(Atom),
  is_char_list_(Chars),
  chars_to_atom(Chars, Atom).
atom_chars_(Atom, Chars) :-
  atom(Atom),
  atom_to_chars(Atom, Chars).

is_char_list_(\.(Char, Chars)) :-
  atom_length(Char, 1),
  is_char_list_(Chars).
is_char_list_([]).

database_(\.(Rule, Rules), L0, L) :-
  rule_(Rule, L0, L1),
  ws_(L1, L2),
  database_(Rules, L2, L).
database_(\.(Rule, []), L0, L) :-
  rule_(Rule, L0, L).

rule_(Rule, L0, L) :-
  clause_(Rule, L0, L).

clause_(clause(Head, Body), L0, L) :-
  struct_(Head, L0, L1),
  ws_(L1, L2),
  \=(L2, \.(\:, \.(\-, L3))),
  ws_(L3, L4),
  terms_(Body, L4, L5),
  ws_(L5, L6),
  \=(L6, \.(\., L)).
clause_(clause(Head, []), L0, L) :-
  struct_(Head, L0, L1),
  ws_(L1, L2),
  \=(L2, \.(\., L)).

terms_(\.(Term, Terms), L0, L) :-
  term_(Term, L0, L1),
  ws_(L1, L2),
  \=(L2, \.(\, , L3)),
  ws_(L3, L4),
  terms_(Terms, L4, L).
terms_(\.(Term, []), L0, L) :-
  term_(Term, L0, L).

term_(Term, L0, L) :-
  struct_(Term, L0, L).
term_(Term, L0, L) :-
  atom_(Term, L0, L).
term_(Term, L0, L) :-
  var_(Term, L0, L).

struct_(struct(Name, Args), L0, L) :-
  atom_(atom(Name), L0, L1),
  \=(L1, \.(\(, L2)),
  ws_(L2, L3),
  terms_(Args, L3, L4),
  ws_(L4, L5),
  \=(L5, \.(\), L)).
struct_(struct(Name, []), L0, L) :-
  atom_(atom(Name), L0, L1),
  \=(L1, \.(\(, L2)),
  ws_(L2, L3),
  \=(L3, \.(\), L)).

atom_(atom(Name), L0, L) :-
  \=(L0, \.(Char, L1)),
  atom_start(Char),
  ident_chars_(Chars, L1, L),
  atom_chars_(Name, \.(Char, Chars)).
atom_(atom(Name), L0, L) :-
  \=(L0, \.(\\, \.(Char, L))),
  atom_chars_(Name, \.(Char, [])).
atom_(atom(Name), L0, L) :-
  \=(L0, \.(\[, \.(\], L))),
  atom_chars_(Name, \.(\[, \.(\], []))).

var_(var(Name), L0, L) :-
  \=(L0, \.(Char, L1)),
  var_start(Char),
  ident_chars_(Chars, L1, L),
  atom_chars_(Name, \.(Char, Chars)).

ident_chars_(\.(Char, Chars), L0, L) :-
  \=(L0, \.(Char, L1)),
  ident(Char),
  ident_chars_(Chars, L1, L).
ident_chars_([], L, L).

ws_(L0, L) :-
  \=(L0, \.(Char, L1)),
  space(Char),
  ws_(L1, L).
ws_(L, L).

atom_start(Char) :-
  lower(Char).
atom_start(Char) :-
  digit(Char).

var_start(\_).
var_start(Char) :-
  upper(Char).

ident(\_).
ident(Char) :-
  lower(Char).
ident(Char) :-
  upper(Char).
ident(Char) :-
  digit(Char).

digit(\0).
digit(\1).
digit(\2).
digit(\3).
digit(\4).
digit(\5).
digit(\6).
digit(\7).
digit(\8).
digit(\9).

lower(\a).
lower(\b).
lower(\c).
lower(\d).
lower(\e).
lower(\f).
lower(\g).
lower(\h).
lower(\i).
lower(\j).
lower(\k).
lower(\l).
lower(\m).
lower(\n).
lower(\o).
lower(\p).
lower(\q).
lower(\r).
lower(\s).
lower(\t).
lower(\u).
lower(\v).
lower(\w).
lower(\x).
lower(\y).
lower(\z).

upper(\A).
upper(\B).
upper(\C).
upper(\D).
upper(\E).
upper(\F).
upper(\G).
upper(\H).
upper(\I).
upper(\J).
upper(\K).
upper(\L).
upper(\M).
upper(\N).
upper(\O).
upper(\P).
upper(\Q).
upper(\R).
upper(\S).
upper(\T).
upper(\U).
upper(\V).
upper(\W).
upper(\X).
upper(\Y).
upper(\Z).

space(\ ).
space(\
).
