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
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
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

	var repo repository

	switch os.Getenv("DB") {
	case "firestore":
		firestoreRepo, err := getFirestoreRepository()
		if err != nil {
			slog.Error("failed to create firestore repo", "err", err)

			return
		}

		defer firestoreRepo.client.Close()
		repo = firestoreRepo
	case "dynamodb":
		sess := session.Must(session.NewSession())
		client := dynamodb.New(sess)

		repo = dynamoDBRepository{
			client:         client,
			locationsTable: os.Getenv("DYNAMODB_LOCATIONS_TABLE"),
			itemsTable:     os.Getenv("DYNAMODB_ITEMS_TABLE"),
		}
	default:
		slog.Error("unknown DB", "db", os.Getenv("DB"))

		return
	}

	var auth authentication

	if os.Getenv("AUTH") == "firebase" {
		var err error

		auth, err = getFirebaseAuthentication()
		if err != nil {
			slog.Error("failed to create firebase auth", "err", err)

			return
		}
	}

	handler := getRouter(repo, validate, auth)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(httpadapter.NewV2(handler).ProxyWithContext)
	} else {
		srv := getServer(handler)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", "err", err)
		}
	}
}
