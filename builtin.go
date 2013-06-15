package main

import (
	"errors"
	"fmt"
	"reflect"
)

type builtin struct {
	// name of the function
	name string

	// minimum number of arguments
	arity int

	// whether we will accept arbitrarily large numbers of arguments or not
	variadic bool

	// function to be called
	fn func([]interface{}) (interface{}, error)
}

// begins by evaluating all of its inputs.  An error on input evaluation will
// stop evaluation of a builtin.  After evaluating its inputs, an arity check
// is performed to see if the proper number of arguments have been supplied.
// Perhaps this is the wrong order, I'm unsure.  Finally, the procudure is
// passed the post-evaluation arguments to be executed.
func (b builtin) call(env *environment, rawArgs []interface{}) (interface{}, error) {
	// eval all arguments first
	args := make([]interface{}, 0, len(rawArgs))
	for _, raw := range rawArgs {
		v, err := eval(raw, env)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}

	if err := b.checkArity(len(rawArgs)); err != nil {
		return nil, err
	}

	return b.fn(args)
}

func (b builtin) checkArity(n int) error {
	if n == b.arity {
		return nil
	}

	if b.variadic && n > b.arity {
		return nil
	}

	return arityError{
		expected: b.arity,
		received: n,
		name:     b.name,
		variadic: b.variadic,
	}
}

var add = builtin{
	name:     "+",
	variadic: true,
	fn: func(vals []interface{}) (interface{}, error) {
		return accumulator{
			name: "+",
			floatFn: func(left, right float64) (float64, error) {
				return left + right, nil
			},
			intFn: func(left, right int64) (int64, error) {
				return left + right, nil
			},
		}.total(vals)
	},
}

var sub = builtin{
	name:     "-",
	variadic: true,
	fn: func(vals []interface{}) (interface{}, error) {
		return accumulator{
			name: "-",
			floatFn: func(left, right float64) (float64, error) {
				return left - right, nil
			},
			intFn: func(left, right int64) (int64, error) {
				return left - right, nil
			},
		}.total(vals)
	},
}

var mul = builtin{
	name:     "*",
	variadic: true,
	fn: func(vals []interface{}) (interface{}, error) {
		return accumulator{
			name: "*",
			floatFn: func(left, right float64) (float64, error) {
				return left * right, nil
			},
			intFn: func(left, right int64) (int64, error) {
				return left * right, nil
			},
			acc:  1,
			accf: 1.0,
		}.total(vals)
	},
}

var div = builtin{
	name:     "/",
	variadic: true,
	fn: func(vals []interface{}) (interface{}, error) {
		return accumulator{
			name: "/",
			floatFn: func(left, right float64) (float64, error) {
				if right == 0.0 {
					return 0.0, errors.New("float division by zero")
				}
				return left / right, nil
			},
			intFn: func(left, right int64) (int64, error) {
				if right == 0 {
					return 0, errors.New("int division by zero")
				}
				return left / right, nil
			},
		}.total(vals)
	},
}

var not = builtin{
	name:  "not",
	arity: 1,
	fn: func(vals []interface{}) (interface{}, error) {
		return !booleanize(vals[0]), nil
	},
}

var length = builtin{
	name:  "length",
	arity: 1,
	fn: func(vals []interface{}) (interface{}, error) {
		s, ok := vals[0].(*sexp)
		if !ok {
			return nil, fmt.Errorf("first argument must be sexp, received %v", reflect.TypeOf(vals[0]))
		}
		return len(s.items), nil
	},
}

var lst = builtin{
	name:     "list",
	variadic: true,
	fn: func(vals []interface{}) (interface{}, error) {
		return &sexp{items: vals, quotelvl: 1}, nil
	},
}

var islist = builtin{
	name:  "list?",
	arity: 1,
	fn: func(vals []interface{}) (interface{}, error) {
		_, ok := vals[0].(*sexp)
		return ok, nil
	},
}

var isnull = builtin{
	name:  "null?",
	arity: 1,
	fn: func(vals []interface{}) (interface{}, error) {
		s, ok := vals[0].(*sexp)
		if !ok {
			return false, nil
		}
		return len(s.items) == 0, nil
	},
}

var issymbol = builtin{
	name:  "symbol?",
	arity: 1,
	fn: func(vals []interface{}) (interface{}, error) {
		_, ok := vals[0].(symbol)
		return ok, nil
	},
}

var cons = builtin{
	name:  "cons",
	arity: 2,
	fn: func(vals []interface{}) (interface{}, error) {
		s := &sexp{items: vals[0:1]}
		switch t := vals[1].(type) {
		case *sexp:
			s.items = append(s.items, t.items...)
		default:
			s.items = append(s.items, t)
		}
		return s, nil
	},
}

var car = builtin{
	name:  "car",
	arity: 1,
	fn: func(vals []interface{}) (interface{}, error) {
		s, ok := vals[0].(*sexp)
		if !ok {
			return nil, errors.New("expected list")
		}
		return s.items[0], nil
	},
}

var cdr = builtin{
	name:  "cdr",
	arity: 1,
	fn: func(vals []interface{}) (interface{}, error) {
		s, ok := vals[0].(*sexp)
		if !ok {
			return nil, errors.New("expected list")
		}
		return s.items[1:], nil
	},
}
