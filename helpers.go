package main

import "errors"

var errValidation = errors.New("validation error")

func getPtr[T any](data T) *T {
	return &data
}
