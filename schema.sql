CREATE TABLE IF NOT EXISTS telemetry (
    id SERIAL PRIMARY KEY,
    device_id TEXT NOT NULL,
    ts BIGINT NOT NULL,          -- Unix timestamp from device
    temp DOUBLE PRECISION,       -- example telemetry value
    state TEXT,
    raw JSONB,                   -- full JSON payload (flexible for future)
    created_at TIMESTAMPTZ DEFAULT now()
);

