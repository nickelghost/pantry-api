package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const httpHeaderTimeout = 1 * time.Second

func respond(w http.ResponseWriter, r *http.Request, code int, err error, res any) {
	requestID, _ := r.Context().Value(requestIDKey).(string)
	statusText := http.StatusText(code)

	switch {
	case code >= http.StatusInternalServerError:
		slog.Error(statusText, "requestID", requestID, "err", err)
	case code >= http.StatusBadRequest:
		slog.Warn(statusText, "requestID", requestID, "err", err)
	default:
		slog.Info(statusText, "requestID", requestID)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		slog.Error("failed to encode response", "err", err, "requestID", requestID)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

type genericResponse struct {
	Message string `json:"message"`
}

func respondFor(w http.ResponseWriter, r *http.Request, code int, err error) {
	res := genericResponse{Message: http.StatusText(code)}
	respond(w, r, code, err, res)
}

func getServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              net.JoinHostPort("0.0.0.0", "8080"),
		Handler:           handler,
		ReadHeaderTimeout: httpHeaderTimeout,
	}
}

func getRouter(repo repository, validate *validator.Validate) http.Handler {
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

	var handler http.Handler = mux

	handler = useCORS(handler)
	handler = useRequestLogging(handler)
	handler = useRequestID(handler)

	return handler
}

func useRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func useRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		requestID, _ := r.Context().Value(requestIDKey).(string)

		slog.Info(
			"Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
			"requestID", requestID,
		)
	})
}

func useCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		for _, allowedOrigin := range strings.Split(os.Getenv("ACCESS_CONTROL_ALLOW_ORIGIN"), ",") {
			if allowedOrigin == origin {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			}
		}

		w.Header().Set("Access-Control-Allow-Headers", os.Getenv("ACCESS_CONTROL_ALLOW_HEADERS"))

		if r.Method == http.MethodOptions {
			respondFor(w, r, http.StatusOK, nil)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func indexLocationsHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var search *string

		var tags *[]string

		if val := r.URL.Query().Get("search"); val != "" {
			search = &val
		}

		if val := r.URL.Query().Get("tags"); val != "" {
			vals := strings.Split(val, ",")
			tags = &vals
		}

		locs, remItems, err := getLocations(repo, search, tags)
		if err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		res := struct {
			Locations      []location `json:"locations"`
			RemainingItems []item     `json:"remainingItems"`
		}{Locations: locs, RemainingItems: remItems}

		respond(w, r, http.StatusOK, nil, res)
	}
}

func getLocationHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var search *string

		var tags *[]string

		if val := r.URL.Query().Get("search"); val != "" {
			search = &val
		}

		if val := r.URL.Query().Get("tags"); val != "" {
			vals := strings.Split(val, ",")
			tags = &vals
		}

		loc, err := getLocation(repo, id, search, tags)
		if errors.Is(err, errLocationNotFound) {
			respondFor(w, r, http.StatusNotFound, err)

			return
		} else if err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		res := struct {
			location `json:"location"`
		}{location: loc}

		respond(w, r, http.StatusOK, nil, res)
	}
}

func createLocationHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		if err := createLocation(repo, validate, body.Name); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusCreated, nil)
	}
}

func updateLocationHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		if err := updateLocation(repo, validate, id, body.Name); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}

func deleteLocationHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := deleteLocation(repo, id); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}

func createItemHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body writeItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		err := createItem(repo, validate, body)
		if err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusCreated, nil)
	}
}

func updateItemHandler(repo repository, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var body writeItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		if err := updateItem(repo, validate, id, body); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}

func updateItemLocationHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			LocationID *string `json:"locationId"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		if err := updateItemLocation(repo, id, body.LocationID); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}

func deleteItemHandler(repo repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := deleteItem(repo, id); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}
