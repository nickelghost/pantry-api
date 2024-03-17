package main

type Repo interface {
	GetLocations(ids *[]string) ([]Location, error)
	CreateLocation(name string) error
	UpdateLocation(id string, name string) error
	DeleteLocation(id string) error
	GetItems(
		search *string,
		tags *[]string,
		locationIDs *[]string,
	) ([]Item, error)
	CreateItem(params WriteItemParams) error
	UpdateItem(id string, params WriteItemParams) error
	UpdateItemQuantity(id string, quantity *int) error
	UpdateItemLocation(id string, locationID *string) error
	DeleteItem(id string) error
}
