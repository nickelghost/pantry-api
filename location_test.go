package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestGetLocationsParams(t *testing.T) {
	search := getPtr("oran")
	tags := getPtr([]string{"fruit", "sweet"})

	mockRepo := &MockRepo{}
	_, _, err := GetLocations(mockRepo, search, tags)
	if err != nil {
		t.Errorf("Got error: %+v", err)
	}

	if mockRepo.GetLocationsCalls != 1 {
		t.Errorf("Called GetLocations %d times instead of once", mockRepo.GetLocationsCalls)
	}

	if mockRepo.GetItemsCalls != 1 {
		t.Errorf("Called GetItems %d times instead of once", mockRepo.GetLocationsCalls)
	}

	if mockRepo.GetItemsSearch == nil {
		t.Error("Did not call GetItems with search")
	} else if mockRepo.GetItemsSearch != search {
		t.Errorf(`Called GetItems with %s search instead of %s`, *mockRepo.GetItemsSearch, *search)
	}

	if mockRepo.GetItemsTags == nil {
		t.Error("Did not call GetItems with tags")
	} else if mockRepo.GetItemsTags != tags {
		t.Errorf("Called GetItems with %v tags instead of %v tags", *mockRepo.GetItemsTags, *tags)
	}
}

func TestGetLocationsRes(t *testing.T) {
	locations := []Location{
		{ID: "pantry", Name: "Pantry"},
		{ID: "fridge", Name: "Fridge"},
	}
	items := []Item{
		{Name: "Potato", LocationID: nil},
		{Name: "Cheese", LocationID: getPtr("fridge")},
		{Name: "Milk", LocationID: getPtr("fridge")},
	}

	mockRepo := &MockRepo{GetLocationsRes: locations, GetItemsRes: items}

	filledLocations, remainingItems, err := GetLocations(mockRepo, nil, nil)
	if err != nil {
		t.Errorf("Got error: %+v", err)
	}

	// id: pantry
	if len(filledLocations[0].Items) != 0 {
		t.Errorf(
			"Location without any items (%s) got filled with %+v",
			filledLocations[0].Name,
			filledLocations[0].Items,
		)
	}

	// id: fridge
	if len(filledLocations[1].Items) != 2 ||
		filledLocations[1].Items[0].Name != "Cheese" ||
		filledLocations[1].Items[1].Name != "Milk" {
		t.Errorf(
			"Instead of Cheese and Milk in the Fridge, the %s contains %+v",
			filledLocations[1].Name,
			filledLocations[1].Items,
		)
	}

	if len(remainingItems) != 1 {
		t.Errorf(
			"Wrong number of remaining items, expected 1 but got %d with %+v",
			len(remainingItems),
			remainingItems,
		)
	} else if remainingItems[0].Name != "Potato" {
		t.Errorf("Remaining items does not contain Potato, instead contains %+v", remainingItems)
	}
}

func TestGetLocationsErrs(t *testing.T) {
	errScenarios := []struct {
		repo *MockRepo
		err  error
	}{
		{
			repo: &MockRepo{GetLocationsErr: errors.New("get locations failed")},
			err:  errors.New("get locations failed"),
		},
		{
			repo: &MockRepo{GetItemsErr: errors.New("get items failed")},
			err:  errors.New("get items failed"),
		},
	}

	for _, s := range errScenarios {
		_, _, err := GetLocations(s.repo, nil, nil)
		if err.Error() != s.err.Error() {
			t.Errorf("Expected %v but got %v", s.err, err)
		}
	}
}

func TestGetLocation(t *testing.T) {
	location := Location{ID: "freezer", Name: "Freezer"}
	mockRepo := &MockRepo{
		GetLocationsRes: []Location{location},
	}
	search := getPtr("fried")
	tags := getPtr([]string{"microwave", "oven"})

	res, err := GetLocation(mockRepo, location.ID, search, tags)
	if err != nil {
		t.Errorf("Got error: %s", err)
	}

	if !reflect.DeepEqual(res, location) {
		t.Errorf("Returned %+v instead of %+v", res, location)
	}

	if mockRepo.GetLocationsCalls != 1 {
		t.Errorf("Called GetLocations %d times instead of once", mockRepo.GetLocationsCalls)
	}

	if mockRepo.GetLocationsIDs == nil {
		t.Error("Did not call GetLocations with loc IDs")
	} else if !reflect.DeepEqual(*mockRepo.GetLocationsIDs, []string{location.ID}) {
		t.Errorf("Called GetLocations with ids %+v instead of %s", *mockRepo.GetLocationsIDs, location.ID)
	}

	if mockRepo.GetItemsCalls != 1 {
		t.Errorf("Called GetItems %d times instead of once", mockRepo.GetItemsCalls)
	}

	if mockRepo.GetItemsSearch == nil {
		t.Error("Did not call GetItems with search")
	} else if mockRepo.GetItemsSearch != search {
		t.Errorf(`Called GetItems with %s search instead of %s`, *mockRepo.GetItemsSearch, *search)
	}

	if mockRepo.GetItemsTags == nil {
		t.Error("Did not call GetItems with tags")
	} else if mockRepo.GetItemsTags != tags {
		t.Errorf("Called GetItems with %v tags instead of %v tags", *mockRepo.GetItemsTags, *tags)
	}

	if mockRepo.GetItemsLocationIDs == nil {
		t.Error("Did not call GetItems with loc IDs")
	} else if !reflect.DeepEqual(*mockRepo.GetItemsLocationIDs, []string{location.ID}) {
		t.Errorf("Called GetItems with %v loc IDs instead of %s", *mockRepo.GetItemsLocationIDs, location.ID)
	}
}

