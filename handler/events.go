package handler

import (
	"encoding/json"
	"errors"
	"event-ingestion/model"
	"event-ingestion/service"
	"log"
	"net/http"
)

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func (h *EventHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Println("POST /events")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event model.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.svc.CreateEvent(event)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingFields):
			http.Error(w, "Missing required fields", http.StatusBadRequest)
		case errors.Is(err, service.ErrFutureTimestamp):
			http.Error(w, "Timestamp cannot be in the future", http.StatusBadRequest)
		case errors.Is(err, service.ErrBufferFull):
			http.Error(w, "Server busy, try again later", http.StatusServiceUnavailable)
		default:
			log.Println("Error:", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"ok"}`))
}
