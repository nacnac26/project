package repository

import (
	"database/sql"
	"encoding/json"
	"event-ingestion/model"
	"strconv"

	"github.com/lib/pq"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Insert(event model.Event) error {
	metadataJSON, _ := json.Marshal(event.Metadata)
	_, err := r.db.Exec(
		`INSERT INTO events (event_name, channel, campaign_id, user_id, timestamp, tags, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		event.EventName, event.Channel, event.CampaignID, event.UserID, event.Timestamp, pq.Array(event.Tags), string(metadataJSON),
	)
	return err
}

func (r *EventRepository) GetMetrics(eventName string, from, to int64, groupBy string) (model.MetricsResponse, error) {
	var metrics model.MetricsResponse
	metrics.EventName = eventName

	where := "event_name = $1"
	args := []interface{}{eventName}

	if from > 0 {
		args = append(args, from)
		where += " AND timestamp >= $" + strconv.Itoa(len(args))
	}
	if to > 0 {
		args = append(args, to)
		where += " AND timestamp <= $" + strconv.Itoa(len(args))
	}

	query := "SELECT COUNT(*), COUNT(DISTINCT user_id) FROM events WHERE " + where
	err := r.db.QueryRow(query, args...).Scan(&metrics.TotalCount, &metrics.UniqueUsers)
	if err != nil {
		return metrics, err
	}

	switch groupBy {
	case "channel":
		rows, err := r.db.Query("SELECT channel, COUNT(*), COUNT(DISTINCT user_id) FROM events WHERE "+where+" GROUP BY channel", args...)
		if err != nil {
			return metrics, err
		}
		defer rows.Close()
		for rows.Next() {
			var cm model.ChannelMetric
			if err := rows.Scan(&cm.Channel, &cm.TotalCount, &cm.UniqueUsers); err != nil {
				continue
			}
			metrics.ByChannel = append(metrics.ByChannel, cm)
		}
	case "daily":
		rows, err := r.db.Query("SELECT TO_CHAR(TO_TIMESTAMP(timestamp), 'YYYY-MM-DD'), COUNT(*), COUNT(DISTINCT user_id) FROM events WHERE "+where+" GROUP BY TO_CHAR(TO_TIMESTAMP(timestamp), 'YYYY-MM-DD') ORDER BY 1", args...)
		if err != nil {
			return metrics, err
		}
		defer rows.Close()
		for rows.Next() {
			var tm model.TimeMetric
			if err := rows.Scan(&tm.Period, &tm.TotalCount, &tm.UniqueUsers); err != nil {
				continue
			}
			metrics.ByTime = append(metrics.ByTime, tm)
		}
	case "hourly":
		rows, err := r.db.Query("SELECT TO_CHAR(TO_TIMESTAMP(timestamp), 'YYYY-MM-DD HH24:00'), COUNT(*), COUNT(DISTINCT user_id) FROM events WHERE "+where+" GROUP BY TO_CHAR(TO_TIMESTAMP(timestamp), 'YYYY-MM-DD HH24:00') ORDER BY 1", args...)
		if err != nil {
			return metrics, err
		}
		defer rows.Close()
		for rows.Next() {
			var tm model.TimeMetric
			if err := rows.Scan(&tm.Period, &tm.TotalCount, &tm.UniqueUsers); err != nil {
				continue
			}
			metrics.ByTime = append(metrics.ByTime, tm)
		}
	}

	return metrics, nil
}
