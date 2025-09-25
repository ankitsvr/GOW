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

type Telemetry struct {
	Device_id string  `json:"device_id"`
	TS        int64   `json:"ts"`
	TEMP      float64 `json:"temp"`
	State     string  `json:"state"`
}

func main() {
	// DB Connection
	db_url := os.Getenv("DB_URL")
	if db_url == "" {
		db_url = "postgres://ha:ha_pass@localhost:5432/ha_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", db_url)
	if err != nil {
		log.Fatal("DB connection error: ", err)
	}
	defer db.Close()

	// MQTT Client
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("ingestor-service")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT connect error: %v", token.Error())
	}

	fmt.Println("Connected the MQTT broker")
	// Message handler

	msgHandler := func(m mqtt.Client, msg mqtt.Message) {
		var t Telemetry
		err := json.Unmarshal(msg.Payload(), &t)
		if err != nil {
			log.Printf("Error unmarshalling message: %v, payload=%S\n", err, msg.Payload())
			return
		}

		// Now we have the msg lets insert the value into the DB

		_, err = db.Exec(`
			INSERT INTO telemetry (device_id, ts, temp, state, raw) VALUES ($1, $2, $3, $4, $5) `, t.Device_id, t.TS, t.TEMP, t.State, string(msg.Payload()))
		if err != nil {
			log.Printf("DB insert error: %v", err)
		} else {
			log.Printf("Inserted telemetry for device %s", t.Device_id)

		}
	}
	if token := client.Subscribe(("devices/+/telemetry"), 0, msgHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT subscribe error: %v", token.Error())
	}
	fmt.Println("Subscribed to topic devices/+/telemetry")

	// Keep the program running
	select {}

}
