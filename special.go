package main

import (
	"errors"
	"fmt"
	"reflect"
)

type special func(*environment, ...interface{}) (interface{}, error)

type nargsInvalidError struct {
	expected int
	received int
	name     string
}

func (n nargsInvalidError) Error() string {
	return fmt.Sprintf(`received %d arguments in *%v*, expected %d`,
		n.received, n.name, n.expected)
}

// defines the built-in "define" construct.  e.g.:
//
//  (define x 5)
//
// would create the symbol "x" and set its value to 5.
func define(env *environment, args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, nargsInvalidError{2, len(args), "define"}
	}

	s, ok := args[0].(symbol)
	if !ok {
		return nil, fmt.Errorf(`first argument to *define* must be symbol, received %v`, reflect.TypeOf(args[0]))
	}

	v, err := eval(args[1], env)
	if err != nil {
		return nil, err
	}
	env.set(s, v)

	return nil, nil
}

// defines the built-in "quote" construct.  e.g.:
//
//  (quote (1 2 3))
//
// would evaluate to the list (1 2 3).  That is, quote is a function of arity 1
// that is effectively a no-op; the input value is not evaluated, which
// prevents evaluation of the first element of the list, in this case 1.
func quote(_ *environment, args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, nargsInvalidError{1, len(args), "quote"}
	}

	return args[0], nil
}

// defines the built-in "if" contruct.  e.g.:
//
//  (if #t "foo" "bar")
//
// would evaluate to "foo", while the following:
//
//  (if #f "foo" "bar")
//
// would evaluate to "bar"
func _if(env *environment, args ...interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, nargsInvalidError{3, len(args), "if"}
	}

	v, err := eval(args[0], env)
	if err != nil {
		return nil, err
	}

	if b, ok := v.(bool); ok && !b {
		return eval(args[2], env)
	}
	return eval(args[1], env)
}

// defines the built-in "set" construct, which is used to set the value of an
// existing symbol in the provided environment.  e.g.:
//
//  (set! x 5)
//
// would set the symbol x to the value 5, if and only if the symbol x was
// previously defined.
func set(env *environment, args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, nargsInvalidError{2, len(args), "set!"}
	}

	s, ok := args[0].(symbol)
	if !ok {
		return nil, fmt.Errorf(`first argument to *set!* must be symbol, received %v`, reflect.TypeOf(args[0]))
	}

	if !env.defined(s) {
		return nil, fmt.Errorf(`cannot *set!* undefined symbol %v`, s)
	}

	v, err := eval(args[1], env)
	if err != nil {
		return nil, err
	}
	env.set(s, v)

	return nil, nil
}

type lambda struct {
	env       *environment
	arglabels []symbol
	body      sexp
}

func (l lambda) call(env *environment, rawArgs []interface{}) (interface{}, error) {
	debugPrint("call lambda")

	args := make([]interface{}, 0, len(rawArgs))
	for _, raw := range rawArgs {
		v, err := eval(raw, env)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}

	if len(args) != len(l.arglabels) {
		return nil, errors.New("parity error")
	}

	for i := range args {
		l.env.set(l.arglabels[i], args[i])
	}

	return eval(l.body, l.env)
}

// defines the built-in lambda construct.  e.g.:
//
//  (lambda (x) (* x x))
//
// would evaluate to a lambda that, when executed, squares its input.
func mklambda(env *environment, args ...interface{}) (interface{}, error) {
	debugPrint("mklambda")
	if len(args) != 2 {
		return nil, nargsInvalidError{2, len(args), "lambda"}
	}

	params, ok := args[0].(sexp)
	if !ok {
		return nil, fmt.Errorf(`first argument to *lambda* must be sexp, received %v`, reflect.TypeOf(args[0]))
	}

	arglabels := make([]symbol, 0, len(params))
	for _, v := range params {
		s, ok := v.(symbol)
		if !ok {
			return nil, fmt.Errorf(`lambda args must all be symbols; received invalid %v`, reflect.TypeOf(v))
		}
		arglabels = append(arglabels, s)
	}

	body, ok := args[1].(sexp)
	if !ok {
		return nil, fmt.Errorf(`second argument to *lambda* must be sexp, received %v`, reflect.TypeOf(args[1]))
	}

	return lambda{env, arglabels, body}, nil
}
