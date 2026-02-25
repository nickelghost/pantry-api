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
	"github.com/nickelghost/ngtel"
)

const (
	httpTimeout       = 10 * time.Second
	httpHeaderTimeout = 1 * time.Second
)

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
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("GET /locations", indexLocationsHandler(repo))
	apiMux.HandleFunc("GET /locations/{id}", getLocationHandler(repo))
	apiMux.HandleFunc("POST /locations", createLocationHandler(repo, validate))
	apiMux.HandleFunc("PUT /locations/{id}", updateLocationHandler(repo, validate))
	apiMux.HandleFunc("DELETE /locations/{id}", deleteLocationHandler(repo))
	apiMux.HandleFunc("POST /items", createItemHandler(repo, validate))
	apiMux.HandleFunc("PUT /items/{id}", updateItemHandler(repo, validate))
	apiMux.HandleFunc("PATCH /items/{id}/location", updateItemLocationHandler(repo))
	apiMux.HandleFunc("DELETE /items/{id}", deleteItemHandler(repo))
	apiMux.HandleFunc("/", nghttp.GetNotFoundHandler(ngtel.GetGCPLogArgs))

	var apiHandler http.Handler = apiMux

	if auth != nil {
		apiHandler = authMiddleware(apiHandler, auth)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle("/", apiHandler)

	var handler http.Handler = mux

	handler = nghttp.UseCORS(
		handler,
		strings.Split(os.Getenv("ACCESS_CONTROL_ALLOW_ORIGIN"), ","),
		strings.Split(os.Getenv("ACCESS_CONTROL_ALLOW_HEADERS"), ","),
		[]string{"*"},
		ngtel.GetGCPLogArgs,
	)
	handler = nghttp.UseRequestLogging(handler, ngtel.GetGCPLogArgs)
	handler = nghttp.UseRequestID(handler, "X-Request-ID")
	handler = ngtel.RequestMiddleware(handler)

	return handler
}

func indexLocationsHandler(repo repository) http.HandlerFunc {
	return ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	return ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	return ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtel.GetGCPLogArgs)

			return
		}

		if err := createLocation(r.Context(), repo, validate, body.Name); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, errValidation) {
				status = http.StatusBadRequest
			}

			nghttp.RespondGeneric(w, r, status, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusCreated, nil, ngtel.GetGCPLogArgs)
	})
}

func updateLocationHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtel.GetGCPLogArgs)

			return
		}

		if err := updateLocation(r.Context(), repo, validate, id, body.Name); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, errValidation) {
				status = http.StatusBadRequest
			}

			nghttp.RespondGeneric(w, r, status, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtel.GetGCPLogArgs)
	})
}

func deleteLocationHandler(repo repository) http.HandlerFunc {
	return ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := deleteLocation(r.Context(), repo, id); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtel.GetGCPLogArgs)
	})
}

func createItemHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var body writeItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtel.GetGCPLogArgs)

			return
		}

		if err := createItem(r.Context(), repo, validate, body); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, errValidation) {
				status = http.StatusBadRequest
			}

			nghttp.RespondGeneric(w, r, status, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusCreated, nil, ngtel.GetGCPLogArgs)
	})
}

func updateItemHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var body writeItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusBadRequest, err, ngtel.GetGCPLogArgs)

			return
		}

		if err := updateItem(r.Context(), repo, validate, id, body); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, errValidation) {
				status = http.StatusBadRequest
			}

			nghttp.RespondGeneric(w, r, status, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtel.GetGCPLogArgs)
	})
}

func updateItemLocationHandler(repo repository) http.HandlerFunc {
	return ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	return ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := deleteItem(r.Context(), repo, id); err != nil {
			nghttp.RespondGeneric(w, r, http.StatusInternalServerError, err, ngtel.GetGCPLogArgs)

			return
		}

		nghttp.RespondGeneric(w, r, http.StatusOK, nil, ngtel.GetGCPLogArgs)
	})
}
