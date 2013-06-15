package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var DEBUG = false

type sexp struct {
	items    []interface{}
	quotelvl int
}

type callable interface {
	call(*environment, []interface{}) (interface{}, error)
}

func newSexp() *sexp {
	return &sexp{
		items:    make([]interface{}, 0, 8),
		quotelvl: 0,
	}
}

func (s sexp) String() string {
	parts := make([]string, len(s.items))
	for i, _ := range s.items {
		parts[i] = fmt.Sprint(s.items[i])
	}
	return "(" + strings.Join(parts, " ") + ")"
}

func (s *sexp) append(item interface{}) {
	s.items = append(s.items, item)
}

func (s sexp) len() int {
	return len(s.items)
}

type symbol string

var universe = &environment{map[symbol]interface{}{
	// predefined values
	"#t":   true,
	"#f":   false,
	"null": nil,

	// builtin functions
	symbol(add.name):      add,
	symbol(sub.name):      sub,
	symbol(mul.name):      mul,
	symbol(div.name):      div,
	symbol(gt.name):       gt,
	symbol(gte.name):      gte,
	symbol(lt.name):       lt,
	symbol(lte.name):      lte,
	symbol(cons.name):     cons,
	symbol(car.name):      car,
	symbol(cdr.name):      cdr,
	symbol(length.name):   length,
	symbol(lst.name):      lst,
	symbol(islist.name):   islist,
	symbol(not.name):      not,
	symbol(isnull.name):   isnull,
	symbol(issymbol.name): issymbol,
	// "=":       builtin(equal),
	// "equal?":  builtin(equal),
	// "eq?"
	// "append"

	// special forms
	symbol(begin.name):    begin,
	symbol(define.name):   define,
	symbol(_if.name):      _if,
	symbol(mklambda.name): mklambda,
	symbol(quote.name):    quote,
	symbol(set.name):      set,
}, nil}

func init() {
	universe.set(symbol(names.name), names)
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
			child := newSexp()
			if err := child.readIn(c); err != nil {
				return err
			}
			s.append(child)
		default:
			v, err := atom(t)
			if err != nil {
				return err
			}
			s.append(v)
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
			s := newSexp()
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
	if v == nil {
		return &sexp{}, nil
	}

	switch t := v.(type) {

	case symbol:
		debugPrint("eval symbol")
		s, err := env.get(t)
		if err != nil {
			return nil, err
		}
		return eval(s, env)

	case *sexp:
		debugPrint("eval sexp")
		if t.len() == 0 {
			return nil, errors.New("illegal evaluation of empty sexp ()")
		}

		if t.quotelvl > 0 {
			return t, nil
		}

		// eval the first item
		v, err := eval(t.items[0], env)
		if err != nil {
			return nil, err
		}

		c, ok := v.(callable)
		if !ok {
			return nil, fmt.Errorf(`expected special form or builtin procedure, received %v`, reflect.TypeOf(v))
		}
		if len(t.items) > 1 {
			return c.call(env, t.items[1:])
		}
		return c.call(env, nil)

	default:
		debugPrint("default eval")
		return v, nil
	}

	panic("not reached")
}

func evalall(c chan token, out chan interface{}, e chan error, env *environment) {
	for {
		v, err := parse(c)
		switch err {
		case io.EOF:
			return
		case nil:
			if v, err := eval(v, env); err != nil {
				e <- err
				return
			} else {
				out <- v
			}
		default:
			e <- err
		}
	}
}

func defaultInterpreter(out chan interface{}, errors chan error) {
	for {
		select {
		case v := <-out:
			fmt.Println(v)
		case err := <-errors:
			fmt.Printf("error: %v", err)
		}
	}
}

func main() {
	if DEBUG {
		fmt.Println(universe)
	}
	if len(os.Args) > 1 {
		args()
		return
	}

	out, errors := make(chan interface{}), make(chan error)
	go defaultInterpreter(out, errors)

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
		evalall(c, out, errors, universe)
	}
}
