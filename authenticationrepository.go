package main

import "context"

type authenticationRepository interface {
	GetAllEmails(ctx context.Context) ([]string, error)
}
