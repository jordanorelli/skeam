package main

import (
	"fmt"
)

type UnknownSymbolError struct{ symbol }

func (u UnknownSymbolError) Error() string {
	return fmt.Sprintf(`unknown symbol "%v"`, u.symbol)
}

type environment struct {
	items map[symbol]interface{}
	outer *environment
}

func (e environment) get(key symbol) (interface{}, error) {
	v, ok := e.items[key]
	if ok {
		debugPrint(fmt.Sprintf(`found key "%v": %v`, key, v))
		return v, nil
	}

	if e.outer != nil {
		return e.outer.get(key)
	}

	return nil, UnknownSymbolError{key}
}

func (e environment) set(key symbol, val interface{}) {
	e.items[key] = val
}

func (e environment) defined(key symbol) bool {
	_, err := e.get(key)
	return err == nil
}
