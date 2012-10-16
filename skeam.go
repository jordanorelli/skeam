package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

var DEBUG = false

type sexp []interface{}

type symbol string

// parses the string lexeme into a value that can be eval'd
func atom(t token) (interface{}, error) {
	switch t.t {
	case integerToken:
		val, err := strconv.ParseInt(t.lexeme, 10, 64)
		if err != nil {
			return nil, err
		}
		return val, nil

	case floatToken:
		val, err := strconv.ParseFloat(t.lexeme, 64)
		if err != nil {
			return nil, err
		}
		return val, nil

	case stringToken:
		return t.lexeme, nil

	case symbolToken:
		return symbol(t.lexeme), nil
	}

	return nil, fmt.Errorf("unable to atomize token: %v", t)
}

// reads in tokens on the channel until a matching close paren is found.
func (s *sexp) readIn(c chan token) error {
	for t := range c {
		switch t.t {
		case closeParenToken:
			return nil
		case openParenToken:
			child := make(sexp, 0)
			if err := child.readIn(c); err != nil {
				return err
			}
			*s = append(*s, child)
		default:
			v, err := atom(t)
			if err != nil {
				return err
			}
			*s = append(*s, v)
		}
	}
	return errors.New("unexpected EOF in sexp.readIn")
}

// parses one value that can be evaled from the channel
func parse(c chan token) (interface{}, error) {
	for t := range c {
		switch t.t {
		case closeParenToken:
			return nil, errors.New("unexpected EOF in read")
		case openParenToken:
			s := make(sexp, 0)
			if err := s.readIn(c); err != nil {
				return nil, err
			}
			return s, nil
		default:
			return atom(t)
		}
	}
	return nil, io.EOF
}

func eval(v interface{}) {
	fmt.Println(v)
}

func evalall(c chan token) {
	for {
		v, err := parse(c)
		switch err {
		case io.EOF:
			return
		case nil:
			eval(v)
		default:
			fmt.Println("error in eval: %v", err)
		}
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

	c := make(chan token, 32)
	go lex(bufio.NewReader(f), c)
	evalall(c)
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

		c := make(chan token, 32)
		go lexs(string(line)+"\n", c)
		evalall(c)
	}
}
