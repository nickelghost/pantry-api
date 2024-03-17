package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func main() {
	validate = validator.New(validator.WithRequiredStructEnabled())

	sess := session.Must(session.NewSession())
	client := dynamodb.New(sess)

	repo := DynamoDBRepo{
		client:         client,
		locationsTable: os.Getenv("DYNAMODB_LOCATIONS_TABLE"),
		itemsTable:     os.Getenv("DYNAMODB_ITEMS_TABLE"),
	}

	handler := GetRouter(repo)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(httpadapter.NewV2(handler).ProxyWithContext)
	} else {
		srv := &http.Server{
			Addr:    net.JoinHostPort("0.0.0.0", "8080"),
			Handler: handler,
		}

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}
}
