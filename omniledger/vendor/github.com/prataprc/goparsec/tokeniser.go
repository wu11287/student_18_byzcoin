// Copyright (c) 2013 Goparsec AUTHORS. All rights reserved.
// Use of this source code is governed by LICENSE file.

/*
This file provides a basic set of token parsers that can be used
to parse a single token or used to create higher order parsers using
the combinator functions. Unless specified, these parsers can be
used with AST combinators and return a Terminal node.
*/

package parsec

import "strings"
import "strconv"
import "unicode"
import "unicode/utf8"
import "unicode/utf16"

// String parse double quoted string in input text, this parser
// returns string type as ParsecNode, hence incompatible with
// AST combinators. Skip leading whitespace.
func String() Parser {
	return func(s Scanner) (ParsecNode, Scanner) {
		s.SkipWS()
		scanner := s.(*SimpleScanner)
		if !scanner.Endof() && scanner.buf[scanner.cursor] == '"' {
			str, readn := scanString(scanner.buf[scanner.cursor:])
			if str == nil || len(str) == 0 {
				return nil, scanner
			}
			scanner.cursor += readn
			return string(str), scanner
		}
		return nil, scanner
	}
}

// Char return parser function to match a single character
// in the input stream. Skip leading whitespace.
func Char() Parser {
	return Token(`'.'`, "CHAR")
}

// Float return parser function to match a float literal
// in the input stream. Skip leading whitespace.
func Float() Parser {
	return Token(`[+-]?([0-9]+\.[0-9]*|\.[0-9]+)`, "FLOAT")
}

// Hex return parser function to match a hexadecimal
// literal in the input stream. Skip leading whitespace.
func Hex() Parser {
	return Token(`0[xX][0-9a-fA-F]+`, "HEX")
}

// Oct return parser function to match an octal number
// literal in the input stream. Skip leading whitespace.
func Oct() Parser {
	return Token(`0[0-7]+`, "OCT")
}

// Int return parser function to match an integer literal
// in the input stream. Skip leading whitespace.
func Int() Parser {
	return Token(`-?[0-9]+`, "INT")
}

// Ident return parser function to match an identifier token
// in the input stream, an identifier is matched with the
// following pattern: `^[A-Za-z][0-9a-zA-Z_]*`.
// Skip leading whitespace.
func Ident() Parser {
	return Token(`[A-Za-z][0-9a-zA-Z_]*`, "IDENT")
}

// Token takes a regular-expression pattern and return a parser that
// will match input stream with supplied pattern. Skip leading whitespace.
// `name` will be used as the Terminal's name.
func Token(pattern string, name string) Parser {
	if pattern[0] != '^' {
		pattern = "^" + pattern
	}
	return func(s Scanner) (ParsecNode, Scanner) {
		news := s.Clone()
		news.SkipWS()
		cursor := news.GetCursor()
		if tok, _ := news.Match(pattern); tok != nil {
			return NewTerminal(name, string(tok), cursor), news
		}
		return nil, s
	}
}

// TokenExact same as Token() but pattern will be matched
// without skipping leading whitespace. `name` will be used as
// the terminal's name.
func TokenExact(pattern string, name string) Parser {
	return func(s Scanner) (ParsecNode, Scanner) {
		news := s.Clone()
		cursor := news.GetCursor()
		if tok, _ := news.Match("^" + pattern); tok != nil {
			return NewTerminal(name, string(tok), cursor), news
		}
		return nil, s
	}
}

// Atom is similar to Token, takes a string to match with input
// byte-by-byte. Internally uses the MatchString() API from Scanner.
// Skip leading whitespace. For example:
//		scanner := NewScanner([]byte("cosmos"))
//		Atom("cos", "ATOM")(scanner) // will match
func Atom(match string, name string) Parser {
	return func(s Scanner) (ParsecNode, Scanner) {
		news := s.Clone()
		news.SkipWS()
		cursor := news.GetCursor()
		if ok, _ := news.MatchString(match); ok {
			return NewTerminal(name, match, cursor), news
		}
		return nil, s
	}
}

