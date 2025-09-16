package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/lib/pq"
)

// Telemetry struct matches incoming JSON
type Telemetry struct {
	DeviceID string  `json:"device_id"`
	TS       int64   `json:"ts"`
	Temp     float64 `json:"temp"`
	State    string  `json:"state"`
}

func main() {
	// DB connection
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://ha:ha_pass@172.20.1.10:5432/ha_db?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	defer db.Close()

	// MQTT client
	opts := mqtt.NewClientOptions().AddBroker("tcp://172.20.1.10:1883").SetClientID("ingestor-service")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT connect error: %v", token.Error())
	}
	fmt.Println("Connected to MQTT broker")

	// message handler
	messageHandler := func(c mqtt.Client, m mqtt.Message) {
		var t Telemetry
		if err := json.Unmarshal(m.Payload(), &t); err != nil {
			log.Printf("JSON parse error: %v, payload=%s\n", err, string(m.Payload()))
			return
		}

		// insert into DB
		_, err := db.Exec(`
			INSERT INTO telemetry (device_id, ts, temp, state, raw)
			VALUES ($1, $2, $3, $4, $5)
		`, t.DeviceID, t.TS, t.Temp, t.State, string(m.Payload()))
		if err != nil {
			log.Printf("DB insert error: %v\n", err)
		} else {
			log.Printf("Inserted telemetry from %s\n", t.DeviceID)
		}
	}

	if token := client.Subscribe("devices/+/telemetry", 0, messageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("Subscribe error: %v", token.Error())
	}
	fmt.Println("Subscribed to devices/+/telemetry")

	select {} // block forever
}
