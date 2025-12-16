package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nickelghost/nghttp"
	"github.com/nickelghost/ngtel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

const httpHeaderTimeout = 1 * time.Second

func getServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              net.JoinHostPort("0.0.0.0", "8080"),
		Handler:           handler,
		ReadHeaderTimeout: httpHeaderTimeout,
	}
}

func getRouter(
	repo repository,
	validate *validator.Validate,
	auth authentication,
) http.Handler {
	mux := http.NewServeMux()

	// todo : add health
	mux.HandleFunc("GET /locations", indexLocationsHandler(repo))
	mux.HandleFunc("GET /locations/{id}", getLocationHandler(repo))
	mux.HandleFunc("POST /locations", createLocationHandler(repo, validate))
	mux.HandleFunc("PUT /locations/{id}", updateLocationHandler(repo, validate))
	mux.HandleFunc("DELETE /locations/{id}", deleteLocationHandler(repo))
	mux.HandleFunc("POST /items", createItemHandler(repo, validate))
	mux.HandleFunc("PUT /items/{id}", updateItemHandler(repo, validate))
	mux.HandleFunc("PATCH /items/{id}/location", updateItemLocationHandler(repo))
	mux.HandleFunc("DELETE /items/{id}", deleteItemHandler(repo))
	mux.HandleFunc("/", nghttp.GetNotFoundHandler(ngtel.GetGCPLogArgs))

	var handler http.Handler = mux

	if auth != nil {
		handler = useAuth(handler, auth)
	}

	handler = nghttp.UseCORS(
		handler,
		strings.Split(os.Getenv("ACCESS_CONTROL_ALLOW_ORIGIN"), ","),
		strings.Split(os.Getenv("ACCESS_CONTROL_ALLOW_HEADERS"), ","),
		[]string{"*"},
		ngtel.GetGCPLogArgs,
	)
	handler = nghttp.UseRequestLogging(handler, ngtel.GetGCPLogArgs)
	handler = nghttp.UseRequestID(handler, "X-Request-ID")
	handler = OtelHTTPMiddleware(handler)

	return handler
}

func OtelHTTPMiddleware(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "request", otelhttp.WithSpanNameFormatter(
		func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		},
	))
}

func SetSpanNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Pattern != "" {
			span := trace.SpanFromContext(r.Context())
			span.SetName(r.Pattern)
		}

		next.ServeHTTP(w, r)
	})
}

func indexLocationsHandler(repo repository) http.HandlerFunc {
	return SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var tags *[]string

		if val := r.URL.Query().Get("tags"); val != "" {
			vals := strings.Split(val, ",")
			tags = &vals
		}

		locs, remItems, err := getLocations(r.Context(), repo, tags)
		if err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		res := struct {
			Locations      []location `json:"locations"`
			RemainingItems []item     `json:"remainingItems"`
		}{Locations: locs, RemainingItems: remItems}

		nghttp.Respond(w, r, http.StatusOK, nil, res, ngtel.GetGCPLogArgs)
	})
}

func getLocationHandler(repo repository) http.HandlerFunc {
	return SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var tags *[]string

		if val := r.URL.Query().Get("tags"); val != "" {
			vals := strings.Split(val, ",")
			tags = &vals
		}

		loc, err := getLocation(r.Context(), repo, id, tags)
		if errors.Is(err, errLocationNotFound) {
			nghttp.RespondGeneric(w, r, http.StatusNotFound, err, ngtel.GetGCPLogArgs)

			return
		} else if err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		res := struct {
			location `json:"location"`
		}{location: loc}

		nghttp.Respond(w, r, http.StatusOK, nil, res, ngtel.GetGCPLogArgs)
	})
}

func createLocationHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtel.GetGCPLogArgs)

			return
		}

		if err := createLocation(r.Context(), repo, validate, body.Name); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusCreated, nil, ngtel.GetGCPLogArgs)
	})
}

func updateLocationHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtel.GetGCPLogArgs)

			return
		}

		if err := updateLocation(r.Context(), repo, validate, id, body.Name); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtel.GetGCPLogArgs)
	})
}

func deleteLocationHandler(repo repository) http.HandlerFunc {
	return SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := deleteLocation(r.Context(), repo, id); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtel.GetGCPLogArgs)
	})
}

func createItemHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var body writeItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtel.GetGCPLogArgs)

			return
		}

		err := createItem(r.Context(), repo, validate, body)
		if err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusCreated, nil, ngtel.GetGCPLogArgs)
	})
}

func updateItemHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var body writeItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtel.GetGCPLogArgs)

			return
		}

		if err := updateItem(r.Context(), repo, validate, id, body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtel.GetGCPLogArgs)
	})
}

func updateItemLocationHandler(repo repository) http.HandlerFunc {
	return SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			LocationID *string `json:"locationId"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtel.GetGCPLogArgs)

			return
		}

		if err := updateItemLocation(r.Context(), repo, id, body.LocationID); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtel.GetGCPLogArgs)
	})
}

func deleteItemHandler(repo repository) http.HandlerFunc {
	return SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := deleteItem(r.Context(), repo, id); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtel.GetGCPLogArgs)
	})
}
