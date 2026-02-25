package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/nickelghost/nghttp"
	"github.com/nickelghost/ngtel"
)

var errNoEmailAddressesFound = errors.New("no email addresses found")

type authentication interface {
	Check(ctx context.Context, r *http.Request) error
}

type authenticationRepository interface {
	GetAllEmails(ctx context.Context) ([]string, error)
}

func authMiddleware(next http.Handler, auth authentication) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := auth.Check(r.Context(), r); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusUnauthorized, err, ngtel.GetGCPLogArgs)

			return
		}

		next.ServeHTTP(w, r)
	})
}
