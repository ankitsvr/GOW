
#!/usr/bin/env python3
"""
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
