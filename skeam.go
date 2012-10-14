package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type stateFn func(*lexer) (stateFn, error)

type lexer struct {
	input *bufio.Reader
	cur   []rune
	depth int
	out   chan string
}

func (l *lexer) next() (rune, error) {
	r, _, err := l.input.ReadRune()
	return r, err
}

func (l *lexer) emit() {
	l.out <- string(l.cur)
	l.cur = nil
}

func (l *lexer) append(r rune) {
	if l.cur == nil {
		l.cur = []rune{r}
		return
	}
	l.cur = append(l.cur, r)
}

func lexRoot(l *lexer) (stateFn, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	switch r {
	case '(':
		l.append(r)
		l.emit()
		return lexOpenParen, nil
	case ' ', '\t', '\n':
		return lexRoot, nil
	}
	return nil, fmt.Errorf("unexpected rune in lexRoot: %c", r)
}

func lexOpenParen(l *lexer) (stateFn, error) {
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
	default:
		l.append(r)
		return lexOnSymbol, nil
	}
	panic("not reached")
}

func lexOnSymbol(l *lexer) (stateFn, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n':
		l.emit()
		return lexWhitespace, nil
	case ')':
		l.emit()
		l.append(r)
		l.emit()
		return lexCloseParen, nil
	default:
		l.append(r)
		return lexOnSymbol, nil
	}
	panic("not reached")
}

func lexWhitespace(l *lexer) (stateFn, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	switch r {
	case ' ', '\t', '\n':
		return lexWhitespace, nil
	case '(':
		l.append(r)
		l.emit()
		return lexOpenParen, nil
	default:
		l.append(r)
		return lexOnSymbol, nil
	}
	panic("not reached")
}

func lexCloseParen(l *lexer) (stateFn, error) {
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
		l.append(r)
		l.emit()
		return lexCloseParen, nil
	}
	return nil, fmt.Errorf("unimplemented")
}

func lex(input io.Reader, c chan string) {
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

	c := make(chan string)
	go lex(f, c)

	for s := range c {
		fmt.Println(s)
	}
}
