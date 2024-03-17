package main

import "time"

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
}

func CreateItem(repo Repo, params WriteItemParams) error {
	// todo: add validation
	return repo.CreateItem(params)
}

func UpdateItem(repo Repo, id string, params WriteItemParams) error {
	// todo: add validation
	return repo.UpdateItem(id, params)
}

func UpdateItemQuantity(repo Repo, id string, quantity *int) error {
	// todo: add validation
	return repo.UpdateItemQuantity(id, quantity)
}

func UpdateItemLocation(repo Repo, id string, locationID *string) error {
	// todo: add validation
	return repo.UpdateItemLocation(id, locationID)
}

func DeleteItem(repo Repo, id string) error {
	return repo.DeleteItem(id)
}
