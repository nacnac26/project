# Event Ingestion Service

## Run PostgreSQL

```bash
docker build -t events-postgres ./docker/postgres
docker run -d -p 5432:5432 -e POSTGRES_DB=events_db -e POSTGRES_USER=events_user -e POSTGRES_PASSWORD=events_password events-postgres
```

## Build & Run

```bash
make build
make run
```

## Examples

```bash
curl -X POST http://localhost:8080/events -d '{"event_name":"product_view","user_id":"user_123","timestamp":1723475612}'

curl http://localhost:8080/metrics
```
