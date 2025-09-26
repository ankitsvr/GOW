package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Telemetry struct {
	DeviceID string  `json:"device_id"`
	TS       int64   `json:"ts"`
	Temp     float64 `json:"temp"`
	State    string  `json:"state"`
}

func main() {

	//Connection to database
	dburl := "postgres://gow:gow_pass@localhost:5432/gow_db?sslmode=disable"

	db, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatal(err)

	}
	defer db.Close()

	// MQTT client

	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("ingestor_service")
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT connect error: %v ", token.Error())
	}
	fmt.Println("connect to the MQTT broker successfully")

	// Subscribe to the topic

	messageHandler := func(c mqtt.Client, m mqtt.Message) {
		if err := json.Unmarshal(m.Payload(), &t); err != nil {
			log.Printf("Failed to parse Json: %v, palyload: %s", err, string(m.Payload()))
		}
	}

}
