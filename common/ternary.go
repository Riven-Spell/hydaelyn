package common

func TernaryString(cond bool, tval, fval string) string {
	if cond {
		return tval
	}
	return fval
}
