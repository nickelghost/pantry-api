package main

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func main() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}
