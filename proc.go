package main

import (
	"fmt"
	"reflect"
)

type proc func(...interface{}) (interface{}, error)

func addition(vals ...interface{}) (interface{}, error) {
	addFloats := false
	var accf float64
	var acc int64

	for _, raw := range vals {
		switch v := raw.(type) {
		case int64:
			if addFloats {
				accf += float64(v)
			} else {
				acc += v
			}
		case float64:
			if !addFloats {
				addFloats = true
				accf += float64(acc)
			}
			accf += v
		default:
			return nil, fmt.Errorf("addition is not defined for %v", reflect.TypeOf(v))
		}
	}

	if addFloats {
		return accf, nil
	} else {
		return acc, nil
	}
	panic("not reached")
}
