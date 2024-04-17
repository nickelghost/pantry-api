package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
)

type firebaseAuthentication struct {
	client *auth.Client
}

func (auth firebaseAuthentication) Check(ctx context.Context, r *http.Request) error {
	idToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	_, err := auth.client.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	if err != nil {
		return fmt.Errorf("failed to verify ID token: %w", err)
	}

	return nil
}
