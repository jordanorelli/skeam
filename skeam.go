package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

var DEBUG = false

type sexp []interface{}

type symbol string

var universe = environment{
	"int":    int64(5),
	"float":  float64(3.14),
	"string": "Jordan",
	"+":      proc(addition),
	"-":      proc(subtraction),
	"*":      proc(multiplication),
}

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

func eval(v interface{}, env environment) (interface{}, error) {
	switch t := v.(type) {

	case symbol:
		debugPrint("eval symbol")
		s, err := env.get(t)
		if err != nil {
			return nil, err
		}
		return eval(s, env)

	case sexp:
		debugPrint("eval sexp")
		if len(t) == 0 {
			return nil, errors.New("illegal evaluation of empty sexp ()")
		}

		s, ok := t[0].(symbol)
		if !ok {
			return nil, errors.New("expected a symbol")
		}

		v, err := env.get(s)
		if err != nil {
			return nil, err
		}

		p, ok := v.(proc)
		if !ok {
			return nil, fmt.Errorf("expected proc, found %v", reflect.TypeOf(v))
		}

		if len(t) > 1 {
			args := make([]interface{}, 0, len(t)-1)
			for _, raw := range t[1:] {
				v, err := eval(raw, env)
				if err != nil {
					return nil, err
				}
				args = append(args, v)
			}
			inner, err := p(args...)
			if err != nil {
				return nil, err
			}
			return eval(inner, env)
		}

		inner, err := p()
		if err != nil {
			return nil, err
		}

		return eval(inner, env)

	default:
		return v, nil
	}
	return nil, nil
}

func evalall(c chan token, env environment) {
	for {
		v, err := parse(c)
		switch err {
		case io.EOF:
			return
		case nil:
			if v, err := eval(v, env); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println(v)
			}
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
	evalall(c, universe)
}

func main() {
	if DEBUG {
		fmt.Println(universe)
	}
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
		evalall(c, universe)
	}
}
