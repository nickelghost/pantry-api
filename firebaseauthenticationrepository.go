package main

import (
	"context"
	"errors"
	"fmt"

	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/iterator"
)

type firebaseAuthenticationRepository struct {
	client *auth.Client
}

func (repo firebaseAuthenticationRepository) GetAllEmails(ctx context.Context) ([]string, error) {
	emails := []string{}
	iter := repo.client.Users(ctx, "")

	for {
		user, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("firebase get user: %w", err)
		}

		emails = append(emails, user.Email)
	}

	return emails, nil
}
