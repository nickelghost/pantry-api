package main

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type firestoreRepository struct {
	client *firestore.Client
	tracer trace.Tracer
}

func firestoreToLocations(iter *firestore.DocumentIterator) ([]location, error) {
	locations := []location{}

	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("firestore to locations next: %w", err)
		}

		l := location{ID: doc.Ref.ID}
		if err := doc.DataTo(&l); err != nil {
			return nil, fmt.Errorf("firestore to location: %w", err)
		}

		locations = append(locations, l)
	}

	return locations, nil
}

func firestoreToItems(iter *firestore.DocumentIterator) ([]item, error) {
	items := []item{}

	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("firestore to items next: %w", err)
		}

		i := item{ID: doc.Ref.ID}
		if err := doc.DataTo(&i); err != nil {
			return nil, fmt.Errorf("firestore to item: %w", err)
		}

		items = append(items, i)
	}

	return items, nil
}

func (repo firestoreRepository) GetLocations(ctx context.Context, ids *[]string) ([]location, error) {
	ctx, span := repo.tracer.Start(ctx, "firestoreRepository.GetLocations")
	defer span.End()

	q := repo.client.Collection("locations").Query

	if ids != nil {
		q = q.Where("id", "in", ids)
	}

	iter := q.Documents(ctx)

	return firestoreToLocations(iter)
}

func (repo firestoreRepository) CreateLocation(ctx context.Context, name string) error {
	id := uuid.NewString()

	_, err := repo.client.
		Collection("locations").
		Doc(id).
		Set(ctx, map[string]any{
			"name": name,
		})
	if err != nil {
		return fmt.Errorf("firestore create location: %w", err)
	}

	return nil
}

func (repo firestoreRepository) UpdateLocation(ctx context.Context, id string, name string) error {
	_, err := repo.client.
		Collection("locations").
		Doc(id).
		Update(ctx, []firestore.Update{{
			Path:  "name",
			Value: name,
		}})
	if err != nil {
		return fmt.Errorf("firestore update location: %w", err)
	}

	return nil
}

func (repo firestoreRepository) DeleteLocation(ctx context.Context, id string) error {
	itemsIter := repo.client.
		Collection("items").
		Where("locationId", "==", id).
		Documents(ctx)

	err := repo.client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		for {
			doc, err := itemsIter.Next()
			if errors.Is(err, iterator.Done) {
				break
			} else if err != nil {
				return fmt.Errorf("firestore get items: %w", err)
			}

			err = tx.Update(doc.Ref, []firestore.Update{{
				Path:  "locationId",
				Value: nil,
			}})
			if err != nil {
				return fmt.Errorf("firestore nullify item location: %w", err)
			}
		}

		err := tx.Delete(repo.client.Collection("locations").Doc(id))
		if err != nil {
			return fmt.Errorf("firestore delete location: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("firestore transaction: %w", err)
	}

	return nil
}

func (repo firestoreRepository) GetItems(ctx context.Context, tags *[]string, locationIDs *[]string) ([]item, error) {
	ctx, span := repo.tracer.Start(ctx, "firestoreRepository.GetItems")
	defer span.End()

	q := repo.client.Collection("items").Query

	if tags != nil {
		q = q.Where("tags", "array-contains", "tag")
	}

	if locationIDs != nil {
		q = q.Where("locationId", "in", locationIDs)
	}

	iter := q.Documents(ctx)

	return firestoreToItems(iter)
}

func (repo firestoreRepository) CreateItem(ctx context.Context, params writeItemParams) error {
	id := uuid.NewString()

	_, err := repo.client.
		Collection("items").
		Doc(id).
		Set(ctx, params)
	if err != nil {
		return fmt.Errorf("firestore create location: %w", err)
	}

	return nil
}

func (repo firestoreRepository) UpdateItem(ctx context.Context, id string, params writeItemParams) error {
	doc := repo.client.Collection("items").Doc(id)

	_, err := doc.Get(ctx)
	if status.Code(err) == codes.NotFound {
		return fmt.Errorf("item not found: %w", err)
	} else if err != nil {
		return fmt.Errorf("firestore get item: %w", err)
	}

	_, err = doc.Set(ctx, params)
	if err != nil {
		return fmt.Errorf("firestore update location: %w", err)
	}

	return nil
}

func (repo firestoreRepository) UpdateItemLocation(ctx context.Context, id string, locationID *string) error {
	_, err := repo.client.
		Collection("items").
		Doc(id).
		Update(ctx, []firestore.Update{{
			Path:  "locationId",
			Value: locationID,
		}})
	if err != nil {
		return fmt.Errorf("firestore update item location: %w", err)
	}

	return nil
}

func (repo firestoreRepository) DeleteItem(ctx context.Context, id string) error {
	_, err := repo.client.
		Collection("items").
		Doc(id).
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("firestore delete item: %w", err)
	}

	return nil
}
