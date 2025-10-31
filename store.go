package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

type Store struct {
	DB *sql.DB
}

func NewStore(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Store{DB: db}, nil
}

type Movie struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	TotalSeats     int    `json:"total_seats"`
	AvailableSeats int    `json:"available_seats"`
}

type Availability struct {
	MovieID        int `json:"movie_id"`
	AvailableSeats int `json:"available_seats"`
}

type BookingRequest struct {
	UserName string `json:"user_name"`
}

type BookingReceipt struct {
	BookingID int    `json:"booking_id"`
	MovieID   int    `json:"movie_id"`
	UserName  string `json:"user_name"`
	Message   string `json:"message"`
}

func (s *Store) GetMovies() ([]Movie, error) {
	rows, err := s.DB.Query("SELECT id, title, total_seats, available_seats FROM movies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.TotalSeats, &m.AvailableSeats); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, nil
}

func (s *Store) GetMovieAvailability(movieID int) (*Availability, error) {
	var a Availability
	a.MovieID = movieID

	err := s.DB.QueryRow("SELECT available_seats FROM movies WHERE id = $1", movieID).Scan(&a.AvailableSeats)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("movie not found")
		}
		return nil, err
	}
	return &a, nil
}

func (s *Store) BookSeat(movieID int, userName string) (*BookingReceipt, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	var availableSeats int
	err = tx.QueryRow("SELECT available_seats FROM movies WHERE id = $1 FOR UPDATE", movieID).Scan(&availableSeats)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("movie not found")
		}
		return nil, fmt.Errorf("could not get movie availability: %w", err)
	}

	if availableSeats <= 0 {
		return nil, errors.New("no available seats")
	}

	_, err = tx.Exec("UPDATE movies SET available_seats = available_seats - 1 WHERE id = $1", movieID)
	if err != nil {
		return nil, fmt.Errorf("could not update movie seats: %w", err)
	}

	var bookingID int
	err = tx.QueryRow(
		"INSERT INTO bookings (movie_id, user_name) VALUES ($1, $2) RETURNING id",
		movieID,
		userName,
	).Scan(&bookingID)
	if err != nil {
		return nil, fmt.Errorf("could not create booking: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	receipt := &BookingReceipt{
		BookingID: bookingID,
		MovieID:   movieID,
		UserName:  userName,
		Message:   "Booking successful",
	}
	return receipt, nil
}