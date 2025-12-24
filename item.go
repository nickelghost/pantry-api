package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/go-playground/validator/v10"
)

type item struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Type       *string    `json:"type"`
	Tags       []string   `json:"tags"`
	Price      *int       `json:"price"`
	BoughtAt   time.Time  `json:"boughtAt"`
	OpenedAt   *time.Time `json:"openedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	Lifespan   *int       `json:"lifespan"`
	LocationID *string    `json:"locationId"`
	Location   *location  `json:"location,omitempty"`
}

type itemExpiry struct {
	item     item
	daysLeft int
}

const expiresSoonThreshold = 2

type writeItemParams struct {
	Name       string     `json:"name"       validate:"required,min=2"`
	Type       *string    `json:"type"`
	Tags       []string   `json:"tags"       validate:"required"`
	Price      *int       `json:"price"      validate:"omitempty,gte=0"`
	BoughtAt   time.Time  `json:"boughtAt"   validate:"required"`
	OpenedAt   *time.Time `json:"openedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	Lifespan   *int       `json:"lifespan"   validate:"omitempty,gte=0"`
	LocationID *string    `json:"locationId"`
}

func getItemDaysLeft(item item) *int {
	daysOpts := []int{}

	// if has an expiry date, add number of remaining days to opts
	if item.ExpiresAt != nil {
		daysOpt := int(math.Ceil(time.Until(*item.ExpiresAt).Hours() / 24)) //nolint:mnd
		daysOpts = append(daysOpts, daysOpt)
	}

	// if was opened and has lifespan, set the remaining lifetime days to opts
	if item.OpenedAt != nil && item.Lifespan != nil {
		lifespanHours := time.Duration(*item.Lifespan) * 24 * time.Hour                      //nolint:mnd
		daysOpt := int(math.Ceil(time.Until(item.OpenedAt.Add(lifespanHours)).Hours() / 24)) //nolint:mnd
		daysOpts = append(daysOpts, daysOpt)
	}

	if len(daysOpts) == 0 {
		return nil
	}

	var daysLeft int

	// find the smallest possible expiry and set it
	for i, do := range daysOpts {
		if i == 0 || do < daysLeft {
			daysLeft = do
		}
	}

	return &daysLeft
}

func notifyAboutItems(ctx context.Context, repo repository, n notifier, authRepo authenticationRepository) error {
	items, err := repo.GetItems(ctx, nil, nil)
	if err != nil {
		return fmt.Errorf("get items: %w", err)
	}

	expiries, comingExpiries := []itemExpiry{}, []itemExpiry{}

	for _, item := range items {
		daysLeft := getItemDaysLeft(item)

		if daysLeft == nil {
			continue
		}

		expiry := itemExpiry{item, *daysLeft}

		// we only want to notify about items that are expired or are soon to be expired
		if *daysLeft < 0 {
			expiries = append(expiries, expiry)
		} else if *daysLeft <= expiresSoonThreshold {
			comingExpiries = append(comingExpiries, expiry)
		}
	}

	if len(expiries) == 0 && len(comingExpiries) == 0 {
		slog.Info("No items expired nor expired soon, skipping the notification.")

		return nil
	}

	if err := n.NotifyAboutItems(ctx, expiries, comingExpiries, authRepo); err != nil {
		return fmt.Errorf("notify about items: %w", err)
	}

	return nil
}

func createItem(ctx context.Context, repo repository, validate *validator.Validate, params writeItemParams) error {
	if err := validate.Struct(params); err != nil {
		return fmt.Errorf("invalid write item params: %w", err)
	}

	if err := repo.CreateItem(ctx, params); err != nil {
		return fmt.Errorf("create item: %w", err)
	}

	return nil
}

func updateItem(
	ctx context.Context,
	repo repository,
	validate *validator.Validate,
	id string,
	params writeItemParams,
) error {
	if err := validate.Struct(params); err != nil {
		return fmt.Errorf("invalid write item params: %w", err)
	}

	if err := repo.UpdateItem(ctx, id, params); err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

func updateItemLocation(ctx context.Context, repo repository, id string, locationID *string) error {
	if err := repo.UpdateItemLocation(ctx, id, locationID); err != nil {
		return fmt.Errorf("update item location: %w", err)
	}

	return nil
}

func deleteItem(ctx context.Context, repo repository, id string) error {
	if err := repo.DeleteItem(ctx, id); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}
