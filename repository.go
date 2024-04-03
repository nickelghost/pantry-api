package main

import "context"

type repository interface {
	GetLocations(ctx context.Context, ids *[]string) ([]location, error)
	CreateLocation(ctx context.Context, name string) error
	UpdateLocation(ctx context.Context, id string, name string) error
	DeleteLocation(ctx context.Context, id string) error
	GetItems(ctx context.Context, tags *[]string, locationIDs *[]string) ([]item, error)
	CreateItem(ctx context.Context, params writeItemParams) error
	UpdateItem(ctx context.Context, id string, params writeItemParams) error
	UpdateItemLocation(ctx context.Context, id string, locationID *string) error
	DeleteItem(ctx context.Context, id string) error
}
