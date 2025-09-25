package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type Telemetry struct {
	ID       int       `json:"id"`
	DeviceID string    `json:"device_id"`
	TS       int64     `json:"ts"`
	Temp     float64   `json:"temp"`
	State    string    `json:"state"`
	Created  time.Time `json:"created_at"`
}

var db *sql.DB

func main() {
	dburl := "postgres://ha:ha_pass@postgres:5432/ha_db?sslmode=disable"
	db, err := sql.Open("postgres", dburl)
	defer db.Close() //creating http request handlers

	http.HandleFunc("/telemetry", handleTelemetry)
	http.HandleFunc("/command", handleCommand)

	http.ListenAndServe(":8080", nil)

}

func handleTelemetry(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, device_id, ts, temp, state, created_at
		FROM telemetry
		ORDER BY id DESC
		LIMIT 20`)

	if err != nil {
		http.Error(w, "DB Query Error", http.StatusInternalServerError)
	}
	defer rows.Close()
	var out []Telemetry
	for rows.Next() {
		var t Telemetry
		if err := rows.Scan(&t.ID, &t.DeviceID, &t.TS, &t.Temp, &t.State, &t.Created); err != nil {
			http.Error(w, "DB Scan Error", http.StatusInternalServerError)
		}
		out = append(out, t)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)

}
func handleCommand(w http.ResponseWriter, r *http.Request) {

}
