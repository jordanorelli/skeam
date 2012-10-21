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
	buf   []rune
	cur   rune
	depth int
	out   chan token
}

// clears the current lexem buffer and emits a token of the given type.
// There's no sanity checking to make sure you don't emit some bullshit, so
// don't fuck it up.
func (l *lexer) emit(t typ3) {
	debugPrint("emit " + string(l.buf))
	l.out <- token{lexeme: string(l.buf), t: t}
	l.buf = nil
}

func (l *lexer) next() error {
	r, _, err := l.ReadRune()
	if err != nil {
		return err
	}
	l.cur = r
	return nil
}

// appends the rune to the current in-progress lexem
func (l *lexer) append(r rune) {
	debugPrint(fmt.Sprintf("append %c\n", (r)))
	if l.buf == nil {
		l.buf = make([]rune, 0, 32)
	}
	l.buf = append(l.buf, r)
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
	l.depth++
	switch l.cur {
	case ' ', '\t', '\n', '\r':
		return lexWhitespace, nil
	case '(':
		return lexOpenParen, nil
	case ')':
		return lexCloseParen, nil
	case ';':
		return lexComment, nil
	}
	if isDigit(l.cur) {
		l.append(l.cur)
		return lexInt, nil
	}
	l.append(l.cur)
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
	}
	if isDigit(l.cur) {
		l.append(l.cur)
		return lexInt, nil
	}
	l.append(l.cur)
	return lexSymbol, nil
}

func lexString(l *lexer) (stateFn, error) {
	debugPrint("-->lexString")
	switch l.cur {
	case '"':
		l.emit(stringToken)
		return lexWhitespace, nil
	case '\\':
		return lexStringEsc, nil
	}
	l.append(l.cur)
	return lexString, nil
}

// lex the character *after* the string escape character \
func lexStringEsc(l *lexer) (stateFn, error) {
	debugPrint("-->lexStringEsc")
	l.append(l.cur)
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
		l.append(l.cur)
		return lexFloat, nil
	case ')':
		l.emit(integerToken)
		return lexCloseParen, nil
	case ';':
		l.emit(integerToken)
		return lexComment, nil
	}
	if isDigit(l.cur) {
		l.append(l.cur)
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
		l.append(l.cur)
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
		l.append(l.cur)
		return lexSymbol, nil
	}
	panic("not reached")
}

// lex a close parenthesis
func lexCloseParen(l *lexer) (stateFn, error) {
	debugPrint("-->lexCloseParen")
	l.out <- token{")", closeParenToken}
	l.depth--
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
	l := &lexer{input, nil, ' ', 0, c}

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
	if l.depth != 0 {
		fmt.Println("error: unbalanced parenthesis")
	}
}
func lexs(input string, c chan token) {
	lex(strings.NewReader(input), c)
}
