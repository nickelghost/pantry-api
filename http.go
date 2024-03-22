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

func GetServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              net.JoinHostPort("0.0.0.0", "8080"),
		Handler:           handler,
		ReadHeaderTimeout: httpHeaderTimeout,
	}
}

func GetRouter(repo Repo, validate *validator.Validate) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /locations", IndexLocationsHandler(repo))
	mux.HandleFunc("GET /locations/{id}", GetLocationHandler(repo))
	mux.HandleFunc("POST /locations", CreateLocationHandler(repo, validate))
	mux.HandleFunc("PUT /locations/{id}", UpdateLocationHandler(repo, validate))
	mux.HandleFunc("DELETE /locations/{id}", DeleteLocationHandler(repo))
	mux.HandleFunc("POST /items", CreateItemHandler(repo, validate))
	mux.HandleFunc("PUT /items/{id}", UpdateItemHandler(repo, validate))
	mux.HandleFunc("PATCH /items/{id}/location", UpdateItemLocationHandler(repo))
	mux.HandleFunc("DELETE /items/{id}", DeleteItemHandler(repo))

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

func IndexLocationsHandler(repo Repo) http.HandlerFunc {
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

		locs, remItems, err := GetLocations(repo, search, tags)
		if err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		res := struct {
			Locations      []Location `json:"locations"`
			RemainingItems []Item     `json:"remainingItems"`
		}{Locations: locs, RemainingItems: remItems}

		respond(w, r, http.StatusOK, nil, res)
	}
}

func GetLocationHandler(repo Repo) http.HandlerFunc {
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

		loc, err := GetLocation(repo, id, search, tags)
		if errors.Is(err, ErrLocationNotFound) {
			respondFor(w, r, http.StatusNotFound, err)

			return
		} else if err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		res := struct {
			Location `json:"location"`
		}{Location: loc}

		respond(w, r, http.StatusOK, nil, res)
	}
}

func CreateLocationHandler(repo Repo, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		if err := CreateLocation(repo, validate, body.Name); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusCreated, nil)
	}
}

func UpdateLocationHandler(repo Repo, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		if err := UpdateLocation(repo, validate, id, body.Name); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}

func DeleteLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := DeleteLocation(repo, id); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}

func CreateItemHandler(repo Repo, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body WriteItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		err := CreateItem(repo, validate, body)
		if err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusCreated, nil)
	}
}

func UpdateItemHandler(repo Repo, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var body WriteItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		if err := UpdateItem(repo, validate, id, body); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}

func UpdateItemLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			LocationID *string `json:"locationId"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, r, http.StatusBadRequest, err)

			return
		}

		if err := UpdateItemLocation(repo, id, body.LocationID); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}

func DeleteItemHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := DeleteItem(repo, id); err != nil {
			respondFor(w, r, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, r, http.StatusOK, nil)
	}
}
