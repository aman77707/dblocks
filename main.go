package main

import (
	"database/sql"
	"fmt"
	"log"
	"github.com/jaswdr/faker"
	_ "github.com/lib/pq"
)

func createTables(db *sql.DB) error {
	// Check if tables exist
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'trips', 'bookings')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if tables exist: %w", err)
	}
	if exists == 3 {
		fmt.Println("Tables already exist. Skipping table creation.")
		return nil
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS trips (
			trip_id SERIAL PRIMARY KEY,
			airline_name TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS bookings (
			booking_id SERIAL PRIMARY KEY,
			airline_name TEXT NOT NULL,
			seatnumber TEXT NOT NULL,
			user_id INTEGER REFERENCES users(user_id) DEFAULT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	fmt.Println("Tables created successfully.")
	return nil
}



func main() {
	user := "postgres"
	password := "postgres"
	dbname := "postgres"
	host := "localhost"
	port := 5432

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Step 1: Create DB if not exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'airline_booking')").Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if database exists: %v", err)
	}
	if !exists {
		_, err = db.Exec("CREATE DATABASE airline_booking")
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		fmt.Println("Database 'airline_booking' created successfully.")
	} else {
		fmt.Println("Database 'airline_booking' already exists. Skipping creation.")
	}

	// Step 2: Connect to the new database
	db.Close()
	dbname = "airline_booking"
	psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect to new database: %v", err)
	}
	defer db.Close()

	// Step 3: Create tables if not exist
	if err := createTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Step 4: Populate tables if not already populated
	if err := populateTables(db); err != nil {
		log.Fatalf("Failed to populate tables: %v", err)
	}

}
