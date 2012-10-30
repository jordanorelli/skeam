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

func list(vals []interface{}) (interface{}, error) {
	return sexp(vals), nil
}

func islist(vals []interface{}) (interface{}, error) {
	if err := checkArity(1, vals, "list?"); err != nil {
		return nil, err
	}

	_, ok := vals[0].(sexp)
	return ok, nil
}

func isnull(vals []interface{}) (interface{}, error) {
	if err := checkArity(1, vals, "null?"); err != nil {
		return nil, err
	}

	s, ok := vals[0].(sexp)
	if !ok {
		return false, nil
	}

	return len(s) == 0, nil
}

func issymbol(vals []interface{}) (interface{}, error) {
	if err := checkArity(1, vals, "symbol?"); err != nil {
		return nil, err
	}

	_, ok := vals[0].(symbol)
	return ok, nil
}

type cmp_bin_i func(int64, int64) bool
type cmp_bin_f func(float64, float64) bool

func cmp_left(vals []interface{}, fni cmp_bin_i, fnf cmp_bin_f) (bool, error) {
	if len(vals) < 2 {
		return false, errors.New("expected at least 2 arguments")
	}

	var lasti int64
	var lastf float64
	var floating bool

	switch v := vals[0].(type) {
	case float64:
		floating = true
		lastf = v
	case int64:
		lasti = v
	default:
		return false, fmt.Errorf("gt is not defined for %v", reflect.TypeOf(v))
	}

	for _, raw := range vals[1:] {
		switch v := raw.(type) {
		case float64:
			if !floating {
				floating = true
				lastf = float64(lasti)
			}
			if !fnf(lastf, v) {
				return false, nil
			}
			lastf = v
		case int64:
			if floating {
				f := float64(v)
				if !fnf(lastf, f) {
					return false, nil
				}
				lastf = f
			} else {
				if !fni(lasti, v) {
					return false, nil
				}
				lasti = v
			}
		default:
			return false, errors.New("ooga booga")
		}
	}

	return true, nil
}

func gt(vals []interface{}) (interface{}, error) {
	fni := func(x, y int64) bool { return x > y }
	fnf := func(x, y float64) bool { return x > y }
	return cmp_left(vals, fni, fnf)
}

func gte(vals []interface{}) (interface{}, error) {
	fni := func(x, y int64) bool { return x >= y }
	fnf := func(x, y float64) bool { return x >= y }
	return cmp_left(vals, fni, fnf)
}

func lt(vals []interface{}) (interface{}, error) {
	fni := func(x, y int64) bool { return x < y }
	fnf := func(x, y float64) bool { return x < y }
	return cmp_left(vals, fni, fnf)
}

func lte(vals []interface{}) (interface{}, error) {
	fni := func(x, y int64) bool { return x <= y }
	fnf := func(x, y float64) bool { return x <= y }
	return cmp_left(vals, fni, fnf)
}
