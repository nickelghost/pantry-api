package main

import (
	"context"
	"fmt"
)

type terminalNotifier struct{}

func (n terminalNotifier) NotifyAboutItems(
	_ context.Context, expiries []itemExpiry, comingExpiries []itemExpiry, _ authenticationRepository,
) error {
	fmt.Print(notificationExpiriesToText(expiries, comingExpiries)) //nolint: forbidigo

	return nil
}
