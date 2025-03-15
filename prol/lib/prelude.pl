comment_(L0, L) :-
  \=(L0, \.(\%, L1)),
  comment_chars_(L1, L).

comment_chars_(L0, L) :-
  \=(L0, \.(\
, L)).
comment_chars_(L0, L) :-
  \=(L0, \.(Char, L1)),
  comment_chars_(L1, L).

ws_(L0, L) :-
  comment_(L0, L1),
  ws_(L1, L).

% We can use comments now!

% quoted_atom_chars//1 reads atoms like 'this'.
% The atom can contain any character. Single quotes are escaped with ''.
%
quoted_atom_chars(\.(\', Chars), L0, L) :-
  \=(L0, \.(\', \.(\', L1))),
  quoted_atom_chars(Chars, L1, L).
quoted_atom_chars([], L0, L) :-
  \=(L0, \.(\', L)).
quoted_atom_chars(\.(Char, Chars), L0, L) :-
  \=(L0, \.(Char, L1)),
  quoted_atom_chars(Chars, L1, L).

% Add quoted_atom_chars as a parsing option to atoms.
atom_(atom(Name), L0, L) :-
  \=(L0, \.(\', L1)),
  quoted_atom_chars(Chars, L1, L),
  atom_chars_(Name, Chars).

test_atom_('with nested -->''<-- single quotes').

