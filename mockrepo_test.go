package main

type MockRepo struct {
	GetLocationsCalls int
	GetLocationsIDs   *[]string
	GetLocationsRes   []Location
	GetLocationsErr   error

	CreateLocationCalls int
	CreateLocationName  string

	UpdateLocationCalls int
	UpdateLocationID    string
	UpdateLocationName  string

	DeleteLocationCalls int
	DeleteLocationID    string

	GetItemsCalls       int
	GetItemsSearch      *string
	GetItemsTags        *[]string
	GetItemsLocationIDs *[]string
	GetItemsRes         []Item
	GetItemsErr         error

	CreateItemCalls  int
	CreateItemParams WriteItemParams

	UpdateItemCalls  int
	UpdateItemID     string
	UpdateItemParams WriteItemParams

	UpdateItemQuantityCalls int
	UpdateItemQuantityID    string
	UpdateItemQuantityValue *int

	UpdateItemLocationCalls int
	UpdateItemLocationID    string
	UpdateItemLocationValue *string

	DeleteItemCalls int
	DeleteItemID    string
}

func (repo *MockRepo) GetLocations(ids *[]string) ([]Location, error) {
	repo.GetLocationsCalls += 1
	repo.GetLocationsIDs = ids

	return repo.GetLocationsRes, repo.GetLocationsErr
}

func (repo *MockRepo) CreateLocation(name string) error {
	repo.CreateLocationCalls += 1
	repo.CreateLocationName = name

	return nil
}

func (repo *MockRepo) UpdateLocation(id string, name string) error {
	repo.UpdateLocationCalls += 1
	repo.UpdateLocationID = id
	repo.UpdateLocationName = name

	return nil
}

func (repo *MockRepo) DeleteLocation(id string) error {
	repo.DeleteLocationCalls += 1
	repo.DeleteLocationID = id

	return nil
}

func (repo *MockRepo) GetItems(
	search *string,
	tags *[]string,
	locationIDs *[]string,
) ([]Item, error) {
	repo.GetItemsCalls += 1
	repo.GetItemsSearch = search
	repo.GetItemsTags = tags
	repo.GetItemsLocationIDs = locationIDs

	return repo.GetItemsRes, repo.GetItemsErr
}

func (repo *MockRepo) CreateItem(params WriteItemParams) error {
	repo.CreateItemCalls += 1
	repo.CreateItemParams = params

	return nil
}

func (repo *MockRepo) UpdateItem(id string, params WriteItemParams) error {
	repo.UpdateItemCalls += 1
	repo.UpdateItemID = id
	repo.UpdateItemParams = params

	return nil
}

func (repo *MockRepo) UpdateItemQuantity(id string, quantity *int) error {
	repo.UpdateItemQuantityCalls += 1
	repo.UpdateItemQuantityID = id
	repo.UpdateItemQuantityValue = quantity

	return nil
}

func (repo *MockRepo) UpdateItemLocation(id string, locationID *string) error {
	repo.UpdateItemLocationCalls += 1
	repo.UpdateItemLocationID = id
	repo.UpdateItemLocationValue = locationID

	return nil
}

func (repo *MockRepo) DeleteItem(id string) error {
	repo.DeleteItemCalls += 1
	repo.DeleteItemID = id

	return nil
}
