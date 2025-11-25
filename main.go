package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

var db *sql.DB

type Event struct {
	EventName  string                 `json:"event_name"`
	Channel    string                 `json:"channel"`
	CampaignID string                 `json:"campaign_id"`
	UserID     string                 `json:"user_id"`
	Timestamp  int64                  `json:"timestamp"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
}

func main() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=events_user password=events_password dbname=events_db sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/events", handleEvents)
	http.HandleFunc("/metrics", handleMetrics)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	log.Println("POST /events")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validation
	if event.EventName == "" || event.UserID == "" || event.Timestamp == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Convert metadata to JSON string
	metadataJSON, _ := json.Marshal(event.Metadata)

	// Insert to DB
	_, err := db.Exec(`INSERT INTO events (event_name, channel, campaign_id, user_id, timestamp, tags, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		event.EventName, event.Channel, event.CampaignID, event.UserID, event.Timestamp, pq.Array(event.Tags), string(metadataJSON))

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			http.Error(w, "Duplicate event", http.StatusConflict)
			return
		}
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

type MetricsResponse struct {
	EventName   string          `json:"event_name"`
	TotalCount  int             `json:"total_count"`
	UniqueUsers int             `json:"unique_users"`
	ByChannel   []ChannelMetric `json:"by_channel,omitempty"`
	ByTime      []TimeMetric    `json:"by_time,omitempty"`
}

type ChannelMetric struct {
	Channel     string `json:"channel"`
	TotalCount  int    `json:"total_count"`
	UniqueUsers int    `json:"unique_users"`
}

type TimeMetric struct {
	Period      string `json:"period"`
	TotalCount  int    `json:"total_count"`
	UniqueUsers int    `json:"unique_users"`
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /metrics")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventName := r.URL.Query().Get("event_name")
	if eventName == "" {
		http.Error(w, "event_name is required", http.StatusBadRequest)
		return
	}

	where := "event_name = $1"
	args := []interface{}{eventName}

	if from := r.URL.Query().Get("from"); from != "" {
		fromInt, err := strconv.ParseInt(from, 10, 64)
		if err != nil {
			http.Error(w, "Invalid from parameter", http.StatusBadRequest)
			return
		}
		args = append(args, fromInt)
		where += " AND timestamp >= $" + strconv.Itoa(len(args))
	}
	if to := r.URL.Query().Get("to"); to != "" {
		toInt, err := strconv.ParseInt(to, 10, 64)
		if err != nil {
			http.Error(w, "Invalid to parameter", http.StatusBadRequest)
			return
		}
		args = append(args, toInt)
		where += " AND timestamp <= $" + strconv.Itoa(len(args))
	}

	var metrics MetricsResponse
	metrics.EventName = eventName

	query := "SELECT COUNT(*), COUNT(DISTINCT user_id) FROM events WHERE " + where
	err := db.QueryRow(query, args...).Scan(&metrics.TotalCount, &metrics.UniqueUsers)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	// Group by
	switch r.URL.Query().Get("group_by") {
	case "channel":
		rows, err := db.Query("SELECT channel, COUNT(*), COUNT(DISTINCT user_id) FROM events WHERE "+where+" GROUP BY channel", args...)
		if err != nil {
			log.Println("DB error:", err)
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var cm ChannelMetric
			if err := rows.Scan(&cm.Channel, &cm.TotalCount, &cm.UniqueUsers); err != nil {
				log.Println("Scan error:", err)
				continue
			}
			metrics.ByChannel = append(metrics.ByChannel, cm)
		}
	case "daily":
		rows, err := db.Query("SELECT TO_CHAR(TO_TIMESTAMP(timestamp), 'YYYY-MM-DD'), COUNT(*), COUNT(DISTINCT user_id) FROM events WHERE "+where+" GROUP BY TO_CHAR(TO_TIMESTAMP(timestamp), 'YYYY-MM-DD')", args...)
		if err != nil {
			log.Println("DB error:", err)
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var tm TimeMetric
			if err := rows.Scan(&tm.Period, &tm.TotalCount, &tm.UniqueUsers); err != nil {
				log.Println("Scan error:", err)
				continue
			}
			metrics.ByTime = append(metrics.ByTime, tm)
		}
	case "hourly":
		rows, err := db.Query("SELECT TO_CHAR(TO_TIMESTAMP(timestamp), 'YYYY-MM-DD HH24:00'), COUNT(*), COUNT(DISTINCT user_id) FROM events WHERE "+where+" GROUP BY TO_CHAR(TO_TIMESTAMP(timestamp), 'YYYY-MM-DD HH24:00')", args...)
		if err != nil {
			log.Println("DB error:", err)
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var tm TimeMetric
			if err := rows.Scan(&tm.Period, &tm.TotalCount, &tm.UniqueUsers); err != nil {
				log.Println("Scan error:", err)
				continue
			}
			metrics.ByTime = append(metrics.ByTime, tm)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
