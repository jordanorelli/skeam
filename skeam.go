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

func (s sexp) String() string {
	return "(" + fmt.Sprint(s...) + ")"
}

type symbol string

var universe = &environment{map[symbol]interface{}{
	"#t":     true,
	"#f":     false,
	"+":      builtin(addition),
	"-":      builtin(subtraction),
	"*":      builtin(multiplication),
	"/":      builtin(division),
	"length": builtin(length),
	"list":   builtin(list),
	"list?":  builtin(islist),
	"not":    builtin(not),
	"begin":  special(begin),
	"define": special(define),
	"if":     special(_if),
	"lambda": special(mklambda),
	"quote":  special(quote),
	"set!":   special(set),
}, nil}

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

func eval(v interface{}, env *environment) (interface{}, error) {
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

		// eval the first item
		v, err := eval(t[0], env)
		if err != nil {
			return nil, err
		}

		// check to see if this is a special form
		if spec, ok := v.(special); ok {
			if len(t) > 1 {
				return spec(env, t[1:]...)
			} else {
				return spec(env)
			}
		}

		// exec builtin func if one exists
		if b, ok := v.(builtin); ok {
			if len(t) > 1 {
				return b.call(env, t[1:])
			} else {
				return b.call(env, nil)
			}
		}

		// exec lambda if possible
		if l, ok := v.(lambda); ok {
			if len(t) > 1 {
				return l.call(env, t[1:])
			} else {
				return l.call(env, nil)
			}
		}

		return nil, fmt.Errorf(`expected special form or builtin procedure, received %v`, reflect.TypeOf(v))

	default:
		return v, nil
	}
	return nil, nil
}

func evalall(c chan token, env *environment) {
	for {
		v, err := parse(c)
		switch err {
		case io.EOF:
			return
		case nil:
			if v, err := eval(v, env); err != nil {
				fmt.Println("error:", err)
				return
			} else {
				if v != nil {
					fmt.Println(v)
				}
			}
		default:
			fmt.Printf("error in eval: %v\n", err)
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
