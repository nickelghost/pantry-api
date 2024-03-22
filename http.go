package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

const httpHeaderTimeout = 1 * time.Second

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
	mux.HandleFunc("POST /items", CreateItemHandler(repo))
	mux.HandleFunc("PUT /items/{id}", UpdateItemHandler(repo))
	mux.HandleFunc("PATCH /items/{id}/quantity", UpdateItemQuantityHandler(repo))
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
			w.WriteHeader(http.StatusOK)

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
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		res := struct {
			Locations      []Location `json:"locations"`
			RemainingItems []Item     `json:"remainingItems"`
		}{Locations: locs, RemainingItems: remItems}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

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
			w.WriteHeader(http.StatusNotFound)

			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		res := struct {
			Location `json:"location"`
		}{Location: loc}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

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
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)

			return
		}

		if err := CreateLocation(repo, validate, body.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateLocationHandler(repo Repo, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)

			return
		}

		if err := UpdateLocation(repo, validate, id, body.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func DeleteLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := DeleteLocation(repo, id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func CreateItemHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body WriteItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)

			return
		}

		if err := CreateItem(repo, body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateItemHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var body WriteItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)

			return
		}

		if err := UpdateItem(repo, id, body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func UpdateItemQuantityHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			Quantity *int `json:"quantity"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)

			return
		}

		if err := UpdateItemQuantity(repo, id, body.Quantity); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func UpdateItemLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		body := struct {
			LocationID *string `json:"locationId"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)

			return
		}

		if err := UpdateItemLocation(repo, id, body.LocationID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func DeleteItemHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := DeleteItem(repo, id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
