package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"go.opentelemetry.io/otel/trace"
)

type firebaseAuthentication struct {
	client *auth.Client
	tracer trace.Tracer
}

func (auth firebaseAuthentication) Check(ctx context.Context, r *http.Request) error {
	ctx, span := auth.tracer.Start(ctx, "firebaseAuthentication.Check")
	defer span.End()

	idToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	_, err := auth.client.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	if err != nil {
		return fmt.Errorf("failed to verify ID token: %w", err)
	}

	return nil
}
