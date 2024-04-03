CREATE DATABASE IF NOT EXISTS media_library ;

USE media_library;

CREATE TABLE IF NOT EXISTS books(
    book_id INT PRIMARY KEY,
    title VARCHAR(255),
    authors TEXT,
    average_rating DECIMAL(3,2),
    isbn VARCHAR(20),
    isbn13 VARCHAR(20),
    language_code VARCHAR(10),
    num_pages INT,
    ratings_count INT,
    text_reviews_count INT,
    publication_date DATE,
    publisher VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS movies(
    movie_id INT PRIMARY KEY,
    title VARCHAR(255),
    publish_year INT,
    summary TEXT,
    short_summary TEXT,
    imdb_id VARCHAR(20),
    runtime INT,
    youtube_trailer VARCHAR(255),
    rating DECIMAL(3,1),
    movie_poster VARCHAR(255),
    director VARCHAR(255),
    writers VARCHAR(255),
    cast TEXT
);

CREATE TABLE IF NOT EXISTS search_events (
    search_id VARCHAR(255) PRIMARY KEY,
    search_query VARCHAR(255),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS search_clicks (
    search_click_id INT AUTO_INCREMENT PRIMARY KEY,
    search_id VARCHAR(255),
    result_type ENUM('book', 'movie'),
    result_id INT,
    result_position INT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (search_id) REFERENCES search_events(search_id)
    );

CREATE INDEX idx_movie_id ON movies (movie_id);
CREATE INDEX idx_book_id ON books (book_id);
