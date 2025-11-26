package service

import (
	"errors"
	"event-ingestion/model"
	"event-ingestion/repository"
	"strings"
	"time"
)

var (
	ErrMissingFields    = errors.New("missing required fields")
	ErrFutureTimestamp  = errors.New("timestamp cannot be in the future")
	ErrDuplicateEvent   = errors.New("duplicate event")
)

type EventService struct {
	repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) CreateEvent(event model.Event) error {
	if event.EventName == "" || event.UserID == "" || event.Timestamp == 0 {
		return ErrMissingFields
	}

	if event.Timestamp > time.Now().Unix() {
		return ErrFutureTimestamp
	}

	err := s.repo.Insert(event)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return ErrDuplicateEvent
		}
		return err
	}

	return nil
}

func (s *EventService) GetMetrics(eventName string, from, to int64, groupBy string) (model.MetricsResponse, error) {
	if eventName == "" {
		return model.MetricsResponse{}, ErrMissingFields
	}

	return s.repo.GetMetrics(eventName, from, to, groupBy)
}
