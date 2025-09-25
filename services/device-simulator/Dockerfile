FROM python:3.11-slim
WORKDIR /app
COPY device_simulator.py /app/
RUN pip install paho-mqtt
CMD ["python", "device_simulator.py", "--device-id", "light1", "--host", "mqtt"]
