package main

type repository interface {
	GetLocations(ids *[]string) ([]location, error)
	CreateLocation(name string) error
	UpdateLocation(id string, name string) error
	DeleteLocation(id string) error
	GetItems(
		tags *[]string,
		locationIDs *[]string,
	) ([]item, error)
	CreateItem(params writeItemParams) error
	UpdateItem(id string, params writeItemParams) error
	UpdateItemLocation(id string, locationID *string) error
	DeleteItem(id string) error
}
