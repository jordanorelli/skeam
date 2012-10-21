package main

import (
	"errors"
	"fmt"
	"reflect"
)

type builtin func([]interface{}) (interface{}, error)

// evaluates all of the arguments, and then calls the function with the results
// of the evaluations
func (b *builtin) call(env *environment, rawArgs []interface{}) (interface{}, error) {
	if rawArgs == nil {
		return (*b)(nil)
	}

	// eval all arguments first
	args := make([]interface{}, 0, len(rawArgs))
	for _, raw := range rawArgs {
		v, err := eval(raw, env)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}

	return (*b)(args)
}

func addition(vals []interface{}) (interface{}, error) {
	a := accumulator{
		name: "addition",
		floatFn: func(left, right float64) (float64, error) {
			return left + right, nil
		},
		intFn: func(left, right int64) (int64, error) {
			return left + right, nil
		},
	}
	return a.total(vals)
}

func subtraction(vals []interface{}) (interface{}, error) {
	a := accumulator{
		name: "subtraction",
		floatFn: func(left, right float64) (float64, error) {
			return left - right, nil
		},
		intFn: func(left, right int64) (int64, error) {
			return left - right, nil
		},
	}
	return a.total(vals)
}

func multiplication(vals []interface{}) (interface{}, error) {
	a := accumulator{
		name: "multiplication",
		floatFn: func(left, right float64) (float64, error) {
			return left * right, nil
		},
		intFn: func(left, right int64) (int64, error) {
			return left * right, nil
		},
		acc:  1,
		accf: 1.0,
	}
	return a.total(vals)
}

func division(vals []interface{}) (interface{}, error) {
	a := accumulator{
		name: "division",
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
	}
	return a.total(vals)
}

func not(vals []interface{}) (interface{}, error) {
	if err := checkArity(1, vals, "not"); err != nil {
		return nil, err
	}
	return !booleanize(vals[0]), nil
}

func length(vals []interface{}) (interface{}, error) {
	if err := checkArity(1, vals, "length"); err != nil {
		return nil, err
	}

	x, ok := vals[0].(sexp)
	if !ok {
		return nil, fmt.Errorf("first argument must be sexp, received %v", reflect.TypeOf(vals[0]))
	}
	return len(x), nil
}

/*
func car(vals []interface{}) (interface{}, error) {

}
*/
