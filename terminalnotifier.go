package main

import (
	"context"
	"fmt"
	"math"
)

type terminalNotifier struct{}

func (n terminalNotifier) expiriesToText(expiries []itemExpiry) string {
	text := "EXPIRED ITEMS\n-------------\n"

	for _, exp := range expiries {
		// makes the int positive
		daysOverdue := int(math.Abs(float64(exp.daysLeft)))

		text += fmt.Sprintf("%s is %d day(s) overdue\n", exp.item.Name, daysOverdue)
	}

	return text
}

func (n terminalNotifier) comingExpiresToText(comingExpires []itemExpiry) string {
	text := "ITEMS ABOUT TO EXPIRE\n---------------------\n"

	for _, exp := range comingExpires {
		text += fmt.Sprintf("%s has %d day(s) left\n", exp.item.Name, exp.daysLeft)
	}

	return text
}

func (n terminalNotifier) NotifyAboutItems(
	_ context.Context, expiries []itemExpiry, comingExpiries []itemExpiry, _ authenticationRepository,
) error {
	fmt.Print(n.expiriesToText(expiries))            //nolint:forbidigo
	fmt.Print("\n")                                  //nolint:forbidigo
	fmt.Print(n.comingExpiresToText(comingExpiries)) //nolint:forbidigo

	return nil
}
