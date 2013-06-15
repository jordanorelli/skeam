package main

import (
	"fmt"
	"sort"
)

type UnknownSymbolError struct{ symbol }

func (u UnknownSymbolError) Error() string {
	return fmt.Sprintf(`unknown symbol "%v"`, u.symbol)
}

type environment struct {
	items map[symbol]interface{}
	outer *environment
}

func newEnvironment(outer *environment) *environment {
	return &environment{
		items: make(map[symbol]interface{}),
		outer: outer,
	}
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

func (e environment) keys() []string {
	keys := make([]string, 0, len(e.items))
	for key, _ := range e.items {
		keys = append(keys, string(key))
	}
    if e.outer != nil {
        keys = append(keys, e.outer.keys()...)
    }
	sort.Strings(keys)
	return keys
}

func (e environment) defined(key symbol) bool {
	_, err := e.get(key)
	return err == nil
}
