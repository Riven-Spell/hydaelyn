package common

import "golang.org/x/exp/constraints"

func Ternary[valType any](cond bool, tVal valType, fVal valType) valType {
	if cond {
		return tVal
	}

	return fVal
}

func Pointer[target any](input target) *target {
	return &input
}

type Number interface {
	constraints.Float | constraints.Integer
}

func Min[num Number](a, b num) num {
	if a <= b {
		return a
	}

	return b
}

func Max[num Number](a, b num) num {
	if a >= b {
		return a
	}

	return b
}
