package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func printSeatMap(db *sql.DB) {
	rows, err := db.Query(`SELECT seatnumber, user_id FROM bookings ORDER BY booking_id ASC`)
	if err != nil {
		log.Fatalf("Failed to query bookings: %v", err)
	}
	defer rows.Close()

	// Build a map of seatnumber to booked status
	seats := make(map[string]sql.NullInt64)
	for rows.Next() {
		var seat string
		var userID sql.NullInt64
		if err := rows.Scan(&seat, &userID); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		seats[seat] = userID
	}

	   // Print seat map (20 rows, 6 seats per row: A-F)
	   for row := 1; row <= 20; row++ {
		   for col := 0; col < 6; col++ {
			   seat := fmt.Sprintf("%d%c", row, 'A'+col)
			   userID := seats[seat]
			   if userID.Valid {
				   fmt.Print("x ")
			   } else {
				   fmt.Print("0 ")
			   }
			   if col == 2 { // After third seat, add aisle space
				   fmt.Print("  ")
			   }
		   }
		   fmt.Println()
	   }
}
