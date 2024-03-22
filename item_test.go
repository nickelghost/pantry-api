package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
)

func TestCreateItem(t *testing.T) {
	t.Parallel()

	validate := validator.New(validator.WithRequiredStructEnabled())
	mockRepo := &mockRepository{}
	params := writeItemParams{
		Name:       "Cheese",
		Type:       getPtr("250g"),
		Tags:       []string{"dairy", "smelly"},
		ImageURL:   nil,
		Price:      getPtr(1499),
		BoughtAt:   time.Now(),
		ExpiresAt:  getPtr(time.Now().Add(time.Hour * 240)), // 10 days
		LocationID: getPtr("my-loc"),
	}

	err := createItem(mockRepo, validate, params)
	if err != nil {
		t.Errorf("Got error: %s", err)
	}

	if mockRepo.CreateItemCalls != 1 {
		t.Errorf("CreateItem called %d times instead of once", mockRepo.CreateLocationCalls)
	}

	if !reflect.DeepEqual(mockRepo.CreateItemParams, params) {
		t.Errorf("Got params %+v instead of %+v", mockRepo.CreateItemParams, params)
	}
}

func TestUpdateItem(t *testing.T) {
	t.Parallel()

	validate := validator.New(validator.WithRequiredStructEnabled())
	mockRepo := &mockRepository{}
	id := "cheese"
	params := writeItemParams{
		Name:       "Cheese",
		Type:       getPtr("250g"),
		Tags:       []string{"dairy", "smelly"},
		ImageURL:   nil,
		Price:      getPtr(1499),
		BoughtAt:   time.Now(),
		ExpiresAt:  getPtr(time.Now().Add(time.Hour * 240)), // 10 days
		LocationID: getPtr("my-loc"),
	}

	err := updateItem(mockRepo, validate, id, params)
	if err != nil {
		t.Errorf("Got error: %s", err)
	}

	if mockRepo.UpdateItemCalls != 1 {
		t.Errorf("UpdateItem called %d times instead of once", mockRepo.UpdateLocationCalls)
	}

	if mockRepo.UpdateItemID != id {
		t.Errorf("UpdateItem called with %s instead of %s", mockRepo.UpdateItemID, id)
	}

	if !reflect.DeepEqual(mockRepo.UpdateItemParams, params) {
		t.Errorf("Got params %+v instead of %+v", mockRepo.UpdateItemParams, params)
	}
}

func TestUpdateItemLocation(t *testing.T) {
	t.Parallel()

	data := []struct {
		id         string
		locationID *string
	}{
		{id: "potato", locationID: getPtr("pantry")},
		{id: "mayo", locationID: nil},
		{id: "strawberry", locationID: getPtr("fruit_basket")},
		{id: "chocolate", locationID: nil},
	}

	for _, row := range data {
		mockRepo := &mockRepository{}

		err := updateItemLocation(mockRepo, row.id, row.locationID)
		if err != nil {
			t.Errorf("Got error: %s", err)
		}

		if mockRepo.UpdateItemLocationID != row.id {
			t.Errorf("Called with ID %s instead of %s", mockRepo.UpdateItemLocationID, row.id)
		}

		if mockRepo.UpdateItemLocationValue != row.locationID {
			t.Errorf(
				"Called with locationID %v instead of %v",
				mockRepo.UpdateItemLocationValue, row.locationID,
			)
		}
	}
}

func TestDeleteItem(t *testing.T) {
	t.Parallel()

	ids := []string{"id1", "id2", "007"}

	for _, id := range ids {
		repo := &mockRepository{}

		err := deleteItem(repo, id)
		if err != nil {
			t.Errorf("Returned unexpected error for %s: %+v", id, err)
		}

		if repo.DeleteItemID != id {
			t.Errorf(`Called with wrong id: "%s" instead of "%s"`, repo.DeleteItemID, id)
		}
	}
}
