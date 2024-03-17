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
