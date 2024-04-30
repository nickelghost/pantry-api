package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/trace"
)

func getFirebaseAuthentication(ctx context.Context, tracer trace.Tracer) (firebaseAuthentication, error) {
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

	return firebaseAuthentication{client: client, tracer: tracer}, nil
}

func getFirestoreRepository(ctx context.Context, tracer trace.Tracer) (firestoreRepository, error) {
	client, err := firestore.NewClientWithDatabase(ctx,
		os.Getenv("CLOUDSDK_CORE_PROJECT"),
		os.Getenv("FIRESTORE_DATABASE"),
	)
	if err != nil {
		return firestoreRepository{}, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return firestoreRepository{client: client, tracer: tracer}, nil
}

func main() {
	ctx := context.Background()

	switch strings.ToLower(os.Getenv("LOG_FORMAT")) {
	case "json":
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	case "google_cloud":
		slog.SetDefault(slog.New(NewCloudLoggingHandler()))
	default:
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	tracer, tracerShutdown, err := getTracer(ctx)
	if err != nil {
		slog.Error("failed to set up tracing", "err", err)

		return
	}

	defer tracerShutdown()

	firestoreRepo, err := getFirestoreRepository(ctx, tracer)
	if err != nil {
		slog.Error("failed to create firestore repo", "err", err)

		return
	}

	defer firestoreRepo.client.Close()

	var auth authentication

	if os.Getenv("SKIP_AUTH") != "true" {
		var err error

		auth, err = getFirebaseAuthentication(ctx, tracer)
		if err != nil {
			slog.Error("failed to create firebase auth", "err", err)

			return
		}
	}

	srv := getServer(getRouter(firestoreRepo, validate, auth))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "err", err)
	}
}
