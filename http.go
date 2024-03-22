package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

const httpHeaderTimeout = 1 * time.Second

type genericResponse struct {
	Message string `json:"message"`
}

func respondFor(w http.ResponseWriter, code int, err error) {
	msg := http.StatusText(code)

	switch {
	case code >= http.StatusInternalServerError:
		slog.Error(msg, "err", err)
	case code >= http.StatusBadRequest:
		slog.Warn(msg, "err", err)
	default:
		slog.Info(msg)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	res := genericResponse{Message: msg}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		slog.Error("failed to encode generic response", "err", err)
	}
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

	handler = UseCORS(handler)

	return handler
}

func UseCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		for _, allowedOrigin := range strings.Split(os.Getenv("ACCESS_CONTROL_ALLOW_ORIGIN"), ",") {
			if allowedOrigin == origin {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			}
		}

		w.Header().Set("Access-Control-Allow-Headers", os.Getenv("ACCESS_CONTROL_ALLOW_HEADERS"))

		if r.Method == http.MethodOptions {
			respondFor(w, http.StatusOK, nil)

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
			respondFor(w, http.StatusInternalServerError, err)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		res := struct {
			Locations      []Location `json:"locations"`
			RemainingItems []Item     `json:"remainingItems"`
		}{Locations: locs, RemainingItems: remItems}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}
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
			respondFor(w, http.StatusNotFound, err)

			return
		} else if err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		res := struct {
			Location `json:"location"`
		}{Location: loc}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}
	}
}

func CreateLocationHandler(repo Repo, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, http.StatusBadRequest, err)

			return
		}

		if err := CreateLocation(repo, validate, body.Name); err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, http.StatusCreated, nil)
	}
}

func UpdateLocationHandler(repo Repo, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, http.StatusBadRequest, err)

			return
		}

		if err := UpdateLocation(repo, validate, id, body.Name); err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, http.StatusOK, nil)
	}
}

func DeleteLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := DeleteLocation(repo, id); err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, http.StatusOK, nil)
	}
}

func CreateItemHandler(repo Repo, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body WriteItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, http.StatusBadRequest, err)

			return
		}

		err := CreateItem(repo, validate, body)
		if err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, http.StatusCreated, nil)
	}
}

func UpdateItemHandler(repo Repo, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var body WriteItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, http.StatusBadRequest, err)

			return
		}

		if err := UpdateItem(repo, validate, id, body); err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, http.StatusOK, nil)
	}
}

func UpdateItemLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			LocationID *string `json:"locationId"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondFor(w, http.StatusBadRequest, err)

			return
		}

		if err := UpdateItemLocation(repo, id, body.LocationID); err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, http.StatusOK, nil)
	}
}

func DeleteItemHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := DeleteItem(repo, id); err != nil {
			respondFor(w, http.StatusInternalServerError, err)

			return
		}

		respondFor(w, http.StatusOK, nil)
	}
}
