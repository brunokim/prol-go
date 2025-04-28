doc(this_is_the_language_lexer).
doc(it_transforms_a_list_of_characters_into_a_list_of_tokens).


doc(the_lexer_keeps_a_small_bit_of_state, namely, the_position_in_the_char_list).
doc(the_state_updates_at_each_new_char_read).

initial_lex_state(1).

update_lex_state(Pos, Pos1) :-
  is(Pos1, \+(Pos, 1)).

read_char(Char, S0, S, \.(Char, L), L) :-
  update_lex_state(S0, S).


doc(lex_tokens, transforms_a_list_of_chars_into_tokens).
doc(each_token_has_three_arguments, token_type, text, and_the_lexer_state).

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


doc(lex_atom, extracts_an_atom_of_an_arbitrary_number_of_identifier_chars).
doc(starting_with_lowercase, a_single_char_prepended_by_backslash, or).
doc(a_nil_atom).

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


doc(lex_var, extracts_an_var_of_an_arbitrary_number_of_identifier_chars).
doc(starting_with_uppercase, or_an_underscore).

lex_var(token(var, \.(Char, Chars), S0), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  var_start(Char),
  lex_identifier_chars(Chars, S1, S, L1, L).


doc(lex_int, extracts_an_integer_with_an_arbitrary_number_of_digits).

lex_int(token(int, \.(Char, Chars), S0), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  ascii_digit(Char),
  lex_digits(Chars, S1, S, L1, L).


doc(lex_whitespace, creates_a_token_for_space_characters).
doc(this_will_be_useful_to_determine_newline_positions).
doc(as_well_as_keep_tabs_on_comments, when_we_implement_them).

lex_whitespace(token(whitespace, \.(Char, Chars), S0), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  space(Char),
  lex_spaces(Chars, S1, S, L1, L).


doc(lex_punctuation, tokenizes_some_characters_with_special_significance).

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


doc(the_following_predicates_tokenize_repeated_chars_of_the_same_kind).

lex_identifier_chars(\.(Char, Chars), S0, S, L0, L) :-
  read_char(Char, S0, S1, L0, L1),
  ident(Char),
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

