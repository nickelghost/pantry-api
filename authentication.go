package main

import (
	"context"
	"net/http"
)

type authentication interface {
	Check(ctx context.Context, r *http.Request) error
}
