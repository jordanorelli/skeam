package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var DEBUG = false

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
	io.RuneReader
	cur   []rune
	depth int
	out   chan token
}

// clears the current lexem buffer and emits a token of the given type.
// There's no sanity checking to make sure you don't emit some bullshit, so
// don't fuck it up.
func (l *lexer) emit(t typ3) {
	debugPrint("emit " + string(l.cur))
	l.out <- token{lexeme: string(l.cur), t: t}
	l.cur = nil
}

// appends the rune to the current in-progress lexem
func (l *lexer) append(r rune) {
	debugPrint(fmt.Sprintf("append %c\n", (r)))
	if l.cur == nil {
		l.cur = make([]rune, 0, 32)
	}
	l.cur = append(l.cur, r)
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
		fmt.Println(s)
	}
}

// lexes an open parenthesis
func lexOpenParen(l *lexer) (stateFn, error) {
	debugPrint("-->lexOpenParen")
	l.out <- token{"(", openParen}
	l.depth++
	r, _, err := l.ReadRune()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n', '\r':
		return lexWhitespace, nil
	case '(':
		return lexOpenParen, nil
	case ';':
		return lexComment, nil
	}
	if isDigit(r) {
		l.append(r)
		return lexInt, nil
	}
	l.append(r)
	return lexSymbol, nil
}

// lexes some whitespace in progress.  Maybe this should be combined with root
// and the lexer shouldn't have a state.  I think wehat I'm doing now is
// "wrong" but who honestly gives a shit.
func lexWhitespace(l *lexer) (stateFn, error) {
	debugPrint("-->lexWhitespace")
	r, _, err := l.ReadRune()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n', '\r':
		return lexWhitespace, nil
	case '"':
		return lexString, nil
	case '(':
		return lexOpenParen, nil
	case ';':
		return lexComment, nil
	}
	if isDigit(r) {
		l.append(r)
		return lexInt, nil
	}
	l.append(r)
	return lexSymbol, nil
}

func lexString(l *lexer) (stateFn, error) {
	debugPrint("-->lexString")
	r, _, err := l.ReadRune()
	if err != nil {
		return nil, err
	}
	switch r {
	case '"':
		l.emit(str1ng)
		return lexWhitespace, nil
	case '\\':
		return lexStringEsc, nil
	}
	l.append(r)
	return lexString, nil
}

// lex the character *after* the string escape character \
func lexStringEsc(l *lexer) (stateFn, error) {
	debugPrint("-->lexStringEsc")
	r, _, err := l.ReadRune()
	if err != nil {
		return nil, err
	}
	l.append(r)
	return lexString, nil
}

// lex an integer.  Once we're on an integer, the only valid characters are
// whitespace, close paren, a period to indicate we want a float, or more
// digits.  Everything else is crap.
func lexInt(l *lexer) (stateFn, error) {
	debugPrint("-->lexInt")
	r, _, err := l.ReadRune()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n', '\r':
		l.emit(int3ger)
		return lexWhitespace, nil
	case '.':
		l.append(r)
		return lexFloat, nil
	case ')':
		l.emit(int3ger)
		return lexCloseParen, nil
	case ';':
		l.emit(int3ger)
		return lexComment, nil
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
	debugPrint("-->lexFloat")
	r, _, err := l.ReadRune()
	if err != nil {
		return nil, err
	}

	switch r {
	case ' ', '\t', '\n', '\r':
		l.emit(fl0at)
		return lexWhitespace, nil
	case ')':
		l.emit(fl0at)
		return lexCloseParen, nil
	case ';':
		l.emit(fl0at)
		return lexComment, nil
	}
	if isDigit(r) {
		l.append(r)
		return lexFloat, nil
	}
	return nil, fmt.Errorf("unexpected run in lexFloat: %c", r)
}

// lexes a symbol in progress
func lexSymbol(l *lexer) (stateFn, error) {
	debugPrint("-->lexSymbol")
	r, _, err := l.ReadRune()
	if err != nil {
		return nil, err
	}

	switch r {
	case ' ', '\t', '\n', '\r':
		debugPrint("ending lexSymbol on whitespace")
		l.emit(symbol)
		return lexWhitespace, nil
	case ')':
		l.emit(symbol)
		return lexCloseParen, nil
	case ';':
		l.emit(symbol)
		return lexComment, nil
	default:
		l.append(r)
		return lexSymbol, nil
	}
	panic("not reached")
}

// lex a close parenthesis
func lexCloseParen(l *lexer) (stateFn, error) {
	debugPrint("-->lexCloseParen")
	l.out <- token{")", closeParen}
	l.depth--
	r, _, err := l.ReadRune()
	if err != nil {
		return nil, err
	}
	switch r {
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
	r, _, err := l.ReadRune()
	if err != nil {
		return nil, err
	}
	switch r {
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
	l := &lexer{input, nil, 0, c}

	var err error
	f := stateFn(lexWhitespace)
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

func args() {
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to read file ", filename)
		os.Exit(1)
	}
	defer f.Close()

	c := make(chan token)
	go lex(bufio.NewReader(f), c)

	for s := range c {
		fmt.Printf("%11s %s\n", s.t, s.lexeme)
	}
}

func lexs(input string, c chan token) {
    lex(strings.NewReader(input), c)
}

func main() {
	if len(os.Args) > 1 {
		args()
		return
	}

	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, prefix, err := r.ReadLine()
		if prefix {
			fmt.Println("(prefix)")
		}
        switch err {
        case nil:
            break
        case io.EOF:
            fmt.Print("\n")
            return
        default:
			fmt.Println("error: ", err)
            continue
        }

        c := make(chan token)
		go lexs(string(line) + "\n", c)
        for s := range c {
            fmt.Printf("%11s %s\n", s.t, s.lexeme)
        }
	}
}
