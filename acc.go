package main

import (
	"fmt"
	"reflect"
)

type accumulator struct {
	name     string
	floatFn  func(float64, float64) float64
	intFn    func(int64, int64) int64
	acc      int64
	accf     float64
	floating bool
}

func (a *accumulator) total(vals ...interface{}) (interface{}, error) {
	if len(vals) == 0 {
		return int64(0), nil
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
			if a.floating {
				a.accf = a.floatFn(a.accf, float64(v))
				break
			}
			a.acc = a.intFn(a.acc, v)
		case float64:
			if !a.floating {
				a.floating = true
				a.accf = a.floatFn(a.accf, float64(a.acc))
			}
			a.accf = a.floatFn(a.accf, v)
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
