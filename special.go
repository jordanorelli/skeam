package main

import (
	"errors"
	"fmt"
	"reflect"
)

// type special is a callable outside of the normal execution workflow.  That
// is, a special receives its arguments unevaluated, unlike lambdas or builtin,
// both of whose arguments are evaluated upon invocation.
// type special func(*environment, ...interface{}) (interface{}, error)
type special struct {
	name     string
	arity    int
	variadic bool
	fn       func(*environment, []interface{}) (interface{}, error)
}

func (s special) checkArity(n int) error {
	if n == s.arity {
		return nil
	}

	if s.variadic && n > s.arity {
		return nil
	}

	return arityError{
		expected: s.arity,
		received: n,
		name:     s.name,
		variadic: s.variadic,
	}
}

func (s special) call(env *environment, rawArgs []interface{}) (interface{}, error) {
	if err := s.checkArity(len(rawArgs)); err != nil {
		return nil, err
	}
	return s.fn(env, rawArgs)
}

// type arityError is used to store information related to arity errors.  That
// is, the invocation of a callable with the wrong number of arguments.
type arityError struct {
	expected int
	received int
	name     string
	variadic bool
}

func (n arityError) Error() string {
	if n.variadic {
		return fmt.Sprintf(`received %d arguments in *%v*, expected %d (or more)`,
			n.received, n.name, n.expected)
	}

	return fmt.Sprintf(`received %d arguments in *%v*, expected %d`,
		n.received, n.name, n.expected)
}

// helper function to check the arity of incoming arguments for a function.
// Also accepts the case that the args slice is nil.
func checkArity(arity int, args []interface{}, name string) error {
	if args == nil {
		if arity == 0 {
			return nil
		}
		return arityError{arity, 0, name, false}
	}
	if len(args) != arity {
		return arityError{arity, len(args), name, false}
	}
	return nil
}

// defines the built-in "define" construct.  e.g.:
//
//  (define x 5)
//
// would create the symbol "x" and set its value to 5.
var define = special{
	name:  "define",
	arity: 2,
	fn: func(env *environment, args []interface{}) (interface{}, error) {
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
	},
}

// defines the built-in "quote" construct.  e.g.:
//
//  (quote (1 2 3))
//
// would evaluate to the list (1 2 3).  That is, quote is a function of arity 1
// that is effectively a no-op; the input value is not evaluated, which
// prevents evaluation of the first element of the list, in this case 1.
var quote = special{
	name:  "quote",
	arity: 1,
	fn: func(_ *environment, args []interface{}) (interface{}, error) {
		switch t := args[0].(type) {
		case *sexp:
			t.quotelvl++
			return t, nil
		default:
			return &sexp{items: []interface{}{t}, quotelvl: 1}, nil
		}
		panic("not reached")
	},
}

// turns an arbitrary lisp value into a boolean.  Apparently the sematics of
// this in lisp are that everything except false is true?  Seems weird to me,
// but ok.
func booleanize(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return true
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
var _if = special{
	name:     "if",
	arity:    2,
	variadic: true,
	fn: func(env *environment, args []interface{}) (interface{}, error) {
		if len(args) > 3 {
			return nil, arityError{
				expected: 3,
				received: len(args),
				name:     "ifpooop",
				variadic: false,
			}
		}
		v, err := eval(args[0], env)
		if err != nil {
			return nil, err
		}

		if booleanize(v) {
			return eval(args[1], env)
		}
		if len(args) == 3 {
			return eval(args[2], env)
		}
		return nil, nil
	},
}

// defines the built-in "set!" construct, which is used to set the value of an
// existing symbol in the provided environment.  e.g.:
//
//  (set! x 5)
//
// would set the symbol x to the value 5, if and only if the symbol x was
// previously defined.
var set = special{
	name:  "set!",
	arity: 2,
	fn: func(env *environment, args []interface{}) (interface{}, error) {
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
	},
}

type lambda struct {
	env       *environment
	arglabels []symbol
	body      *sexp
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
var mklambda = special{
	name:  "lambda",
	arity: 2,
	fn: func(env *environment, args []interface{}) (interface{}, error) {
		debugPrint("mklambda")

		params, ok := args[0].(*sexp)
		if !ok {
			return nil, fmt.Errorf(`first argument to *lambda* must be sexp, received %v`, reflect.TypeOf(args[0]))
		}

		arglabels := make([]symbol, 0, len(params.items))
		for _, v := range params.items {
			s, ok := v.(symbol)
			if !ok {
				return nil, fmt.Errorf(`lambda args must all be symbols; received invalid %v`, reflect.TypeOf(v))
			}
			arglabels = append(arglabels, s)
		}

		body, ok := args[1].(*sexp)
		if !ok {
			return nil, fmt.Errorf(`second argument to *lambda* must be sexp, received %v`, reflect.TypeOf(args[1]))
		}

		return lambda{env, arglabels, body}, nil
	},
}

// defines the built-in "begin" construct.  A "begin" statement evaluates each
// of its inputs, and returns the value of the evaluation of the last
// statement.  E.g.:
//
//  (begin (+ 1 1) (* 2 2) (+ 3 3))
//
// would evaluate to 6.
var begin = special{
	name:     "begin",
	variadic: true,
	fn: func(env *environment, args []interface{}) (interface{}, error) {
		debugPrint("begin")
		var err error
		var v interface{}
		for _, arg := range args {
			v, err = eval(arg, env)
			if err != nil {
				return nil, err
			}
		}
		return v, nil
	},
}

var names = special{
	name:     "names",
	arity:    0,
	variadic: false,
	fn: func(env *environment, args []interface{}) (interface{}, error) {
		return env.keys(), nil
	},
}

var and = special{
	name:     "and",
	arity:    1,
	variadic: true,
	fn: func(env *environment, args []interface{}) (interface{}, error) {
		for _, v := range args {
			t, err := eval(v, env)
			if err != nil {
				return false, err
			}
			if !booleanize(t) {
				return false, nil
			}
		}
		return true, nil
	},
}

var or = special{
	name:     "or",
	arity:    1,
	variadic: true,
	fn: func(env *environment, args []interface{}) (interface{}, error) {
		for _, v := range args {
			t, err := eval(v, env)
			if err != nil {
				return false, err
			}
			if booleanize(t) {
				return true, nil
			}
		}
		return false, nil
	},
}
