package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/iterator"
)

func getFirebaseAuthentication(ctx context.Context) (firebaseAuthentication, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: os.Getenv("CLOUDSDK_CORE_PROJECT"),
	})
	if err != nil {
		return firebaseAuthentication{}, fmt.Errorf("failed to create firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return firebaseAuthentication{}, fmt.Errorf("failed to create firebase auth client: %w", err)
	}

	return firebaseAuthentication{client: client, tracer: otel.Tracer("firebase-auth")}, nil
}

type firebaseAuthentication struct {
	client *auth.Client
	tracer trace.Tracer
}

func (auth firebaseAuthentication) Check(ctx context.Context, r *http.Request) error {
	ctx, span := auth.tracer.Start(ctx, "firebaseAuthentication.Check")
	defer span.End()

	idToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	_, err := auth.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return fmt.Errorf("failed to verify ID token: %w", err)
	}

	return nil
}

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

	if len(emails) == 0 {
		return nil, errNoEmailAddressesFound
	}

	return emails, nil
}
