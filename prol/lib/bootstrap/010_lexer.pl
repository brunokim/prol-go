% This is the language lexer.
% It transforms a list of characters into a list of tokens.


% The lexer keeps a small bit of state, namely, the position in the char list.
% The state updates at each new char read.

initial_lex_state(0).

update_lex_state(Pos, Pos1) :-
  is(Pos1, \+(Pos, 1)).

read_char(Char, S0, S, \.(Char, L), L) :-
  update_lex_state(S0, S).


% lex_tokens/2 transforms a list of chars into tokens.
% A token is represented by token(Type, Text, State).
% - Type is the type of token (atom, var, int, whitespace, punctuation).
% - Text is the text of the token.
% - State is the lexer state before reading the token.

lex_tokens(Text, Tokens) :-
  initial_lex_state(S0),
  lex_tokens(Tokens, S0, _, Text, []).

lex_tokens(\.(Token, Tokens), S0, S, L0, L) :-
  lex_token(Token, S0, S1, L0, L1),
  lex_tokens(Tokens, S1, S, L1, L).
lex_tokens(\.(token(end, [], S), []), S, S, L, L).

lex_token(Token, S0, S, L0, L) :-
  lex_atom(Token, S0, S, L0, L).
lex_token(Token, S0, S, L0, L) :-
  lex_var(Token, S0, S, L0, L).
lex_token(Token, S0, S, L0, L) :-
  lex_int(Token, S0, S, L0, L).
lex_token(Token, S0, S, L0, L) :-
  lex_whitespace(Token, S0, S, L0, L).
lex_token(Token, S0, S, L0, L) :-
  lex_punctuation(Token, S0, S, L0, L).


% lex_atom//3 extracts an atom from the character list. An atom may be an arbitrary number of
% identifier characters, starting with lowercase; or a single character prepended by a backslash; or
% a nil atom (i.e., []).
% 
% Later, we will also accept single-quoted atoms.

lex_atom(token(atom, \.(Char, Chars), S0), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  atom_start(Char),
  lex_identifier_chars(Chars, S1, S, L1, L).
lex_atom(token(atom, \.(\\, Char), S0), S0, S, L0, L) :-
  read_char(\\, S0, S1, L0, L1),
  read_char(Char, S1, S, L1, L).
lex_atom(token(atom, \.(\[, \.(\], [])), S0), S0, S, L0, L) :-
  read_char(\[, S0, S1, L0, L1),
  read_char(\], S1, S, L1, L).


% lex_var//3 extracts a variable from the character list. A variable may be an arbitrary number of
% identifier characters, starting with an uppercase letter or an underscore.

lex_var(token(var, \.(Char, Chars), S0), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  var_start(Char),
  lex_identifier_chars(Chars, S1, S, L1, L).


% lex_int//3 extracts an integer from the character list. An integer may be an arbitrary number of digits.

lex_int(token(int, \.(Char, Chars), S0), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  ascii_digit(Char),
  lex_digits(Chars, S1, S, L1, L).


% lex_whitespace//3 extracts whitespace from the character list. Whitespace may be an arbitrary number of
% spaces or newlines; or a line comment starting with '%' and ending with a newline.
%
% Later, we will also accept block comments starting with '/*' and ending with '*/'.

lex_whitespace(token(whitespace, \.(Char, Chars), S0), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  space(Char),
  lex_spaces(Chars, S1, S, L1, L).
lex_whitespace(token(whitespace, \.(Char, Chars), S0), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  line_comment_start(Char),
  lex_line_comment(Chars, S1, S, L1, L).


% lex_punctuation//3 extracts punctuation from the character list.

lex_punctuation(token(open_paren, \.(\(, []), S0), S0, S, L0, L) :-
  read_char(\(, S0, S, L0, L).
lex_punctuation(token(close_paren, \.(\), []), S0), S0, S, L0, L) :-
  read_char(\), S0, S, L0, L).
lex_punctuation(token(comma, \.(\,, []), S0), S0, S, L0, L) :-
  read_char(\,, S0, S, L0, L).
lex_punctuation(token(full_stop, \.(\., []), S0), S0, S, L0, L) :-
  read_char(\., S0, S, L0, L).
lex_punctuation(token(implied_by, \.(\:, \.(\-, [])), S0), S0, S, L0, L) :-
  read_char(\:, S0, S1, L0, L1).
  read_char(\-, S1, S, L1, L).


% The following predicates tokenize repeated characters of the same kind.

lex_identifier_chars(\.(Char, Chars), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  identifier_char(Char),
  lex_identifier_chars(Chars, S1, S, L1, L).
lex_identifier_chars([], S, S, L, L).

lex_spaces(\.(Char, Chars), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  space(Char),
  lex_spaces(Chars, S1, S, L1, L).
lex_spaces([], S, S, L, L).

lex_digits(\.(Char, Chars), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  ascii_digit(Char),
  lex_digits(Chars, S1, S, L1, L).
lex_digits([], S, S, L, L).

lex_line_comment(\.(Char, Chars), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  neq(Char, \
),
  lex_line_comment(Chars, S1, S, L1, L).
lex_line_comment(\.(Char, []), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  \=(Char, \
).


% Character classes.

atom_start(Char) :-
  ascii_lower(Char).

var_start(\_).
var_start(Char) :-
  ascii_upper(Char).

identifier_char(\_).
identifier_char(Char) :-
  ascii_lower(Char).
identifier_char(Char) :-
  ascii_upper(Char).
identifier_char(Char) :-
  ascii_digit(Char).

