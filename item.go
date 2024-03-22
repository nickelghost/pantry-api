package main

import (
	"fmt"
	"time"
)

type Item struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Type           *string    `json:"type"`
	Tags           []string   `json:"tags"`
	Price          *int       `json:"price"`
	ImageURL       *string    `json:"imageUrl"`
	BoughtAt       time.Time  `json:"boughtAt"`
	OpenedAt       *time.Time `json:"openedAt"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	Quantity       *int       `json:"quantity"`
	QuantityTarget *int       `json:"quantityTarget"`
	LocationID     *string    `json:"locationId"`
	Location       *Location  `json:"location,omitempty"`
}

type WriteItemParams struct {
	Name       string     `json:"name"`
	Type       *string    `json:"type"`
	Tags       []string   `json:"tags"`
	Price      *int       `json:"price"`
	ImageURL   *string    `json:"imageUrl"`
	BoughtAt   time.Time  `json:"boughtAt"`
	OpenedAt   *time.Time `json:"openedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	LocationID *string    `json:"locationId"`
}

func CreateItem(repo Repo, params WriteItemParams) error {
	if err := repo.CreateItem(params); err != nil {
		return fmt.Errorf("create item: %w", err)
	}

	return nil
}

func UpdateItem(repo Repo, id string, params WriteItemParams) error {
	if err := repo.UpdateItem(id, params); err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

func UpdateItemQuantity(repo Repo, id string, quantity *int) error {
	if err := repo.UpdateItemQuantity(id, quantity); err != nil {
		return fmt.Errorf("update item quantity: %w", err)
	}

	return nil
}

func UpdateItemLocation(repo Repo, id string, locationID *string) error {
	if err := repo.UpdateItemLocation(id, locationID); err != nil {
		return fmt.Errorf("update item location: %w", err)
	}

	return nil
}

func DeleteItem(repo Repo, id string) error {
	if err := repo.DeleteItem(id); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}
