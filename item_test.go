package main

import (
	"reflect"
	"testing"
	"time"
)

func TestCreateItem(t *testing.T) {
	mockRepo := &MockRepo{}
	params := WriteItemParams{
		Name:           "Cheese",
		Type:           getPtr("250g"),
		Tags:           []string{"dairy", "smelly"},
		ImageURL:       nil,
		Price:          getPtr(1499),
		BoughtAt:       time.Now(),
		ExpiresAt:      getPtr(time.Now().Add(time.Hour * 240)), // 10 days
		Quantity:       getPtr(1),
		QuantityTarget: getPtr(2),
		LocationID:     getPtr("my-loc"),
	}

	err := CreateItem(mockRepo, params)
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
	mockRepo := &MockRepo{}
	id := "cheese"
	params := WriteItemParams{
		Name:           "Cheese",
		Type:           getPtr("250g"),
		Tags:           []string{"dairy", "smelly"},
		ImageURL:       nil,
		Price:          getPtr(1499),
		BoughtAt:       time.Now(),
		ExpiresAt:      getPtr(time.Now().Add(time.Hour * 240)), // 10 days
		Quantity:       getPtr(1),
		QuantityTarget: getPtr(2),
		LocationID:     getPtr("my-loc"),
	}

	err := UpdateItem(mockRepo, id, params)
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

func TestUpdateItemQuantity(t *testing.T) {
	data := []struct {
		id       string
		quantity *int
	}{
		{id: "potato", quantity: getPtr(8)},
		{id: "mayo", quantity: nil},
		{id: "strawberry", quantity: getPtr(21)},
		{id: "chocolate", quantity: getPtr(0)},
	}

	for _, row := range data {
		mockRepo := &MockRepo{}
		err := UpdateItemQuantity(mockRepo, row.id, row.quantity)
		if err != nil {
			t.Errorf("Got error: %s", err)
		}

		if mockRepo.UpdateItemQuantityID != row.id {
			t.Errorf("Called with ID %s instead of %s", mockRepo.UpdateItemQuantityID, row.id)
		}

		if mockRepo.UpdateItemQuantityValue != row.quantity {
			t.Errorf(
				"Called with quantity %v instead of %v",
				mockRepo.UpdateItemQuantityValue, row.quantity,
			)
		}
	}
}

func TestUpdateItemLocation(t *testing.T) {
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
		mockRepo := &MockRepo{}
		err := UpdateItemLocation(mockRepo, row.id, row.locationID)
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
	ids := []string{"id1", "id2", "007"}

	for _, id := range ids {
		repo := &MockRepo{}
		err := DeleteItem(repo, id)
		if err != nil {
			t.Errorf("Returned unexpected error for %s: %+v", id, err)
		}

		if repo.DeleteItemID != id {
			t.Errorf(`Called with wrong id: "%s" instead of "%s"`, repo.DeleteItemID, id)
		}
	}
}
