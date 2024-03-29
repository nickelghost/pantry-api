package main

import (
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
	ImageURL   *string    `json:"imageUrl"`
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
	Tags       []string   `json:"tags"`
	Price      *int       `json:"price"      validate:"omitempty,gte=0"`
	ImageURL   *string    `json:"imageUrl"   validate:"omitempty,url"`
	BoughtAt   time.Time  `json:"boughtAt"`
	OpenedAt   *time.Time `json:"openedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	Lifespan   *int       `json:"lifespan"   validate:"omitempty,gte=0"`
	LocationID *string    `json:"locationId"`
}

func createItem(repo repository, validate *validator.Validate, params writeItemParams) error {
	if err := validate.Struct(params); err != nil {
		return fmt.Errorf("invalid write item params: %w", err)
	}

	if err := repo.CreateItem(params); err != nil {
		return fmt.Errorf("create item: %w", err)
	}

	return nil
}

func updateItem(repo repository, validate *validator.Validate, id string, params writeItemParams) error {
	if err := validate.Struct(params); err != nil {
		return fmt.Errorf("invalid write item params: %w", err)
	}

	if err := repo.UpdateItem(id, params); err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

func updateItemLocation(repo repository, id string, locationID *string) error {
	if err := repo.UpdateItemLocation(id, locationID); err != nil {
		return fmt.Errorf("update item location: %w", err)
	}

	return nil
}

func deleteItem(repo repository, id string) error {
	if err := repo.DeleteItem(id); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}
