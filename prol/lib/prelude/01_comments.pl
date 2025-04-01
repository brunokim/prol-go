doc(prelude, extends_the_language_syntax_further, using_only_itself).


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
