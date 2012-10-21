package main

import (
	"fmt"
	"reflect"
)

type special func(*environment, ...interface{}) (interface{}, error)

func define(env *environment, args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(`received %d arguments in *define*, expected exactly 2`, len(args))
	}
	s, ok := args[0].(symbol)
	if !ok {
		return nil, fmt.Errorf(`first argument to *define* must be symbol, received %v`, reflect.TypeOf(args[0]))
	}
	env.set(s, args[1])
	return nil, nil
}

func quote(_ *environment, args ...interface{}) (interface{}, error) {
	return sexp(args), nil
}

func _if(env *environment, args ...interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(`received %d arguments in *if*, expected exactly 3`, len(args))
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
