package main

type proc func(...interface{}) (interface{}, error)

func addition(vals ...interface{}) (interface{}, error) {
	a := accumulator{
		name: "addition",
		floatFn: func(left, right float64) float64 {
			return left + right
		},
		intFn: func(left, right int64) int64 {
			return left + right
		},
	}
	return a.total(vals...)
}

func subtraction(vals ...interface{}) (interface{}, error) {
	a := accumulator{
		name: "subtraction",
		floatFn: func(left, right float64) float64 {
			return left - right
		},
		intFn: func(left, right int64) int64 {
			return left - right
		},
	}
	return a.total(vals...)
}

func multiplication(vals ...interface{}) (interface{}, error) {
	a := accumulator{
		name: "multiplication",
		floatFn: func(left, right float64) float64 {
			return left * right
		},
		intFn: func(left, right int64) int64 {
			return left * right
		},
	}
	return a.total(vals...)
}
