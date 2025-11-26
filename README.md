# Event Ingestion Service

A simple event tracking API built with Go and PostgreSQL.

## How to Run

### Docker Compose

```bash
docker-compose up --build
```

## API Endpoints

### POST /events

Send an event to the system.

```bash
curl -X POST http://localhost:8080/events \
  -d '{"event_name":"product_view","channel":"web","user_id":"user_123","timestamp":1723475612}'
```

Required fields: `event_name`, `user_id`, `timestamp`

### GET /metrics

Get event metrics. The `event_name` parameter is required.

```bash
# Basic query (served from cache)
curl "http://localhost:8080/metrics?event_name=product_view"

# With time range
curl "http://localhost:8080/metrics?event_name=product_view&from=1723475000&to=1723476000"

# Group by channel
curl "http://localhost:8080/metrics?event_name=product_view&group_by=channel"

# Group by daily
curl "http://localhost:8080/metrics?event_name=product_view&group_by=daily"
```

## Design Decisions

- **PostgreSQL**: I work with PostgreSQL so I chose it.
- **Async Write**: Events are batched before writing. Faster response.
- **Metrics Cache**: Simple queries use cache. Refreshes every 20 seconds.
- **Idempotency**: Unique constraint prevents duplicates.

## Trade-offs

- Filtered queries hit the database.
- Metrics can be 20 seconds stale.
- No bulk endpoint.

## What I Would Add

- Swagger
- More detailed error responses
- Health check endpoint
- Rate Limiting
- Message queue
