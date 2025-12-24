package main

import (
	"context"
	"fmt"
	"math"
	"strings"
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
	var textBuilder strings.Builder

	if len(expiries) > 0 {
		textBuilder.WriteString("EXPIRED ITEMS\n-------------\n")

		for _, exp := range expiries {
			// makes the int positive
			daysOverdue := int(math.Abs(float64(exp.daysLeft)))

			fmt.Fprintf(&textBuilder, "%s is %d day(s) overdue\n", exp.item.Name, daysOverdue)
		}
	}

	if len(comingExpiries) > 0 {
		if len(expiries) > 0 {
			textBuilder.WriteString("\n")
		}

		textBuilder.WriteString("ITEMS ABOUT TO EXPIRE\n---------------------\n")

		for _, exp := range comingExpiries {
			fmt.Fprintf(&textBuilder, "%s has %d day(s) left\n", exp.item.Name, exp.daysLeft)
		}
	}

	if textBuilder.Len() == 0 {
		textBuilder.WriteString("NO EXPIRING ITEMS")
	}

	return textBuilder.String()
}
