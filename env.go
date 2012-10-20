package main

import (
	"fmt"
)

type UnknownSymbolError struct{ symbol }

func (u UnknownSymbolError) Error() string {
	return fmt.Sprintf(`unknown symbol "%v"`, u.symbol)
}

type environment map[symbol]interface{}

func (e environment) get(key symbol) (interface{}, error) {
	v, ok := e[key]
	if ok {
		debugPrint(fmt.Sprintf(`found key "%v": %v`, key, v))
		return v, nil
	}
	return nil, UnknownSymbolError{key}
}
