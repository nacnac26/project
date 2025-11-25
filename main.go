package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

type MetricsResponse struct {
	EventName   string `json:"event_name"`
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

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var metrics MetricsResponse
	metrics.EventName = eventName

	query := `SELECT COUNT(*), COUNT(DISTINCT user_id) FROM events WHERE event_name = $1`
	args := []interface{}{eventName}

	if fromStr != "" {
		from, _ := strconv.ParseInt(fromStr, 10, 64)
		query += ` AND timestamp >= $2`
		args = append(args, from)
	}
	if toStr != "" {
		to, _ := strconv.ParseInt(toStr, 10, 64)
		query += ` AND timestamp <= $` + strconv.Itoa(len(args)+1)
		args = append(args, to)
	}

	err := db.QueryRow(query, args...).Scan(&metrics.TotalCount, &metrics.UniqueUsers)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
