package common

func BoolVar(b bool) *bool {
	return &b
}

func StringVar(s string) *string {
	return &s
}

func Int64Var(i int64) *int64 {
	return &i
}
