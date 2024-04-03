package main

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type location struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Items []item `json:"items,omitempty"`
}

func (location) GetNameConstraints() string {
	return "required,min=1,max=50"
}

func fillLocations(locations []location, items []item) ([]location, []item) {
	remainingItems := []item{}

	for _, i := range items {
		if i.LocationID == nil {
			remainingItems = append(remainingItems, i)

			continue
		}

		for li, loc := range locations {
			if loc.ID == *i.LocationID {
				// add item to the location
				locations[li].Items = append(locations[li].Items, i)
			}
		}
	}

	return locations, remainingItems
}

func getLocationsCommon(repo repository, ids *[]string, tags *[]string) ([]location, []item, error) {
	locs, err := repo.GetLocations(ids)
	if err != nil {
		return nil, nil, fmt.Errorf("get locations: %w", err)
	}

	items, err := repo.GetItems(tags, ids)
	if err != nil {
		return nil, nil, fmt.Errorf("get items: %w", err)
	}

	filledLocs, remainingItems := fillLocations(locs, items)

	return filledLocs, remainingItems, nil
}

func getLocations(repo repository, tags *[]string) ([]location, []item, error) {
	return getLocationsCommon(repo, nil, tags)
}

var errLocationNotFound = errors.New("location not found")

func getLocation(repo repository, id string, tags *[]string) (location, error) {
	locations, _, err := getLocationsCommon(repo, getPtr([]string{id}), tags)
	if err != nil {
		return location{}, err
	}

	if len(locations) == 0 {
		return location{}, errLocationNotFound
	}

	return locations[0], nil
}

func createLocation(repo repository, validate *validator.Validate, name string) error {
	if err := validate.Var(name, location{}.GetNameConstraints()); err != nil {
		return fmt.Errorf("name validation: %w", err)
	}

	if err := repo.CreateLocation(name); err != nil {
		return fmt.Errorf("create location: %w", err)
	}

	return nil
}

func updateLocation(repo repository, validate *validator.Validate, id string, name string) error {
	if err := validate.Var(name, location{}.GetNameConstraints()); err != nil {
		return fmt.Errorf("name validation: %w", err)
	}

	if err := repo.UpdateLocation(id, name); err != nil {
		return fmt.Errorf("update location: %w", err)
	}

	return nil
}

func deleteLocation(repo repository, id string) error {
	if err := repo.DeleteLocation(id); err != nil {
		return fmt.Errorf("delete location: %w", err)
	}

	return nil
}