// AtomExact is similar to Atom(), but string will be matched without
// skipping leading whitespace.
func AtomExact(match string, name string) Parser {
	return func(s Scanner) (ParsecNode, Scanner) {
		news := s.Clone()
		cursor := news.GetCursor()
		if ok, _ := news.MatchString(match); ok {
			return NewTerminal(name, match, cursor), news
		}
		return nil, s
	}
}

// OrdTokens to parse a single token based on one of the
// specified `patterns`. Skip leading whitespaces.
func OrdTokens(patterns []string, names []string) Parser {
	var group string
	groups := make([]string, 0, len(patterns))
	for i, pattern := range patterns {
		if names[i] == "" {
			group = "^(" + pattern + ")"
		} else {
			group = "^(?P<" + names[i] + ">" + pattern + ")"
		}
		groups = append(groups, group)
	}
	ordPattern := strings.Join(groups, "|")
	return func(s Scanner) (ParsecNode, Scanner) {
		news := s.Clone()
		news.SkipWS()
		cursor := news.GetCursor()
		if captures, _ := news.SubmatchAll(ordPattern); captures != nil {
			for name, tok := range captures {
				return NewTerminal(name, string(tok), cursor), news
			}
		}
		return nil, s
	}
}

// End is a parser function to detect end of scanner output, return
// boolean as ParseNode, hence incompatible with AST{}. Instead, use
// AST:End method.
func End() Parser {
	return func(s Scanner) (ParsecNode, Scanner) {
		if s.Endof() {
			return true, s
		}
		return nil, s
	}
}

// NoEnd is a parser function to detect not-an-end of
// scanner output, return boolean as ParsecNode, hence
// incompatible with AST{}.
func NoEnd() Parser {
	return func(s Scanner) (ParsecNode, Scanner) {
		if !s.Endof() {
			return true, s
		}
		return nil, s
	}
}

var escapeCode = [256]byte{ // TODO: size can be optimized
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'\'': '\'',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

func scanString(txt []byte) (tok []byte, readn int) {
	if len(txt) < 2 {
		return nil, 0
	}

	e := 1
	for txt[e] != '"' {
		c := txt[e]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			e++
			continue
		}
		r, size := utf8.DecodeRune(txt[e:])
		if r == utf8.RuneError && size == 1 {
			return nil, 0
		}
		e += size
	}

	if txt[e] == '"' { // done we have nothing to unquote
		return txt[:e+1], e + 1
	}

	out := make([]byte, len(txt)+2*utf8.UTFMax)
	oute := copy(out, txt[:e]) // copy so far

loop:
	for e < len(txt) {
		switch c := txt[e]; {
		case c == '"':
			out[oute] = c
			e++
			break loop

		case c == '\\':
			if txt[e+1] == 'u' {
				r := getu4(txt[e:])
				if r < 0 { // invalid
					return nil, 0
				}
				e += 6
				if utf16.IsSurrogate(r) {
					nextr := getu4(txt[e:])
					dec := utf16.DecodeRune(r, nextr)
					if dec != unicode.ReplacementChar { // A valid pair consume
						oute += utf8.EncodeRune(out[oute:], dec)
						e += 6
						break loop
					}
					// Invalid surrogate; fall back to replacement rune.
					r = unicode.ReplacementChar
				}
				oute += utf8.EncodeRune(out[oute:], r)

			} else { // escaped with " \ / ' b f n r t
				out[oute] = escapeCode[txt[e+1]]
				e += 2
				oute++
			}

		case c < ' ': // control character is invalid
			return nil, 0

		case c < utf8.RuneSelf: // ASCII
			out[oute] = c
			oute++
			e++

		default: // coerce to well-formed UTF-8
			r, size := utf8.DecodeRune(txt[e:])
			e += size
			oute += utf8.EncodeRune(out[oute:], r)
		}
	}

	if out[oute] == '"' {
		return out[:oute+1], e
	}
	return nil, 0
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	r, err := strconv.ParseUint(string(s[2:6]), 16, 64)
	if err != nil {
		return -1
	}
	return rune(r)
}
