package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
	_ "github.com/lib/pq"
)

type User struct {
	ID   int
	Name string
}

func getUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT user_id, name FROM users ORDER BY user_id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func bookSeatsAsync(db *sql.DB, users []User) {
	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)
		go func(u User) {
			defer wg.Done()
			startBooking(db, u)
		}(user)
	}
	wg.Wait()
}

func startBooking(db *sql.DB, user User) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Printf("User %d: failed to begin transaction: %v", user.ID, err)
		return
	}

	var seatnumber string
	var bookingid int
	var userid sql.NullInt64
	err = tx.QueryRow(`SELECT booking_id, seatnumber, user_id FROM bookings WHERE user_id IS NULL ORDER BY booking_id LIMIT 1 SKIP LOCKED`).Scan(&bookingid, &seatnumber, &userid)
	if err != nil {
		log.Printf("No Rows found for user %s with userid %d, Reattempting.", seatnumber, user.Name, user.ID)
		tx.Rollback()
	} 

	_, err = tx.Exec(`UPDATE bookings SET user_id = $1 WHERE booking_id = $2`, user.ID, bookingid)
	if err != nil {
		tx.Rollback()
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
	}
	
	log.Printf("Seat %s is booked by user %s with userid %d", seatnumber, user.Name, user.ID)
	return
}

func main() {
	user := "postgres"
	password := "postgres"
	dbname := "airline_booking"
	host := "localhost"
	port := 5432

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	users, err := getUsers(db)
	if err != nil {
		log.Fatalf("Failed to get users: %v", err)
	}
	if len(users) < 120 {
		log.Fatalf("Not enough users to book all seats (found %d)", len(users))
	}
	
	// Reset all bookings before starting
	_, err = db.Exec("UPDATE bookings SET user_id = NULL")
	if err != nil {
		log.Fatalf("Failed to reset bookings: %v", err)
	}
	
	fmt.Println("\n")
	start := time.Now()
	bookSeatsAsync(db, users)
	elapsed := time.Since(start)
	fmt.Println("\n")
	fmt.Printf("Total time taken to book all seats: %v\n", elapsed)
	fmt.Println("\n")
	printSeatMap(db)
	fmt.Println("\n")
}
