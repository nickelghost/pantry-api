package main

import (
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestMain(m *testing.M) {
	validate = validator.New(validator.WithRequiredStructEnabled())

	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestMainFunc(t *testing.T) {
	main()

	if validate == nil {
		t.Error("validate was not initialised")
	}
}
