package main

func getPtr[T any](data T) *T {
	return &data
}
