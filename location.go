package main

import (
	"errors"
)

type Location struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Items []Item `json:"items,omitempty"`
}

var locationNameConstraints = "required,min=1,max=50"

func fillLocations(locations []Location, items []Item) ([]Location, []Item) {
	remainingItems := []Item{}
	for _, item := range items {
		if item.LocationID == nil {
			remainingItems = append(remainingItems, item)

			continue
		}

		for li, loc := range locations {
			if loc.ID == *item.LocationID {
				// add item to the location
				locations[li].Items = append(locations[li].Items, item)
			}
		}
	}

	return locations, remainingItems
}

func getLocations(repo Repo, ids *[]string, search *string, tags *[]string) ([]Location, []Item, error) {
	locs, err := repo.GetLocations(ids)
	if err != nil {
		return nil, nil, err
	}

	items, err := repo.GetItems(search, tags, ids)
	if err != nil {
		return nil, nil, err
	}

	filledLocs, remainingItems := fillLocations(locs, items)

	return filledLocs, remainingItems, nil
}

func GetLocations(repo Repo, search *string, tags *[]string) ([]Location, []Item, error) {
	return getLocations(repo, nil, search, tags)
}

func GetLocation(repo Repo, id string, search *string, tags *[]string) (Location, error) {
	locations, _, err := getLocations(repo, getPtr([]string{id}), search, tags)
	if err != nil {
		return Location{}, err
	}

	if len(locations) == 0 {
		return Location{}, errors.New("not found")
	}

	return locations[0], nil
}

func CreateLocation(repo Repo, name string) error {
	err := validate.Var(name, locationNameConstraints)
	if err != nil {
		return err
	}

	return repo.CreateLocation(name)
}

func UpdateLocation(repo Repo, id string, name string) error {
	err := validate.Var(name, locationNameConstraints)
	if err != nil {
		return err
	}

	return repo.UpdateLocation(id, name)
}

func DeleteLocation(repo Repo, id string) error {
	return repo.DeleteLocation(id)
}
