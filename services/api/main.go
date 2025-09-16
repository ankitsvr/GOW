package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2/log"
)

type Telemetry struct {
	ID int  `json:"id"`
	DeviceID string `json:"deviceid"`
	TS int64 `json:"ts"`
	Temp float64 `json:"temp"`
	State string `json:"state"`
	Created time.Time `json:"created"`


}

var db *sql.DB
var mqttClient mqtt.Client

func main() {
	//DB connection 

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL ="postgres://ha:ha_pass@172.20.1.10:5432/ha_db?sslmode=disable"
	}

	var err error
	db, err := sql.Open("postgres", dbURL)
	if err !=nil {
		log.Fatalf("DB Connect error %s", err)
	}

	defer db.Close()

	//MQTT connection 

	broker := os.Getenv("MQTT_URL")
	if broker == "" {
		broker = "tcp://172.20.1.10:1883"
	}

	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID("api-service")
	mqttClient := mqtt.NewClient(opts)

	if token  := mqttClient.Connect(); token.Wait() && token.Error() !=nil {
		log.Fatalf("MQTT connect error: %v", token.Error())
	}

	fmt.Println("connected to MQTT")
	
	//routes
	http.HandleFunc("/telemetry",handleGetTelemetry)
	http.HandleFunc("/command", handleSendCommand)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("API service running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))




}

func handleGetTelemetry(w http.ResponseWriter, r *http.Request) {
	row, err := db.Query(`
		SELECT id, device_id, ts, temp, state, created_at
		FROM telemetry
		ORDER BY id DESC
		LIMIT 20	
	
	`)

	if err !=nil {
		http.Error(w, err.Error(),500)
		return
	}
	defer row.Close()

	var out []Telemetry

	for row.Next(){
		var t Telemetry
		if err := row.Scan(&t.ID, &t.DeviceID, &t.TS, &t.Temp, &t.State, &t.Created); err != nil {
			http.Error(w, err.Error(),500)
		}
		out = append(out, t)
		
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)

}


