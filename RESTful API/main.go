package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	connStr := "postgres://postgres:postgres@localhost:5432/movie_db?sslmode=disable"

	store, err := NewStore(connStr)
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	
	api := &API{Store: store}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/movies", api.handleListMovies)
	r.Get("/movies/{id}/availability", api.handleGetAvailability)
	r.Post("/movies/{id}/book", api.handleBookMovie)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

type API struct {
	Store *Store
}

func (a *API) handleListMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := a.Store.GetMovies()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not fetch movies"})
		return
	}
	writeJSON(w, http.StatusOK, movies)
}

func (a *API) handleGetAvailability(w http.ResponseWriter, r *http.Request) {
	movieID, err := getMovieID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid movie ID"})
		return
	}

	availability, err := a.Store.GetMovieAvailability(movieID)
	if err != nil {
		if err.Error() == "movie not found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, availability)
}

func (a *API) handleBookMovie(w http.ResponseWriter, r *http.Request) {
	movieID, err := getMovieID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid movie ID"})
		return
	}

	var req BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if req.UserName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_name is required"})
		return
	}

	receipt, err := a.Store.BookSeat(movieID, req.UserName)
	if err != nil {
		if err.Error() == "no available seats" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		} else if err.Error() == "movie not found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		} else {
			log.Printf("Internal error booking seat: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not process booking"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, receipt) // 201 Created
}

func getMovieID(r *http.Request) (int, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.Atoi(idStr)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}