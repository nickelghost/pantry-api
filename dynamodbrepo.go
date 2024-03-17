package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type DynamoDBRepo struct {
	client         *dynamodb.DynamoDB
	locationsTable string
}

func dynamoToLocations(dynamoItems []map[string]*dynamodb.AttributeValue) ([]Location, error) {
	locations := []Location{}

	for _, i := range dynamoItems {
		loc := Location{}

		err := dynamodbattribute.UnmarshalMap(i, &loc)
		if err != nil {
			return nil, err
		}

		locations = append(locations, loc)
	}

	return locations, nil
}

func (repo DynamoDBRepo) scanGetLocations() ([]Location, error) {
	output, err := repo.client.Scan(&dynamodb.ScanInput{
		TableName: &repo.locationsTable,
	})
	if err != nil {
		return nil, err
	}

	return dynamoToLocations(output.Items)
}

func (repo DynamoDBRepo) batchGetLocations(ids []string) ([]Location, error) {
	keys := []map[string]*dynamodb.AttributeValue{}

	for _, id := range ids {
		keys = append(keys, map[string]*dynamodb.AttributeValue{
			"id": {
				S: &id,
			},
		})
	}

	output, err := repo.client.BatchGetItem(&dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			repo.locationsTable: {
				Keys: keys,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return dynamoToLocations(output.Responses[repo.locationsTable])
}

func (repo DynamoDBRepo) GetLocations(ids *[]string) ([]Location, error) {
	if ids != nil {
		return repo.batchGetLocations(*ids)
	} else {
		return repo.scanGetLocations()
	}
}

func (repo DynamoDBRepo) CreateLocation(name string) error {
	loc := Location{
		ID:   uuid.NewString(),
		Name: name,
	}

	dynamoLoc, err := dynamodbattribute.MarshalMap(loc)
	if err != nil {
		return err
	}

	_, err = repo.client.PutItem(&dynamodb.PutItemInput{
		Item:      dynamoLoc,
		TableName: &repo.locationsTable,
	})

	return err
}

func (repo DynamoDBRepo) UpdateLocation(id string, name string) error {
	_, err := repo.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: &repo.locationsTable,
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: &id,
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name": {
				S: &name,
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#name": aws.String("name"),
		},
		UpdateExpression: aws.String(`set #name = :name`),
	})

	return err
}

func (repo DynamoDBRepo) DeleteLocation(id string) error {
	// todo: use a transaction and update items as well

	_, err := repo.client.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(repo.locationsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: &id,
			},
		},
	})

	return err
}
