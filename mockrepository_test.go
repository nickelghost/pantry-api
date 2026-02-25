package main

import "context"

type mockRepository struct {
	GetLocationsCalls int
	GetLocationsIDs   *[]string
	GetLocationsRes   []location
	GetLocationsErr   error

	CreateLocationCalls int
	CreateLocationName  string

	UpdateLocationCalls int
	UpdateLocationID    string
	UpdateLocationName  string

	DeleteLocationCalls int
	DeleteLocationID    string

	GetItemsCalls       int
	GetItemsTags        *[]string
	GetItemsLocationIDs *[]string
	GetItemsRes         []item
	GetItemsErr         error

	CreateItemCalls  int
	CreateItemParams writeItemParams

	UpdateItemCalls  int
	UpdateItemID     string
	UpdateItemParams writeItemParams

	UpdateItemLocationCalls int
	UpdateItemLocationID    string
	UpdateItemLocationValue *string

	DeleteItemCalls int
	DeleteItemID    string
}

func (repo *mockRepository) GetLocations(_ context.Context, ids *[]string) ([]location, error) {
	repo.GetLocationsCalls++
	repo.GetLocationsIDs = ids

	return repo.GetLocationsRes, repo.GetLocationsErr
}

func (repo *mockRepository) CreateLocation(_ context.Context, name string) error {
	repo.CreateLocationCalls++
	repo.CreateLocationName = name

	return nil
}

func (repo *mockRepository) UpdateLocation(_ context.Context, id string, name string) error {
	repo.UpdateLocationCalls++
	repo.UpdateLocationID = id
	repo.UpdateLocationName = name

	return nil
}

func (repo *mockRepository) DeleteLocation(_ context.Context, id string) error {
	repo.DeleteLocationCalls++
	repo.DeleteLocationID = id

	return nil
}

func (repo *mockRepository) GetItems(_ context.Context,
	tags *[]string,
	locationIDs *[]string,
) ([]item, error) {
	repo.GetItemsCalls++
	repo.GetItemsTags = tags
	repo.GetItemsLocationIDs = locationIDs

	return repo.GetItemsRes, repo.GetItemsErr
}

func (repo *mockRepository) CreateItem(_ context.Context, params writeItemParams) error {
	repo.CreateItemCalls++
	repo.CreateItemParams = params

	return nil
}

func (repo *mockRepository) UpdateItem(_ context.Context, id string, params writeItemParams) error {
	repo.UpdateItemCalls++
	repo.UpdateItemID = id
	repo.UpdateItemParams = params

	return nil
}

func (repo *mockRepository) UpdateItemLocation(_ context.Context, id string, locationID *string) error {
	repo.UpdateItemLocationCalls++
	repo.UpdateItemLocationID = id
	repo.UpdateItemLocationValue = locationID

	return nil
}

func (repo *mockRepository) DeleteItem(_ context.Context, id string) error {
	repo.DeleteItemCalls++
	repo.DeleteItemID = id

	return nil
}
