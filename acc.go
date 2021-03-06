package main

import (
	"fmt"
	"reflect"
)

// type accumulator describes an accumulator.  That is, it is a numerical
// structure that applies a pair of functions across a list of values that are
// expected to be numerical; i.e. of type int64 or float64.
type accumulator struct {
	name     string
	floatFn  func(float64, float64) (float64, error)
	intFn    func(int64, int64) (int64, error)
	acc      int64
	accf     float64
	floating bool
}

// runs the accumulator accros the set of values, applying the accumulator's
// intFn and floatFn functions in order.  It's basically just a left fold, and
// it starts using floatFn once the first float value is encountered, and then
// thereafter.
func (a accumulator) total(vals []interface{}) (interface{}, error) {
	if vals == nil || len(vals) == 0 {
		return a.acc, nil
	}

	switch v := vals[0].(type) {
	case int64:
		a.acc = v
	case float64:
		a.floating = true
		a.accf = v
	default:
		return nil, fmt.Errorf("%v is not defined for %v", a.name, reflect.TypeOf(v))
	}

	if len(vals) == 1 {
		if a.floating {
			return a.accf, nil
		} else {
			return a.acc, nil
		}
	}

	for _, raw := range vals[1:] {
		switch v := raw.(type) {
		case int64:
			var err error
			if a.floating {
				a.accf, err = a.floatFn(a.accf, float64(v))
			} else {
				a.acc, err = a.intFn(a.acc, v)
			}
			if err != nil {
				return nil, err
			}
		case float64:
			var err error
			if a.floating {
				a.accf, err = a.floatFn(a.accf, v)
			} else {
				a.floating = true
				a.accf, err = a.floatFn(a.accf, float64(a.acc))
			}
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("%v is not defined for %v", a.name, reflect.TypeOf(v))
		}
	}

	if a.floating {
		return a.accf, nil
	} else {
		return a.acc, nil
	}
	panic("not reached")
}
