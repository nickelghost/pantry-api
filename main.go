package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/go-playground/validator/v10"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	validate := validator.New(validator.WithRequiredStructEnabled())

	var repo repository

	switch os.Getenv("DB") {
	case "firestore":
		client, err := firestore.NewClient(context.Background(), "personal-419019")
		if err != nil {
			slog.Error("failed to create firestore client", "err", err)

			return
		}

		defer client.Close()

		repo = firestoreRepository{client: client}
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
	}

	handler := getRouter(repo, validate)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(httpadapter.NewV2(handler).ProxyWithContext)
	} else {
		srv := getServer(handler)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", "err", err)
		}
	}
}
