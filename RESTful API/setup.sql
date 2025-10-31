CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    total_seats INT NOT NULL,
    available_seats INT NOT NULL
);

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
    ('Race for the Last Seat', 1, 1);