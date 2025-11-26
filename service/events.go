package service

import (
	"errors"
	"event-ingestion/model"
	"event-ingestion/repository"
	"log"
	"sync"
	"time"
)

const (
	batchSize    = 100
	batchTimeout = 100 * time.Millisecond
)

var (
	ErrMissingFields   = errors.New("missing required fields")
	ErrFutureTimestamp = errors.New("timestamp cannot be in the future")
	ErrBufferFull      = errors.New("buffer full, try again later")
)

type EventService struct {
	repo         *repository.EventRepository
	eventCh      chan model.Event
	metricsCache map[string]model.MetricsResponse
	cacheMu      sync.RWMutex
}

func NewEventService(repo *repository.EventRepository) *EventService {
	svc := &EventService{
		repo:         repo,
		eventCh:      make(chan model.Event, 10000),
		metricsCache: make(map[string]model.MetricsResponse),
	}
	svc.startWorkers(20)
	go svc.refreshMetricsLoop()
	return svc
}

func (s *EventService) startWorkers(count int) {
	for i := 0; i < count; i++ {
		go s.batchWorker()
	}
}

func (s *EventService) batchWorker() {
	batch := make([]model.Event, 0, batchSize)
	timer := time.NewTimer(batchTimeout)

	for {
		select {
		case event := <-s.eventCh:
			batch = append(batch, event)
			if len(batch) >= batchSize {
				s.flushBatch(batch)
				batch = batch[:0]
				timer.Reset(batchTimeout)
			}
		case <-timer.C:
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
			timer.Reset(batchTimeout)
		}
	}
}

func (s *EventService) flushBatch(batch []model.Event) {
	if err := s.repo.InsertBatch(batch); err != nil {
		log.Println("Batch insert error:", err)
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

	if from == 0 && to == 0 && groupBy == "" {
		s.cacheMu.RLock()
		if cached, ok := s.metricsCache[eventName]; ok {
			log.Println("Serving metrics for", eventName, "from cache")
			s.cacheMu.RUnlock()
			return cached, nil
		}
		s.cacheMu.RUnlock()
	}

	return s.repo.GetMetrics(eventName, from, to, groupBy)
}

func (s *EventService) refreshMetricsLoop() {
	ticker := time.NewTicker(20 * time.Second)
	s.refreshMetrics()
	for range ticker.C {
		s.refreshMetrics()
	}
}

func (s *EventService) refreshMetrics() {
	rows, err := s.repo.GetAllEventNames()
	if err != nil {
		log.Println("Failed to get event names:", err)
		return
	}

	for _, eventName := range rows {
		metrics, err := s.repo.GetMetrics(eventName, 0, 0, "")
		if err != nil {
			continue
		}
		s.cacheMu.Lock()
		s.metricsCache[eventName] = metrics
		s.cacheMu.Unlock()
	}
}
