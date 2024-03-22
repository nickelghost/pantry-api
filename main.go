package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/go-playground/validator/v10"
)

func main() {
	validate := validator.New(validator.WithRequiredStructEnabled())

	sess := session.Must(session.NewSession())
	client := dynamodb.New(sess)

	repo := DynamoDBRepo{
		client:         client,
		locationsTable: os.Getenv("DYNAMODB_LOCATIONS_TABLE"),
		itemsTable:     os.Getenv("DYNAMODB_ITEMS_TABLE"),
	}

	handler := GetRouter(repo, validate)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(httpadapter.NewV2(handler).ProxyWithContext)
	} else {
		srv := GetServer(handler)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}
}
