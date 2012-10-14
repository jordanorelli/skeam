package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type typ3 int

const (
	invalid typ3 = iota
	int3ger
	symbol
	openParen
	closeParen
	str1ng
	fl0at
)

func (t typ3) String() string {
	switch t {
	case int3ger:
		return "integer"
	case symbol:
		return "symbol"
	case openParen:
		return "open_paren"
	case closeParen:
		return "close_paren"
	case str1ng:
		return "string"
	case fl0at:
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
	input *bufio.Reader
	cur   []rune
	depth int
	out   chan token
}

func (l *lexer) next() (rune, error) {
	r, _, err := l.input.ReadRune()
	return r, err
}

// clears the current lexem buffer and emits a token of the given type.
// There's no sanity checking to make sure you don't emit some bullshit, so
// don't fuck it up.
func (l *lexer) emit(t typ3) {
	l.out <- token{lexeme: string(l.cur), t: t}
	l.cur = nil
}

// appends the rune to the current in-progress lexem
func (l *lexer) append(r rune) {
	if l.cur == nil {
        l.cur = make([]rune, 0, 32)
	}
	l.cur = append(l.cur, r)
}

// lexes stuff at the root level of the input.
func lexRoot(l *lexer) (stateFn, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	switch r {
	case '(':
		return lexOpenParen, nil
	case ' ', '\t', '\n':
		return lexRoot, nil
	}
	return nil, fmt.Errorf("unexpected rune in lexRoot: %c", r)
}

func isDigit(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

// lexes an open parenthesis
func lexOpenParen(l *lexer) (stateFn, error) {
	l.out <- token{"(", openParen}
	l.depth++
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n':
		return lexRoot, nil
	case '(':
		return nil, fmt.Errorf("the whole (( thing isn't supported yet")
	}
	if isDigit(r) {
		l.append(r)
		return lexInt, nil
	}
	l.append(r)
	return lexSymbol, nil
}

func lexInt(l *lexer) (stateFn, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n':
		l.emit(int3ger)
		return lexWhitespace, nil
	case '.':
		l.append(r)
		return lexFloat, nil
	case ')':
		l.emit(int3ger)
		return lexCloseParen, nil
	}
	if isDigit(r) {
		l.append(r)
		return lexInt, nil
	}
	return nil, fmt.Errorf("unexpected rune in lexInt: %c", r)
}

// once we're in a float, the only valid values are digits, whitespace or close
// paren.
func lexFloat(l *lexer) (stateFn, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}

	switch r {
	case ' ', '\t', '\n':
		l.emit(fl0at)
		return lexWhitespace, nil
	case ')':
		l.emit(fl0at)
		return lexCloseParen, nil
	}
	if isDigit(r) {
		l.append(r)
		return lexFloat, nil
	}
	return nil, fmt.Errorf("unexpected run in lexFloat: %c", r)
}

// lexes a symbol in progress
func lexSymbol(l *lexer) (stateFn, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n':
		l.emit(symbol)
		return lexWhitespace, nil
	case ')':
		l.emit(symbol)
		return lexCloseParen, nil
	default:
		l.append(r)
		return lexSymbol, nil
	}
	panic("not reached")
}

// lexes some whitespace in progress.  Maybe this should be combined with root
// and the lexer shouldn't have a state.  I think wehat I'm doing now is
// "wrong" but who honestly gives a shit.
func lexWhitespace(l *lexer) (stateFn, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n':
		return lexWhitespace, nil
	case '(':
		return lexOpenParen, nil
	}
	if isDigit(r) {
		l.append(r)
		return lexInt, nil
	}
	l.append(r)
	return lexSymbol, nil
}

// lex a close parenthesis
func lexCloseParen(l *lexer) (stateFn, error) {
	l.out <- token{")", closeParen}
	l.depth--
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n':
		if l.depth == 0 {
			return lexRoot, nil
		} else {
			return lexWhitespace, nil
		}
	case ')':
		return lexCloseParen, nil
	}
	return nil, fmt.Errorf("unimplemented")
}

// lexes some lispy input from an io.Reader, emiting tokens on chan c.  The
// channel is closed when the input reaches EOF, signaling that there are no
// new tokens.
func lex(input io.Reader, c chan token) {
	defer close(c)
	l := &lexer{
		input: bufio.NewReader(input),
		out:   c,
	}

	var err error
	f := stateFn(lexRoot)
	for err == nil {
		f, err = f(l)
	}
	if err != io.EOF {
		fmt.Println(err)
	}
	if l.depth != 0 {
		fmt.Println("error: unbalanced parenthesis")
	}
}

func main() {
	filename := "input.lisp"

	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to read file ", filename)
		os.Exit(1)
	}

	c := make(chan token)
	go lex(f, c)

	for s := range c {
		fmt.Println(s.t, s.lexeme)
	}
}
