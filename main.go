package main

import (
	"log"
	"net/http"
)

func main() {
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /metrics")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"metrics":[]}`))
}
