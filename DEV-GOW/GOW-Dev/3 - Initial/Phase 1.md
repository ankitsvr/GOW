
### Initial 

We will start with the Mono-Repo

```
Gow/
├─ docker-compose.yml
├─ mosquitto/
│  └─ config/
│     └─ mosquitto.conf
├─ services/
│  ├─ device-simulator/
│  │  ├─ device_simulator.py
│  │  └─ Dockerfile
│  └─ ingestor/
│     ├─ main.go
│     └─ go.mod

```


### Step 1

I have an ubuntu  server VM installed at my Proxomox server . I will be using this server as Mosquitto and Postgres DB

I have installed the services with the below docker-compose configuration 

```
version: '3.8'
services:
  mqtt:
    image: eclipse-mosquitto:2.0
    container_name: ha_mosquitto
    ports:
      - "1883:1883"    # MQTT
      - "9001:9001"    # Websockets (useful later)
    volumes:
      - ./mosquitto/config/mosquitto.conf:/mosquitto/config/mosquitto.conf:ro
      - mosq_data:/mosquitto/data

  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: ha
      POSTGRES_PASSWORD: ha_pass
      POSTGRES_DB: ha_db
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  mosq_data:
  pgdata:

```

**Just run the docker compose up from the same directory where docker-compose.yml file lives**


### Issue
The port 1883 is already in use

```
sudo netstat -tulpn | grep 1883
sudo lsof -i :1883
sudo ss -tulpn | grep 1883


sudo systemctl stop mosquitto
sudo systemctl disable mosquitto

Both services (Mosquitto and Postgre) are up now
```


#### Add the following config on the `mosquitto/config/mosquitto.conf`

```
listener 1883
allow_anonymous true

listener 9001
protocol websockets

persistence true
persistence_file mosquitto.db

```



#### Step 2

*Now we will create a device simulator in the python `/services/device-simulator/device-simulator.py`




```


Simple device simulator:
 - publishes telemetry to: devices/<device_id>/telemetry
 - listens for commands on: devices/<device_id>/commands
"""
import time
import json
import argparse
import random
import threading
import paho.mqtt.client as mqtt


def on_message(client, userdata, msg):
    try:
        payload = msg.payload.decode()
        print(f"[{userdata['device_id']}] CMD recv on {msg.topic}: {payload}")
        # Simple command handling:
        j = json.loads(payload)
        if j.get("command") == "toggle":
            userdata['state'] = "on" if userdata['state'] == "off" else "off"
            print(f"[{userdata['device_id']}] toggled -> {userdata['state']}")
    except Exception as e:
        print("cmd parse error:", e)


def run_sim(device_id, broker_host="localhost", broker_port=1883):
    userdata = {"device_id": device_id, "state": "off"}
    client = mqtt.Client(client_id=f"sim-{device_id}", userdata=userdata)
    client.on_message = on_message
    client.connect(broker_host, broker_port)
    client.loop_start()
    client.subscribe(f"devices/{device_id}/commands")

    try:
        while True:
            telemetry = {
                "device_id": device_id,
                "ts": int(time.time()),
                "temp": round(20 + random.random()*10, 2),
                "state": userdata['state']
            }
            topic = f"devices/{device_id}/telemetry"
            payload = json.dumps(telemetry)
            client.publish(topic, payload)
            print(f"[{device_id}] published -> {payload}")
            time.sleep(5)
    except KeyboardInterrupt:
        pass
    finally:
        client.loop_stop()
        client.disconnect()


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--device-id", default="light1")
    parser.add_argument("--host", default="localhost")
    parser.add_argument("--port", type=int, default=1883)
    args = parser.parse_args()
    run_sim(args.device_id, args.host, args.port)

```


The above script will publish the telemetry  to the MQTT broker ( to the given host) and receive command 


*Now we will move to the Subscriber which will subscribe to the topic and ingest those published telemetry* `/services/ingestor/main.go`


```
package main

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://172.20.1.10:1883").SetClientID("ingestor-service")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT connect error: %v", token.Error())
	}
	fmt.Println("Connected to MQTT broker")

	messageHandler := func(c mqtt.Client, m mqtt.Message) {
		fmt.Printf("Received on %s: %s\n", m.Topic(), string(m.Payload()))
		// TODO: parse JSON, persist to DB
	}

	if token := client.Subscribe("devices/+/telemetry", 0, messageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("Subscribe error: %v", token.Error())
	}
	fmt.Println("Subscribed to devices/+/telemetry")
	select {} // block forever
}
```



##### Step 2
- Persist telemetry to Postgres (schema + simple INSERTs).
    
- Build a small REST API (Go) that serves device state and can send commands (via MQTT publish).
    
- Implement device provisioning and secure device auth (tokens & TLS).
    
- Add a simple Rules engine (Python) that subscribes to telemetry and triggers commands (e.g., turn on lights when motion).
    
- Replace localhost comms with robust internal messaging (gRPC or NATS) and containerize each service + CI.
    
- Add monitoring, logging, and deploy to Kubernetes.


**Persist telemetry to Postgres (schema + simple INSERTs).**


- create a simple DB schema (table for telemetry),
    
- update the Go ingestor to insert rows,
    
- test that messages are stored.


We will be creating the SQL Migration schema in the server (place this schema where your Docker image for the Mosquitto and Postgres lives)

```
CREATE TABLE IF NOT EXISTS telemetry (
    id SERIAL PRIMARY KEY,
    device_id TEXT NOT NULL,
    ts BIGINT NOT NULL,          -- Unix timestamp from device
    temp DOUBLE PRECISION,       -- example telemetry value
    state TEXT,
    raw JSONB,                   -- full JSON payload (flexible for future)
    created_at TIMESTAMPTZ DEFAULT now()
);
```

Follow the below command to copy the schema file to the `Postgres` container


`docker cp schema.sql <docker-image-name>:/schema.sql`

`docker exec -it home-automation-postgres-1 psql -U ha -d ha_db -f /schema.sqo`



### REST API


- Expose endpoints to **query telemetry** from Postgres.
    
- Expose an endpoint to **send commands** to devices (publish to MQTT).
    
- Run in Docker so it can join the same network as broker + DB.