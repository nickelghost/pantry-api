package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/go-playground/validator/v10"
)

func getFirebaseAuthentication() (firebaseAuthentication, error) {
	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: os.Getenv("CLOUDSDK_CORE_PROJECT"),
	})
	if err != nil {
		return firebaseAuthentication{}, fmt.Errorf("failed to create firebase app: %w", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return firebaseAuthentication{}, fmt.Errorf("failed to create firebase auth client: %w", err)
	}

	return firebaseAuthentication{client: client}, nil
}

func getFirestoreRepository() (firestoreRepository, error) {
	client, err := firestore.NewClientWithDatabase(
		context.Background(),
		os.Getenv("CLOUDSDK_CORE_PROJECT"),
		os.Getenv("FIRESTORE_DATABASE"),
	)
	if err != nil {
		return firestoreRepository{}, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return firestoreRepository{client: client}, nil
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	validate := validator.New(validator.WithRequiredStructEnabled())

	firestoreRepo, err := getFirestoreRepository()
	if err != nil {
		slog.Error("failed to create firestore repo", "err", err)

		return
	}

	defer firestoreRepo.client.Close()

	var auth authentication

	if os.Getenv("AUTH") == "firebase" {
		var err error

		auth, err = getFirebaseAuthentication()
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
