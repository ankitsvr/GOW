package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client

func publishCommand(topic, message string) {
	token := client.Publish(topic, 0, false, message)
	token.Wait()
	fmt.Printf("sent : %s-- > %s\n", topic, message)
}

func lightHandler(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	if action != "on" && action != "off" {
		http.Error(w, "Invalid action. Use ?action=0 or ?action=off", http.StatusBadRequest)
		return
	}
	publishCommand("home/light", action)
	fmt.Fprintf(w, "Light Command sent", action)

}
func main() {
	// MQTT broker details
	broker := "tcp://172.20.1.10:1883"
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("go_api_server")

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("❌ MQTT connection error:", token.Error())
		os.Exit(1)
	}

	fmt.Println("✅ Connected to MQTT broker")

	// HTTP routes
	http.HandleFunc("/light", lightHandler)

	fmt.Println("🌐 HTTP API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
