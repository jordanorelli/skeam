package main

import (
	"fmt"
	"io"
	"strings"
)

type typ3 int

const (
	invalid typ3 = iota
	integerToken
	symbolToken
	openParenToken
	closeParenToken
	stringToken
	floatToken
)

func (t typ3) String() string {
	switch t {
	case integerToken:
		return "integer"
	case symbolToken:
		return "symbol"
	case openParenToken:
		return "open_paren"
	case closeParenToken:
		return "close_paren"
	case stringToken:
		return "string"
	case floatToken:
		return "float"
	}
	panic("wtf")
}

type token struct {
	lexeme string
	t      typ3
}

type stateFn func(*lexer) (stateFn, error)

type lexer struct {
	io.RuneReader
	buf []rune
	cur rune
	out chan token
}

// clears the current lexem buffer and emits a token of the given type.
// There's no sanity checking to make sure you don't emit some bullshit, so
// don't fuck it up.
func (l *lexer) emit(t typ3) {
	debugPrint("emit " + string(l.buf))
	l.out <- token{lexeme: string(l.buf), t: t}
	l.buf = nil
}

// reads a rune from the input and assigns it to the current rune, l.cur.
// Returns an error if we were unable to read a rune from the input.  I'm
// pretty sure it's always io.EOF but I'm not positive.
func (l *lexer) next() error {
	r, _, err := l.ReadRune()
	if err != nil {
		return err
	}
	l.cur = r
	return nil
}

// stores the current rune in our in-progress lexeme buffer
func (l *lexer) keep() {
	if l.buf == nil {
		l.buf = make([]rune, 0, 32)
	}
	l.buf = append(l.buf, l.cur)
}

