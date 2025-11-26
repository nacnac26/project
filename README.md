# Event Ingestion Service

## Run with Docker Compose

```bash
docker-compose up --build
```

## Run Locally

```bash
# PostgreSQL
docker build -t events-postgres ./docker/postgres
docker run -d -p 5432:5432 -e POSTGRES_DB=events_db -e POSTGRES_USER=events_user -e POSTGRES_PASSWORD=events_password events-postgres

# API
make build
make run
```

## Examples

### Post Event

```bash
curl -X POST http://localhost:8080/events -d '{"event_name":"product_view","channel":"web","user_id":"user_123","timestamp":1723475612}'
```

### Get Metrics

```bash
# Basic
curl "http://localhost:8080/metrics?event_name=product_view"

# With time range
curl "http://localhost:8080/metrics?event_name=product_view&from=1723475000&to=1723476000"

# Group by channel
curl "http://localhost:8080/metrics?event_name=product_view&group_by=channel"

# Group by daily
curl "http://localhost:8080/metrics?event_name=product_view&group_by=daily"

# Group by hourly
curl "http://localhost:8080/metrics?event_name=product_view&group_by=hourly"
```
