package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type item struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Type       *string    `json:"type"`
	Tags       []string   `json:"tags"`
	Price      *int       `json:"price"`
	BoughtAt   time.Time  `json:"boughtAt"`
	OpenedAt   *time.Time `json:"openedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	Lifespan   *int       `json:"lifespan"`
	LocationID *string    `json:"locationId"`
	Location   *location  `json:"location,omitempty"`
}

type writeItemParams struct {
	Name       string     `json:"name"       validate:"required,min=2"`
	Type       *string    `json:"type"`
	Tags       []string   `json:"tags"       validate:"required"`
	Price      *int       `json:"price"      validate:"omitempty,gte=0"`
	BoughtAt   time.Time  `json:"boughtAt"   validate:"required"`
	OpenedAt   *time.Time `json:"openedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	Lifespan   *int       `json:"lifespan"   validate:"omitempty,gte=0"`
	LocationID *string    `json:"locationId"`
}

func createItem(ctx context.Context, repo repository, validate *validator.Validate, params writeItemParams) error {
	if err := validate.Struct(params); err != nil {
		return fmt.Errorf("invalid write item params: %w", err)
	}

	if err := repo.CreateItem(ctx, params); err != nil {
		return fmt.Errorf("create item: %w", err)
	}

	return nil
}

func updateItem(
	ctx context.Context,
	repo repository,
	validate *validator.Validate,
	id string,
	params writeItemParams,
) error {
	if err := validate.Struct(params); err != nil {
		return fmt.Errorf("invalid write item params: %w", err)
	}

	if err := repo.UpdateItem(ctx, id, params); err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

func updateItemLocation(ctx context.Context, repo repository, id string, locationID *string) error {
	if err := repo.UpdateItemLocation(ctx, id, locationID); err != nil {
		return fmt.Errorf("update item location: %w", err)
	}

	return nil
}

func deleteItem(ctx context.Context, repo repository, id string) error {
	if err := repo.DeleteItem(ctx, id); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}
