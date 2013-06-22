package main

import (
	"bufio"
	"fmt"
	"io"
)

func eval(v interface{}, env *environment) (interface{}, error) {
	if v == nil {
		return &sexp{}, nil
	}

	switch t := v.(type) {
	case symbol:
		return t.eval(env)
	case *sexp:
		return t.eval(env)
	default:
		debugPrint("default eval")
		return v, nil
	}

	panic("not reached")
}

type interpreter struct {
	in     io.Reader        // reader of input source code
	out1   io.Writer        // writer of evaluated values
	out2   io.Writer        // writer of error info
	tokens chan token       // tokens returns from the lexer (internal only)
	values chan interface{} // values returned from the interpreter (internal only)
	errors chan error       // errors returned from the interpreter (internal only)
}

func newInterpreter(in io.Reader, out1, out2 io.Writer) *interpreter {
	return &interpreter{
		in:     in,
		out1:   out1,
		out2:   out2,
		tokens: make(chan token),
		values: make(chan interface{}),
		errors: make(chan error),
	}
}

func (i interpreter) run(env *environment) {
	go lex(bufio.NewReader(i.in), i.tokens)
	go i.send()
	for {
		v, err := parse(i.tokens)
		switch err {
		case io.EOF:
			return
		case nil:
			i.eval(v, env)
		default:
			i.errors <- err
		}
	}
}

func (i interpreter) eval(v interface{}, env *environment) {
	val, err := eval(v, env)
	if err != nil {
		i.errors <- err
		return
	}
	i.values <- val
}

func (i interpreter) send() {
	for {
		select {
		case v := <-i.values:
			if _, err := fmt.Fprintln(i.out1, v); err != nil {
				fmt.Println("can't write out to client: ", err)
			}
		case e := <-i.errors:
			if _, err := fmt.Fprintln(i.out2, e); err != nil {
				fmt.Println("can't write error to client: ", err)
			}
		}
	}
}
