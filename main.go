package main

import (
	"database/sql"
	"event-ingestion/handler"
	"event-ingestion/repository"
	"event-ingestion/service"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	db, err := sql.Open("postgres", "host="+host+" port=5432 user=events_user password=events_password dbname=events_db sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewEventRepository(db)
	svc := service.NewEventService(repo)
	eventHandler := handler.NewEventHandler(svc)
	metricsHandler := handler.NewMetricsHandler(svc)

	http.HandleFunc("/events", eventHandler.Handle)
	http.HandleFunc("/metrics", metricsHandler.Handle)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
