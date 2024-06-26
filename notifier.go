package main

import (
	"context"
	"fmt"
	"math"
	"time"
)

type notifier interface {
	NotifyAboutItems(
		ctx context.Context,
		expiries []itemExpiry,
		comingExpiries []itemExpiry,
		authRepo authenticationRepository,
	) error
}

func getNotificationTitle() string {
	return "Pantry - Expiring items on " + time.Now().Format(time.DateOnly)
}

func notificationExpiriesToText(expiries []itemExpiry, comingExpiries []itemExpiry) string {
	text := ""

	if len(expiries) > 0 {
		text += "EXPIRED ITEMS\n-------------\n"

		for _, exp := range expiries {
			// makes the int positive
			daysOverdue := int(math.Abs(float64(exp.daysLeft)))

			text += fmt.Sprintf("%s is %d day(s) overdue\n", exp.item.Name, daysOverdue)
		}
	}

	if len(comingExpiries) > 0 {
		if len(expiries) > 0 {
			text += "\n"
		}

		text += "ITEMS ABOUT TO EXPIRE\n---------------------\n"

		for _, exp := range comingExpiries {
			text += fmt.Sprintf("%s has %d day(s) left\n", exp.item.Name, exp.daysLeft)
		}
	}

	if text == "" {
		text = "NO EXPIRING ITEMS"
	}

	return text
}
