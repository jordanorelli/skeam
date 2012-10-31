package main

import (
	"errors"
	"fmt"
	"reflect"
)

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

var gt = builtin{
	name:     ">",
	variadic: true,
	fn: func(vals []interface{}) (interface{}, error) {
		fni := func(x, y int64) bool { return x > y }
		fnf := func(x, y float64) bool { return x > y }
		return cmp_left(vals, fni, fnf)
	},
}

var gte = builtin{
	name:     ">=",
	variadic: true,
	fn: func(vals []interface{}) (interface{}, error) {
		fni := func(x, y int64) bool { return x >= y }
		fnf := func(x, y float64) bool { return x >= y }
		return cmp_left(vals, fni, fnf)
	},
}

var lt = builtin{
	name:     "<",
	variadic: true,
	fn: func(vals []interface{}) (interface{}, error) {
		fni := func(x, y int64) bool { return x < y }
		fnf := func(x, y float64) bool { return x < y }
		return cmp_left(vals, fni, fnf)
	},
}

var lte = builtin{
	name:     "<=",
	variadic: true,
	fn: func(vals []interface{}) (interface{}, error) {
		fni := func(x, y int64) bool { return x <= y }
		fnf := func(x, y float64) bool { return x <= y }
		return cmp_left(vals, fni, fnf)
	},
}
