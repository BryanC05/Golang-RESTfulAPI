# Golang Movie Reservation API

This is a simple RESTful API for a movie reservation system, built with Go (Golang).

It demonstrates how to handle complex business logic like **atomic bookings** and **preventing concurrency issues** (race conditions) using PostgreSQL database transactions and row-level locking.

This backend provides endpoints to list movies, check seat availability, and book a ticket.

-----

## 🚀 API Endpoints

### 1\. List All Movies

  * **Endpoint:** `GET /movies`
  * **Description:** Retrieves a JSON array of all available movies.
  * **Success Response (200 OK):**
    ```json
    [
      {
        "id": 1,
        "title": "Dune: Part Two",
        "total_seats": 100,
        "available_seats": 100
      },
      {
        "id": 2,
        "title": "The Matrix",
        "total_seats": 50,
        "available_seats": 50
      }
    ]
    ```

### 2\. Get Movie Availability

  * **Endpoint:** `GET /movies/{id}/availability`
  * **Description:** Retrieves the number of available seats for a single movie.
  * **Success Response (200 OK):**
    ```json
    {
      "movie_id": 3,
      "available_seats": 1
    }
    ```

### 3\. Book a Seat

  * **Endpoint:** `POST /movies/{id}/book`
  * **Description:** Attempts to book one seat for the specified movie.
  * **Request Body (JSON):**
    ```json
    {
      "user_name": "Alex"
    }
    ```
  * **Success Response (201 Created):**
    ```json
    {
      "booking_id": 1,
      "movie_id": 3,
      "user_name": "Alex",
      "message": "Booking successful"
    }
    ```
  * **Failure Response (409 Conflict):**
    This response is returned if no seats are available.
    ```json
    {
      "error": "no available seats"
    }
    ```

-----

## 🛠️ How to Run

### Prerequisites

  * **Go:** [Version 1.21+](https://go.dev/dl/)
  * **PostgreSQL:** [A running PostgreSQL server](https://www.postgresql.org/download/)

### 1\. Set Up the Database

1.  Log in to your PostgreSQL server (e.g., using `psql`).

2.  Create a database for the project:

    ```sql
    CREATE DATABASE movie_db;
    ```

3.  Connect to your new database (`\c movie_db`).

4.  Run the `setup.sql` script (or just copy-paste the text below) to create the tables and add sample data:

    ```sql
    -- Create the movies table
    CREATE TABLE movies (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        total_seats INT NOT NULL,
        available_seats INT NOT NULL
    );

    -- Create the bookings table to log reservations
    CREATE TABLE bookings (
        id SERIAL PRIMARY KEY,
        movie_id INT NOT NULL REFERENCES movies(id),
        user_name VARCHAR(255) NOT NULL,
        booked_at TIMESTAMPTZ DEFAULT NOW()
    );

    -- Insert some dummy data
    INSERT INTO movies (title, total_seats, available_seats)
    VALUES
        ('Dune: Part Two', 100, 100),
        ('The Matrix', 50, 50),
        ('Race for the Last Seat', 1, 1); -- Movie to test concurrency
    ```

### 2\. Configure and Run the Server

1.  **Clone or download** this project's files.

2.  **Install dependencies:**

    ```bash
    go get github.com/go-chi/chi/v5
    go get github.com/lib/pq
    ```

3.  **Update Connection String:** Open `main.go` and find this line:

    ```go
    connStr := "postgres://your_user:your_password@localhost:5432/your_db?sslmode=disable"
    ```

    Change it to match your PostgreSQL username, password, and database name (e.g., `movie_db`).

4.  **Run the server:**

    ```bash
    go run .
    ```

    The server will start on `http://localhost:8080`.

-----

## 🧪 How to Test

You can test the API using `curl` from your terminal.

1.  **Get all movies:**

    ```bash
    curl http://localhost:8080/movies
    ```

2.  **Book the last seat (for movie ID 3):**

    ```bash
    curl -X POST http://localhost:8080/movies/3/book -d '{"user_name": "MyUser"}'
    ```

3.  **Try to book it again (will fail):**
    This demonstrates the concurrency-safe logic.

    ```bash
    curl -X POST http://localhost:8080/movies/3/book -d '{"user_name": "AnotherUser"}'
    ```

    **Expected Output:**

    ```json
    {"error":"no available seats"}
    ```

## 💡 Key Concepts

This project uses `SELECT ... FOR UPDATE` within a SQL transaction to prevent a "race condition," where two users might try to book the last seat at the same time.

Here's the flow for `BookSeat`:

1.  `tx, err := s.DB.Begin()`: Start a new transaction.
2.  `SELECT ... FOR UPDATE`: Locks the specific movie row. Any other transaction trying to read this row must wait.
3.  The code checks if `available_seats > 0`.
4.  `UPDATE movies SET ...`: The seat count is safely decremented.
5.  `INSERT INTO bookings ...`: The booking is logged.
6.  `tx.Commit()`: All changes are saved, and the row lock is released.

If any step fails, `tx.Rollback()` is called automatically (using `defer`), and no changes are made.
