
**Device / Edge agent / Simulator** — _Python_: fast for prototypes, many libs (paho-mqtt).

**Device ingestion / core event consumer** — _Go_: efficient concurrency for ingesting many device messages.

**API Gateway / REST API** — _Go_ (or Go + gRPC) for performance.

**Rules engine / automation scripts** — _Python_ (easy to write DSLs, plugins)

**Message broker** — _MQTT (Mosquitto)_ for device-to-cloud; later add _NATS_ or _Kafka_ for internal microservice events.

**Persistence** — _Postgres_ for state; _InfluxDB/Timescale_ if heavy time-series.

**Dev-run orchestration** — Docker Compose → later Kubernetes.

**Auth** — JWT for MVP, Keycloak/OAuth2 for bigger deployments.

**Observability** — Prometheus + Grafana + OpenTelemetry for services.