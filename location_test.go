package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func TestGetLocationsParams(t *testing.T) {
	t.Parallel()

	tags := getPtr([]string{"fruit", "sweet"})

	mockRepo := &mockRepository{}

	_, _, err := getLocations(mockRepo, tags)
	if err != nil {
		t.Errorf("Got error: %+v", err)
	}

	if mockRepo.GetLocationsCalls != 1 {
		t.Errorf("Called GetLocations %d times instead of once", mockRepo.GetLocationsCalls)
	}

	if mockRepo.GetItemsCalls != 1 {
		t.Errorf("Called GetItems %d times instead of once", mockRepo.GetLocationsCalls)
	}

	if mockRepo.GetItemsTags == nil {
		t.Error("Did not call GetItems with tags")
	} else if mockRepo.GetItemsTags != tags {
		t.Errorf("Called GetItems with %v tags instead of %v tags", *mockRepo.GetItemsTags, *tags)
	}
}

func TestGetLocationsRes(t *testing.T) {
	t.Parallel()

	locations := []location{
		{ID: "pantry", Name: "Pantry"},
		{ID: "fridge", Name: "Fridge"},
	}
	items := []item{
		{Name: "Potato", LocationID: nil},
		{Name: "Cheese", LocationID: getPtr("fridge")},
		{Name: "Milk", LocationID: getPtr("fridge")},
	}

	mockRepo := &mockRepository{GetLocationsRes: locations, GetItemsRes: items}

	filledLocations, remainingItems, err := getLocations(mockRepo, nil)
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
	t.Parallel()

	errScenarios := []struct {
		repo   *mockRepository
		errStr string
	}{
		{
			//nolint:goerr113
			repo:   &mockRepository{GetLocationsErr: errors.New("get locations failed")},
			errStr: "get locations failed",
		},
		{
			//nolint:goerr113
			repo:   &mockRepository{GetItemsErr: errors.New("get items failed")},
			errStr: "get items failed",
		},
	}

	for _, s := range errScenarios {
		_, _, err := getLocations(s.repo, nil)
		if !strings.Contains(err.Error(), s.errStr) {
			t.Errorf(`Expected "%s" to contain "%s"`, err, s.errStr)
		}
	}
}

//nolint:cyclop
func TestGetLocation(t *testing.T) {
	t.Parallel()

	l := location{ID: "freezer", Name: "Freezer"}
	mockRepo := &mockRepository{
		GetLocationsRes: []location{l},
	}
	tags := getPtr([]string{"microwave", "oven"})

	res, err := getLocation(mockRepo, l.ID, tags)
	if err != nil {
		t.Errorf("Got error: %s", err)
	}

	if !reflect.DeepEqual(res, l) {
		t.Errorf("Returned %+v instead of %+v", res, l)
	}

	if mockRepo.GetLocationsCalls != 1 {
		t.Errorf("Called GetLocations %d times instead of once", mockRepo.GetLocationsCalls)
	}

	if mockRepo.GetLocationsIDs == nil {
		t.Error("Did not call GetLocations with loc IDs")
	} else if !reflect.DeepEqual(*mockRepo.GetLocationsIDs, []string{l.ID}) {
		t.Errorf("Called GetLocations with ids %+v instead of %s", *mockRepo.GetLocationsIDs, l.ID)
	}

	if mockRepo.GetItemsCalls != 1 {
		t.Errorf("Called GetItems %d times instead of once", mockRepo.GetItemsCalls)
	}

	if mockRepo.GetItemsTags == nil {
		t.Error("Did not call GetItems with tags")
	} else if mockRepo.GetItemsTags != tags {
		t.Errorf("Called GetItems with %v tags instead of %v tags", *mockRepo.GetItemsTags, *tags)
	}

	if mockRepo.GetItemsLocationIDs == nil {
		t.Error("Did not call GetItems with loc IDs")
	} else if !reflect.DeepEqual(*mockRepo.GetItemsLocationIDs, []string{l.ID}) {
		t.Errorf("Called GetItems with %v loc IDs instead of %s", *mockRepo.GetItemsLocationIDs, l.ID)
	}
}

func TestLocationErrs(t *testing.T) {
	t.Parallel()

	errScenarios := []struct {
		repo   *mockRepository
		errStr string
	}{
		{
			repo:   &mockRepository{},
			errStr: "location not found",
		},
		{
			//nolint:goerr113
			repo:   &mockRepository{GetLocationsErr: errors.New("get locations failed")},
			errStr: "get locations failed",
		},
	}

	for _, s := range errScenarios {
		_, err := getLocation(s.repo, uuid.NewString(), nil)
		if !strings.Contains(err.Error(), s.errStr) {
			t.Errorf(`Expected "%s" to contain "%s"`, err, s.errStr)
		}
	}
}

func TestCreateLocation(t *testing.T) {
	t.Parallel()

	validate := validator.New(validator.WithRequiredStructEnabled())

	correctNames := []string{
		"Fridge",
		"Childrens' Pantry",
		"@ the top of the shelf",
		"Привет как ты мой друг? Я скучал по тебе. Я давно",
	}

	for _, cn := range correctNames {
		repo := &mockRepository{}

		err := createLocation(repo, validate, cn)
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
		repo := &mockRepository{}
		err := createLocation(repo, validate, in)

		if err == nil {
			t.Errorf("Did not return error on %s", in)
		}

		if repo.CreateLocationCalls > 0 {
			t.Errorf("Called repo %d number of times instead of none on %s", repo.CreateLocationCalls, in)
		}
	}
}

func TestUpdateLocation(t *testing.T) {
	t.Parallel()

	validate := validator.New(validator.WithRequiredStructEnabled())

	correctNames := []string{
		"Childrens' Pantry",
		"@ the top of the shelf",
		"Привет как ты мой друг? Я скучал по тебе. Я давно",
	}

	for _, cn := range correctNames {
		repo := &mockRepository{}
		id := uuid.New().String()

		err := updateLocation(repo, validate, id, cn)
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
		repo := &mockRepository{}
		err := updateLocation(repo, validate, "id", in)

		if err == nil {
			t.Errorf("Did not return error on %s", in)
		}

		if repo.UpdateLocationCalls > 0 {
			t.Errorf("Called repo %d number of times instead of none on %s", repo.UpdateLocationCalls, in)
		}
	}
}

func TestDeleteLocation(t *testing.T) {
	t.Parallel()

	ids := []string{"id1", "id2", "007"}

	for _, id := range ids {
		repo := &mockRepository{}

		err := deleteLocation(repo, id)
		if err != nil {
			t.Errorf("Returned unexpected error for %s: %+v", id, err)
		}

		if repo.DeleteLocationID != id {
			t.Errorf(`Called with wrong id: "%s" instead of "%s"`, repo.DeleteLocationID, id)
		}
	}
}
