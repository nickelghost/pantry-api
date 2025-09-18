package main

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nickelghost/nghttp"
	"github.com/nickelghost/ngtelgcp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

	mux.HandleFunc("GET /locations", indexLocationsHandler(repo))
	mux.HandleFunc("GET /locations/{id}", getLocationHandler(repo))
	mux.HandleFunc("POST /locations", createLocationHandler(repo, validate))
	mux.HandleFunc("PUT /locations/{id}", updateLocationHandler(repo, validate))
	mux.HandleFunc("DELETE /locations/{id}", deleteLocationHandler(repo))
	mux.HandleFunc("POST /items", createItemHandler(repo, validate))
	mux.HandleFunc("PUT /items/{id}", updateItemHandler(repo, validate))
	mux.HandleFunc("PATCH /items/{id}/location", updateItemLocationHandler(repo))
	mux.HandleFunc("DELETE /items/{id}", deleteItemHandler(repo))
	mux.HandleFunc("/", nghttp.GetNotFoundHandler(ngtelgcp.GetLogArgs))

	var handler http.Handler = mux

	if auth != nil {
		handler = useAuth(handler, auth)
	}

	handler = nghttp.UseCORS(
		handler,
		strings.Split(os.Getenv("ACCESS_CONTROL_ALLOW_ORIGIN"), ","),
		strings.Split(os.Getenv("ACCESS_CONTROL_ALLOW_HEADERS"), ","),
		[]string{"*"},
		ngtelgcp.GetLogArgs,
	)
	handler = nghttp.UseRequestLogging(handler, ngtelgcp.GetLogArgs)
	handler = nghttp.UseRequestID(handler, "X-Request-ID")
	handler = otelhttp.NewHandler(handler, "request")

	return handler
}

func useAuth(next http.Handler, auth authentication) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := auth.Check(r.Context(), r); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusUnauthorized, err, ngtelgcp.GetLogArgs)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func indexLocationsHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tags *[]string

		if val := r.URL.Query().Get("tags"); val != "" {
			vals := strings.Split(val, ",")
			tags = &vals
		}

		locs, remItems, err := getLocations(r.Context(), repo, tags)
		if err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtelgcp.GetLogArgs)

			return
		}

		res := struct {
			Locations      []location `json:"locations"`
			RemainingItems []item     `json:"remainingItems"`
		}{Locations: locs, RemainingItems: remItems}

		nghttp.Respond(w, r, http.StatusOK, nil, res, ngtelgcp.GetLogArgs)
	}
}

func getLocationHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var tags *[]string

		if val := r.URL.Query().Get("tags"); val != "" {
			vals := strings.Split(val, ",")
			tags = &vals
		}

		loc, err := getLocation(r.Context(), repo, id, tags)
		if errors.Is(err, errLocationNotFound) {
			nghttp.RespondGeneric(w, r, http.StatusNotFound, err, ngtelgcp.GetLogArgs)

			return
		} else if err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtelgcp.GetLogArgs)

			return
		}

		res := struct {
			location `json:"location"`
		}{location: loc}

		nghttp.Respond(w, r, http.StatusOK, nil, res, ngtelgcp.GetLogArgs)
	}
}

func createLocationHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtelgcp.GetLogArgs)

			return
		}

		if err := createLocation(r.Context(), repo, validate, body.Name); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtelgcp.GetLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusCreated, nil, ngtelgcp.GetLogArgs)
	}
}

func updateLocationHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtelgcp.GetLogArgs)

			return
		}

		if err := updateLocation(r.Context(), repo, validate, id, body.Name); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtelgcp.GetLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtelgcp.GetLogArgs)
	}
}

func deleteLocationHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := deleteLocation(r.Context(), repo, id); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtelgcp.GetLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtelgcp.GetLogArgs)
	}
}

func createItemHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body writeItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtelgcp.GetLogArgs)

			return
		}

		err := createItem(r.Context(), repo, validate, body)
		if err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtelgcp.GetLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusCreated, nil, ngtelgcp.GetLogArgs)
	}
}

func updateItemHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var body writeItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtelgcp.GetLogArgs)

			return
		}

		if err := updateItem(r.Context(), repo, validate, id, body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtelgcp.GetLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtelgcp.GetLogArgs)
	}
}

func updateItemLocationHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			LocationID *string `json:"locationId"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtelgcp.GetLogArgs)

			return
		}

		if err := updateItemLocation(r.Context(), repo, id, body.LocationID); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtelgcp.GetLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtelgcp.GetLogArgs)
	}
}

func deleteItemHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := deleteItem(r.Context(), repo, id); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtelgcp.GetLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtelgcp.GetLogArgs)
	}
}
