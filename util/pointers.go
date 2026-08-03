package util

import "reflect"

func Pointer[T any](v T) *T {
	if reflect.ValueOf(v).IsZero() {
		return nil
	}
	return &v
}

func ZeroValuedPointer[T any](v T) *T {
	return &v
}

func DeReference[T any](v *T) T {
	if v == nil {
		var res T
		return res
	}

	return *v
}
