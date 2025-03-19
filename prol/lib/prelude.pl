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

directive() :-
  clause(comment(_, _), Comment),
  predicate_state(ws(_, _), \.(C1, \.(C2, [])), \.(C1, \.(Comment, \.(C2, [])))).
