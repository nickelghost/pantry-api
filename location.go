package main

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Location struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Items []Item `json:"items,omitempty"`
}

func (Location) GetNameConstraints() string {
	return "required,min=1,max=50"
}

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
		return nil, nil, fmt.Errorf("get locations: %w", err)
	}

	items, err := repo.GetItems(search, tags, ids)
	if err != nil {
		return nil, nil, fmt.Errorf("get items: %w", err)
	}

	filledLocs, remainingItems := fillLocations(locs, items)

	return filledLocs, remainingItems, nil
}

func GetLocations(repo Repo, search *string, tags *[]string) ([]Location, []Item, error) {
	return getLocations(repo, nil, search, tags)
}

var ErrLocationNotFound = errors.New("location not found")

func GetLocation(repo Repo, id string, search *string, tags *[]string) (Location, error) {
	locations, _, err := getLocations(repo, getPtr([]string{id}), search, tags)
	if err != nil {
		return Location{}, err
	}

	if len(locations) == 0 {
		return Location{}, ErrLocationNotFound
	}

	return locations[0], nil
}

func CreateLocation(repo Repo, validate *validator.Validate, name string) error {
	if err := validate.Var(name, Location{}.GetNameConstraints()); err != nil {
		return fmt.Errorf("name validation: %w", err)
	}

	if err := repo.CreateLocation(name); err != nil {
		return fmt.Errorf("create location: %w", err)
	}

	return nil
}

func UpdateLocation(repo Repo, validate *validator.Validate, id string, name string) error {
	if err := validate.Var(name, Location{}.GetNameConstraints()); err != nil {
		return fmt.Errorf("name validation: %w", err)
	}

	if err := repo.UpdateLocation(id, name); err != nil {
		return fmt.Errorf("update location: %w", err)
	}

	return nil
}

func DeleteLocation(repo Repo, id string) error {
	if err := repo.DeleteLocation(id); err != nil {
		return fmt.Errorf("delete location: %w", err)
	}

	return nil
}
