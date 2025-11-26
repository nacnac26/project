package handler

import (
	"encoding/json"
	"event-ingestion/service"
	"log"
	"net/http"
	"strconv"
)

type MetricsHandler struct {
	svc *service.EventService
}

func NewMetricsHandler(svc *service.EventService) *MetricsHandler {
	return &MetricsHandler{svc: svc}
}

func (h *MetricsHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	var from, to int64
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		var err error
		from, err = strconv.ParseInt(fromStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid from parameter", http.StatusBadRequest)
			return
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		var err error
		to, err = strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid to parameter", http.StatusBadRequest)
			return
		}
	}

	groupBy := r.URL.Query().Get("group_by")

	metrics, err := h.svc.GetMetrics(eventName, from, to, groupBy)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
