CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    event_name VARCHAR(255) NOT NULL,
    channel VARCHAR(255),
    campaign_id VARCHAR(255),
    user_id VARCHAR(255) NOT NULL,
    timestamp BIGINT NOT NULL,
    tags TEXT[],
    metadata TEXT
);