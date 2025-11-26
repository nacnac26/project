package service

import (
	"errors"
	"event-ingestion/model"
	"event-ingestion/repository"
	"log"
	"time"
)

var (
	ErrMissingFields   = errors.New("missing required fields")
	ErrFutureTimestamp = errors.New("timestamp cannot be in the future")
	ErrBufferFull      = errors.New("buffer full, try again later")
)

type EventService struct {
	repo    *repository.EventRepository
	eventCh chan model.Event
}

func NewEventService(repo *repository.EventRepository) *EventService {
	svc := &EventService{
		repo:    repo,
		eventCh: make(chan model.Event, 10000),
	}
	svc.startWorkers(20)
	return svc
}

func (s *EventService) startWorkers(count int) {
	for i := 0; i < count; i++ {
		go s.worker()
	}
}

func (s *EventService) worker() {
	for event := range s.eventCh {
		err := s.repo.Insert(event)
		if err != nil {
			log.Println("Worker insert error:", err)
		}
	}
}

func (s *EventService) CreateEvent(event model.Event) error {
	if event.EventName == "" || event.UserID == "" || event.Timestamp == 0 {
		return ErrMissingFields
	}

	if event.Timestamp > time.Now().Unix() {
		return ErrFutureTimestamp
	}

	select {
	case s.eventCh <- event:
		return nil
	default:
		return ErrBufferFull
	}
}

func (s *EventService) GetMetrics(eventName string, from, to int64, groupBy string) (model.MetricsResponse, error) {
	if eventName == "" {
		return model.MetricsResponse{}, ErrMissingFields
	}

	return s.repo.GetMetrics(eventName, from, to, groupBy)
}
