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
curl -X POST http://localhost:8080/events -d '{"event_name":"product_view","channel":"web", "user_id":"user_123", "campaign_id": "cmp_987"
, "tags": ["electronics", "homepage", "flash_sale"], "timestamp":1723475612, "metadata": {"product_id": "prod-789", "price": 129.99, "currency": "TRY", "referrer": "google"}}'

curl "http://localhost:8080/metrics?event_name=product_view"
curl "http://localhost:8080/metrics?event_name=product_view&from=1723475000&to=1723476000"
curl "http://localhost:8080/metrics?event_name=product_view&group_by=hourly"
```
