package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func GetRouter(repo Repo) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /locations", IndexLocationsHandler(repo))
	mux.HandleFunc("GET /locations/{id}", GetLocationHandler(repo))
	mux.HandleFunc("POST /locations", CreateLocationHandler(repo))
	mux.HandleFunc("PUT /locations/{id}", UpdateLocationHandler(repo))
	mux.HandleFunc("DELETE /locations/{id}", DeleteLocationHandler(repo))
	mux.HandleFunc("POST /items", CreateItemHandler(repo))
	mux.HandleFunc("PUT /items/{id}", UpdateItemHandler(repo))
	mux.HandleFunc("PATCH /items/{id}/quantity", UpdateItemQuantityHandler(repo))
	mux.HandleFunc("PATCH /items/{id}/location", UpdateItemLocationHandler(repo))
	mux.HandleFunc("DELETE /items/{id}", DeleteItemHandler(repo))

	return mux
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
			w.WriteHeader(500)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		res := struct {
			Locations      []Location `json:"locations"`
			RemainingItems []Item     `json:"remainingItems"`
		}{Locations: locs, RemainingItems: remItems}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			w.WriteHeader(500)
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
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		res := struct {
			Location `json:"location"`
		}{Location: loc}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
	}
}

func CreateLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(400)
			log.Println(err)
			return
		}

		if err := CreateLocation(repo, body.Name); err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
	}
}

func UpdateLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		body := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(400)
			log.Println(err)
			return
		}

		if err := UpdateLocation(repo, id, body.Name); err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
	}
}

func DeleteLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := DeleteLocation(repo, id); err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
	}
}

func CreateItemHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body WriteItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(400)
			log.Println(err)
			return
		}

		if err := CreateItem(repo, body); err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
	}
}

func UpdateItemHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body WriteItemParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(400)
			log.Println(err)
			return
		}

		if err := UpdateItem(repo, id, body); err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
	}
}

func UpdateItemQuantityHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		body := struct {
			Quantity *int `json:"quantity"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(400)
			log.Println(err)
			return
		}

		if err := UpdateItemQuantity(repo, id, body.Quantity); err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
	}
}

func UpdateItemLocationHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		body := struct {
			LocationID *string `json:"locationId"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(400)
			log.Println(err)
			return
		}

		if err := UpdateItemLocation(repo, id, body.LocationID); err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
	}
}

func DeleteItemHandler(repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := DeleteItem(repo, id); err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
	}
}
