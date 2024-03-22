package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/go-playground/validator/v10"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	validate := validator.New(validator.WithRequiredStructEnabled())

	sess := session.Must(session.NewSession())
	client := dynamodb.New(sess)

	repo := dynamoDBRepository{
		client:         client,
		locationsTable: os.Getenv("DYNAMODB_LOCATIONS_TABLE"),
		itemsTable:     os.Getenv("DYNAMODB_ITEMS_TABLE"),
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
