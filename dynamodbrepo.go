package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

type DynamoDBRepo struct {
	client         *dynamodb.DynamoDB
	locationsTable string
	itemsTable     string
}

func dynamoToLocations(dynamoItems []map[string]*dynamodb.AttributeValue) ([]Location, error) {
	locations := []Location{}

	for _, i := range dynamoItems {
		loc := Location{}

		if err := dynamodbattribute.UnmarshalMap(i, &loc); err != nil {
			return nil, fmt.Errorf("unmarshal location: %w", err)
		}

		locations = append(locations, loc)
	}

	return locations, nil
}

func dynamoToItems(dynamoItems []map[string]*dynamodb.AttributeValue) ([]Item, error) {
	items := []Item{}

	for _, i := range dynamoItems {
		item := Item{}

		if err := dynamodbattribute.UnmarshalMap(i, &item); err != nil {
			return nil, fmt.Errorf("unmarshal item: %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}

func (repo DynamoDBRepo) scanGetLocations() ([]Location, error) {
	output, err := repo.client.Scan(&dynamodb.ScanInput{
		TableName: &repo.locationsTable,
	})
	if err != nil {
		return nil, fmt.Errorf("scan locations: %w", err)
	}

	return dynamoToLocations(output.Items)
}

func (repo DynamoDBRepo) batchGetLocations(ids []string) ([]Location, error) {
	keys := []map[string]*dynamodb.AttributeValue{}

	for _, id := range ids {
		keys = append(keys, map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
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
		return nil, fmt.Errorf("dynamodb batch get locations: %w", err)
	}

	return dynamoToLocations(output.Responses[repo.locationsTable])
}

func (repo DynamoDBRepo) GetLocations(ids *[]string) ([]Location, error) {
	if ids != nil {
		return repo.batchGetLocations(*ids)
	}

	return repo.scanGetLocations()
}

func (repo DynamoDBRepo) CreateLocation(name string) error {
	loc := Location{
		ID:   uuid.NewString(),
		Name: name,
	}

	dynamoLoc, err := dynamodbattribute.MarshalMap(loc)
	if err != nil {
		return fmt.Errorf("marshal location create params: %w", err)
	}

	_, err = repo.client.PutItem(&dynamodb.PutItemInput{
		TableName: &repo.locationsTable,
		Item:      dynamoLoc,
	})
	if err != nil {
		return fmt.Errorf("dynamodb create location: %w", err)
	}

	return nil
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
	if err != nil {
		return fmt.Errorf("dynamodb update location in: %w", err)
	}

	return nil
}

func (repo DynamoDBRepo) DeleteLocation(id string) error {
	items, err := repo.GetItems(nil, nil, &[]string{id})
	if err != nil {
		return err
	}

	updates := []*dynamodb.TransactWriteItem{}
	for _, item := range items {
		updates = append(updates, &dynamodb.TransactWriteItem{
			Update: &dynamodb.Update{
				TableName: &repo.itemsTable,
				Key: map[string]*dynamodb.AttributeValue{
					"id": {
						S: aws.String(item.ID),
					},
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":null": {
						NULL: aws.Bool(true),
					},
				},
				UpdateExpression: aws.String(`set locationId = :null`),
			},
		})
	}

	_, err = repo.client.TransactWriteItems(&dynamodb.TransactWriteItemsInput{
		TransactItems: append(updates, &dynamodb.TransactWriteItem{
			Delete: &dynamodb.Delete{
				TableName: &repo.locationsTable,
				Key: map[string]*dynamodb.AttributeValue{
					"id": {
						S: &id,
					},
				},
			},
		}),
	})
	if err != nil {
		return fmt.Errorf("dynamodb transaction write: %w", err)
	}

	return nil
}

func (repo DynamoDBRepo) GetItems(
	search *string,
	tags *[]string,
	locationIDs *[]string,
) ([]Item, error) {
	conditions := []expression.ConditionBuilder{}

	if search != nil {
		conditions = append(conditions, expression.Name("name").Contains(*search))
	}

	if tags != nil {
		for _, tag := range *tags {
			conditions = append(conditions, expression.Name("tags").Contains(tag))
		}
	}

	if locationIDs != nil {
		values := []expression.OperandBuilder{}

		for _, locationID := range *locationIDs {
			values = append(values, expression.Value(locationID))
		}

		conditions = append(conditions, expression.Name("locationId").In(values[0], values[1:]...))
	}

	scanInput := dynamodb.ScanInput{
		TableName: &repo.itemsTable,
	}

	if len(conditions) > 0 {
		fltr := conditions[0]

		for _, cond := range conditions[1:] {
			fltr = fltr.And(cond)
		}

		expr, err := expression.NewBuilder().WithFilter(fltr).Build()
		if err != nil {
			return nil, fmt.Errorf("build items filter: %w", err)
		}

		scanInput.ExpressionAttributeNames = expr.Names()
		scanInput.ExpressionAttributeValues = expr.Values()
		scanInput.FilterExpression = expr.Filter()
	}

	res, err := repo.client.Scan(&scanInput)
	if err != nil {
		return nil, fmt.Errorf("scan item: %w", err)
	}

	return dynamoToItems(res.Items)
}

type DynamoDBWriteItemParams struct {
	Name       string     `json:":name"`
	Type       *string    `json:":type"`
	Tags       []string   `json:":tags"`
	Price      *int       `json:":price"`
	ImageURL   *string    `json:":imageUrl"`
	BoughtAt   time.Time  `json:":boughtAt"`
	OpenedAt   *time.Time `json:":openedAt"`
	ExpiresAt  *time.Time `json:":expiresAt"`
	LocationID *string    `json:":locationId"`
}

func (repo DynamoDBRepo) CreateItem(params WriteItemParams) error {
	i := Item{
		ID:         uuid.NewString(),
		Name:       params.Name,
		Type:       params.Type,
		Tags:       params.Tags,
		Price:      params.Price,
		ImageURL:   params.ImageURL,
		BoughtAt:   params.BoughtAt,
		OpenedAt:   params.OpenedAt,
		ExpiresAt:  params.ExpiresAt,
		LocationID: params.LocationID,
	}

	dynamoItem, err := dynamodbattribute.MarshalMap(i)
	if err != nil {
		return fmt.Errorf("marshal write item params: %w", err)
	}

	_, err = repo.client.PutItem(&dynamodb.PutItemInput{
		TableName: &repo.itemsTable,
		Item:      dynamoItem,
	})
	if err != nil {
		return fmt.Errorf("dynamodb create item: %w", err)
	}

	return nil
}

func (repo DynamoDBRepo) UpdateItem(id string, params WriteItemParams) error {
	attrs, err := dynamodbattribute.MarshalMap(DynamoDBWriteItemParams(params))
	if err != nil {
		return fmt.Errorf("marshal write item params: %w", err)
	}

	_, err = repo.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: &repo.itemsTable,
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: &id,
			},
		},
		ExpressionAttributeValues: attrs,
		ExpressionAttributeNames: map[string]*string{
			"#name": aws.String("name"),
			"#type": aws.String("type"),
		},
		UpdateExpression: aws.String(`set
			#name = :name,
			#type = :type,
			tags = :tags,
			price = :price,
			imageUrl = :imageUrl,
			boughtAt = :boughtAt,
			openedAt = :openedAt,
			expiresAt = :expiresAt,
			locationId = :locationId
		`),
	})
	if err != nil {
		return fmt.Errorf("dynamodb update item: %w", err)
	}

	return nil
}

func (repo DynamoDBRepo) UpdateItemLocation(id string, locationID *string) error {
	attrs, err := dynamodbattribute.MarshalMap(struct {
		LocationID *string `json:":locationId"`
	}{
		LocationID: locationID,
	})
	if err != nil {
		return fmt.Errorf("update fields marshal: %w", err)
	}

	_, err = repo.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: &repo.itemsTable,
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: &id,
			},
		},
		ExpressionAttributeValues: attrs,
		UpdateExpression:          aws.String(`set locationId = :locationId`),
	})
	if err != nil {
		return fmt.Errorf("dynamodb update item: %w", err)
	}

	return nil
}

func (repo DynamoDBRepo) DeleteItem(id string) error {
	_, err := repo.client.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(repo.itemsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: &id,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("dynamodb delete item: %w", err)
	}

	return nil
}
