package main

import "context"

type notifier interface {
	NotifyAboutItems(
		ctx context.Context,
		expiries []itemExpiry,
		comingExpiries []itemExpiry,
		authRepo authenticationRepository,
	) error
}
