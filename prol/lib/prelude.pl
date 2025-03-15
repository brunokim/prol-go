comment_(L0, L) :-
  \=(L0, \.(\%, L1)),
  comment_chars_(L1, L).

comment_chars_(L0, L) :-
  \=(L0, \.(Char, L1)),
  neq(Char, \
),
  comment_chars_(L1, L).
comment_chars_(L0, L) :-
  \=(L0, \.(\
, L)).

ws_(L0, L) :-
  comment_(L0, L1),
  ws_(L1, L).
directive() :- move_clause_in_predicate(ws_(2), 2, 3).

% We can use comments now!

% quoted_atom_chars//1 reads atoms like 'this'.
% The atom can contain any character. Single quotes are escaped with ''.
%
quoted_atom_chars(\.(\', Chars), L0, L) :-
  \=(L0, \.(\', \.(\', L1))),
  quoted_atom_chars(Chars, L1, L).
quoted_atom_chars(\.(Char, Chars), L0, L) :-
  \=(L0, \.(Char, L1)),
  neq(Char, \'),
  quoted_atom_chars(Chars, L1, L).
quoted_atom_chars([], L0, L) :-
  \=(L0, \.(\', L)).

% Add quoted_atom_chars as a parsing option to atoms.
atom_(atom(Name), L0, L) :-
  \=(L0, \.(\', L1)),
  quoted_atom_chars(Chars, L1, L),
  atom_chars_(Name, Chars).

test_atom_('with nested -->''<-- single quotes').

% Parse lists.
list_(List, L0, L) :-
  '='(L0, '.'('[', L1)),
  ws_(L1, L2),
  terms_(List, L2, L3),
  ws_(L3, L4),
  '='(L4, '.'(']', L)).
list_([], L0, L) :-
  '='(L0, '.'('[', L1)),
  ws_(L1, L2),
  '='(L2, '.'(']', L)).

term_(Term, L0, L) :-
  list_(Term, L0, L).

test_list_([a, X, f(b)]).