func TestLocationErrs(t *testing.T) {
	errScenarios := []struct {
		repo *MockRepo
		err  error
	}{
		{
			repo: &MockRepo{},
			err:  errors.New("not found"),
		},
		{
			repo: &MockRepo{GetLocationsErr: errors.New("get locations failed")},
			err:  errors.New("get locations failed"),
		},
	}

	for _, s := range errScenarios {
		_, err := GetLocation(s.repo, uuid.NewString(), nil, nil)
		if err.Error() != s.err.Error() {
			t.Errorf("Expected %v but got %v", s.err, err)
		}
	}
}

func TestCreateLocation(t *testing.T) {
	correctNames := []string{
		"Fridge",
		"Childrens' Pantry",
		"@ the top of the shelf",
		"Привет как ты мой друг? Я скучал по тебе. Я давно",
	}

	for _, cn := range correctNames {
		repo := &MockRepo{}
		err := CreateLocation(repo, cn)
		if err != nil {
			t.Errorf("Returned unexpected error for %s: %+v", cn, err)
		}

		if repo.CreateLocationCalls != 1 {
			t.Errorf(
				`Called repo wrong number of times: %d instead of 1 on "%s"`,
				repo.CreateLocationCalls,
				cn,
			)
		}

		if repo.CreateLocationName != cn {
			t.Errorf(`Called repo with wrong argument %+v instead of %+v`, repo.CreateLocationName, cn)
		}
	}

	incorrectNames := []string{
		"",
		"Despite my best efforts, I could not replicate such a complex environment",
		"Привет как ты мой друг? Я скучал по тебе. Я давно не писал.",
	}

	for _, in := range incorrectNames {
		repo := &MockRepo{}
		err := CreateLocation(repo, in)

		if err == nil {
			t.Errorf("Did not return error on %s", in)
		}

		if repo.CreateLocationCalls > 0 {
			t.Errorf("Called repo %d number of times instead of none on %s", repo.CreateLocationCalls, in)
		}
	}
}

func TestUpdateLocation(t *testing.T) {
	correctNames := []string{
		"Childrens' Pantry",
		"@ the top of the shelf",
		"Привет как ты мой друг? Я скучал по тебе. Я давно",
	}

	for _, cn := range correctNames {
		repo := &MockRepo{}
		id := uuid.New().String()
		err := UpdateLocation(repo, id, cn)
		if err != nil {
			t.Errorf("Returned unexpected error for %s: %+v", cn, err)
		}

		if repo.UpdateLocationCalls != 1 {
			t.Errorf(
				`Called repo wrong number of times: %d instead of 1 on "%s"`,
				repo.UpdateLocationCalls,
				cn,
			)
		}

		if repo.UpdateLocationID != id {
			t.Errorf(`Called repo with wrong argument %+v instead of %+v`, repo.UpdateLocationID, id)
		}

		if repo.UpdateLocationName != cn {
			t.Errorf(`Called repo with wrong argument %+v instead of %+v`, repo.UpdateLocationName, cn)
		}
	}

	incorrectNames := []string{
		"",
		"Despite my best efforts, I could not replicate such a complex environment",
		"Привет как ты мой друг? Я скучал по тебе. Я давно не писал.",
	}

	for _, in := range incorrectNames {
		repo := &MockRepo{}
		err := UpdateLocation(repo, "id", in)

		if err == nil {
			t.Errorf("Did not return error on %s", in)
		}

		if repo.UpdateLocationCalls > 0 {
			t.Errorf("Called repo %d number of times instead of none on %s", repo.UpdateLocationCalls, in)
		}
	}
}

func TestDeleteLocation(t *testing.T) {
	ids := []string{"id1", "id2", "007"}

	for _, id := range ids {
		repo := &MockRepo{}
		err := DeleteLocation(repo, id)
		if err != nil {
			t.Errorf("Returned unexpected error for %s: %+v", id, err)
		}

		if repo.DeleteLocationID != id {
			t.Errorf(`Called with wrong id: "%s" instead of "%s"`, repo.DeleteLocationID, id)
		}
	}
}
