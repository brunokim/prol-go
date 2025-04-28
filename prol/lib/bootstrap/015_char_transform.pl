% atom_chars/2 converts between an atom and its list of characters representation.

atom_chars(Atom, Chars) :-
  var(Atom),
  is_char_list(Chars),
  chars_to_atom(Chars, Atom).
atom_chars(Atom, Chars) :-
  atom(Atom),
  atom_to_chars(Atom, Chars).


% int_chars/2 converts between an integer and its list of characters representation.

int_chars(Int, Chars) :-
  var(Int),
  is_char_list(Chars),
  chars_to_int(Chars, Int).
int_chars(Int, Chars) :-
  int(Int),
  int_to_chars(Int, Chars).


% is_char_list/2 checks whether the list is composed only of one-char atoms.

is_char_list(\.(Char, Chars)) :-
  atom_length(Char, 1),
  is_char_list(Chars).
is_char_list([]).