func isDigit(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

func debugPrint(s string) {
	if DEBUG {
		fmt.Println("#", s)
	}
}

// lexes an open parenthesis
func lexOpenParen(l *lexer) (stateFn, error) {
	debugPrint("-->lexOpenParen")
	l.out <- token{"(", openParenToken}
	switch l.cur {
	case ' ', '\t', '\n', '\r':
		return lexWhitespace, nil
	case '(':
		return lexOpenParen, nil
	case ')':
		return lexCloseParen, nil
	case ';':
		return lexComment, nil
	case '.':
		l.keep()
		return lexFloat, nil
	case '-':
		l.keep()
		return lexMinus, nil
	}
	if isDigit(l.cur) {
		l.keep()
		return lexInt, nil
	}
	l.keep()
	return lexSymbol, nil
}

func lexMinus(l *lexer) (stateFn, error) {
	switch l.cur {
	case ' ', '\t', '\n', '\r':
		l.emit(symbolToken)
		return lexWhitespace, nil
	case '.':
		l.keep()
		return lexFloat, nil
	}
	if isDigit(l.cur) {
		l.keep()
		return lexInt, nil
	}
	l.keep()
	return lexSymbol, nil
}

// lexes some whitespace in progress.  Maybe this should be combined with root
// and the lexer shouldn't have a state.  I think wehat I'm doing now is
// "wrong" but who honestly gives a shit.
func lexWhitespace(l *lexer) (stateFn, error) {
	debugPrint("-->lexWhitespace")
	switch l.cur {
	case ' ', '\t', '\n', '\r':
		return lexWhitespace, nil
	case '"':
		return lexString, nil
	case '(':
		return lexOpenParen, nil
	case ')':
		return lexCloseParen, nil
	case ';':
		return lexComment, nil
	case '.':
		l.keep()
		return lexFloat, nil
	case '-':
		l.keep()
		return lexMinus, nil
	}
	if isDigit(l.cur) {
		l.keep()
		return lexInt, nil
	}
	l.keep()
	return lexSymbol, nil
}

// lexes an in-progress string.  Basically we just keep all of the tokens until
// we see a double-quote character, signifying the end of the string.  We also
// switch into escape mode if we come across a backslash.
func lexString(l *lexer) (stateFn, error) {
	debugPrint("-->lexString")
	switch l.cur {
	case '"':
		l.emit(stringToken)
		return lexWhitespace, nil
	case '\\':
		return lexStringEsc, nil
	}
	l.keep()
	return lexString, nil
}

// lex the character *after* the string escape character \.  We always keep the
// next character, then just go back to string lexing.
func lexStringEsc(l *lexer) (stateFn, error) {
	debugPrint("-->lexStringEsc")
	l.keep()
	return lexString, nil
}

// lex an integer.  Once we're on an integer, the only valid characters are
// whitespace, close paren, a period to indicate we want a float, or more
// digits.  Everything else is crap.
func lexInt(l *lexer) (stateFn, error) {
	debugPrint("-->lexInt")
	switch l.cur {
	case ' ', '\t', '\n', '\r':
		l.emit(integerToken)
		return lexWhitespace, nil
	case '.':
		l.keep()
		return lexFloat, nil
	case ')':
		l.emit(integerToken)
		return lexCloseParen, nil
	case ';':
		l.emit(integerToken)
		return lexComment, nil
	}
	if isDigit(l.cur) {
		l.keep()
		return lexInt, nil
	}
	return nil, fmt.Errorf("unexpected rune in lexInt: %c", l.cur)
}

// once we're in a float, the only valid values are digits, whitespace or close
// paren.
func lexFloat(l *lexer) (stateFn, error) {
	debugPrint("-->lexFloat")
	switch l.cur {
	case ' ', '\t', '\n', '\r':
		l.emit(floatToken)
		return lexWhitespace, nil
	case ')':
		l.emit(floatToken)
		return lexCloseParen, nil
	case ';':
		l.emit(floatToken)
		return lexComment, nil
	}
	if isDigit(l.cur) {
		l.keep()
		return lexFloat, nil
	}
	return nil, fmt.Errorf("unexpected rune in lexFloat: %c", l.cur)
}

// lexes a symbol in progress
func lexSymbol(l *lexer) (stateFn, error) {
	debugPrint("-->lexSymbol")
	switch l.cur {
	case ' ', '\t', '\n', '\r':
		debugPrint("ending lexSymbol on whitespace")
		l.emit(symbolToken)
		return lexWhitespace, nil
	case ')':
		l.emit(symbolToken)
		return lexCloseParen, nil
	case ';':
		l.emit(symbolToken)
		return lexComment, nil
	default:
		l.keep()
		return lexSymbol, nil
	}
	panic("not reached")
}

// lex a close parenthesis
func lexCloseParen(l *lexer) (stateFn, error) {
	debugPrint("-->lexCloseParen")
	l.out <- token{")", closeParenToken}
	switch l.cur {
	case ' ', '\t', '\n', '\r':
		return lexWhitespace, nil
	case ')':
		return lexCloseParen, nil
	case ';':
		return lexComment, nil
	}
	return nil, fmt.Errorf("unimplemented")
}

// lexes a comment
func lexComment(l *lexer) (stateFn, error) {
	debugPrint("-->lexComment")
	switch l.cur {
	case '\n', '\r':
		return lexWhitespace, nil
	}
	return lexComment, nil
}

// lexes some lispy input from an io.Reader, emiting tokens on chan c.  The
// channel is closed when the input reaches EOF, signaling that there are no
// new tokens.
func lex(input io.RuneReader, c chan token) {
	defer close(c)
	l := &lexer{input, nil, ' ', c}

	var err error
	f := stateFn(lexWhitespace)
	for {
		f, err = f(l)
		if err != nil {
			break
		}
		err = l.next()
		if err != nil {
			break
		}
	}
	if err != io.EOF {
		fmt.Println(err)
	}
}

// lexes a lispy string onto a token channel
func lexs(input string, c chan token) {
	lex(strings.NewReader(input), c)
}
