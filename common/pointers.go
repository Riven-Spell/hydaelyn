package common

func Pointer[target any](input target) *target {
	return &input
}
