length(L, N) :-
  length_(L, 0, N).

length_([], Acc, Acc).
length_(\.(_, T), Acc0, N) :-
  is(Acc, \+(Acc0, 1)),
  length_(T, Acc, N).

append([], T, T).
append(\.(H, L0), L1, \.(H, L2)) :-
  append(L0, L1, L2).

indicator(struct(Name, Args), indicator(Name, NumArgs)) :-
  length(Args, NumArgs).

asserta(Clause) :-
  \=(Clause, clause(Head, _)),
  indicator(Head, Ind),
  get_predicate(Ind, Clauses0),
  put_predicate(Ind, \.(Clause, Clauses0)).

assertz(Clause) :-
  \=(Clause, clause(Head, _)),
  indicator(Head, Ind),
  get_predicate(Ind, Clauses0),
  append(Clauses0, \.(Clause, []), Clauses),
  put_predicate(Ind, Clauses).
