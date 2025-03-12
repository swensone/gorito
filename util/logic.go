package util

func Ternary[A any](condition bool, trueVal A, falseVal A) A {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}
