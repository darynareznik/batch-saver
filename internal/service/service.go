package service

import (
	"batch-saver/internal/models"
	"batch-saver/internal/service/batching"
	"context"
	"sync"
	"time"
)

type repository interface {
	Save(context.Context, []models.Event) error
}

type Config struct {
	BatchMaxSize      int
	BatchFlushTimeout time.Duration
}

type service struct {
	events chan models.Event
	repo   repository
}

func NewService(ctx context.Context, wg *sync.WaitGroup, repo repository, cfg Config) *service {
	events := make(chan models.Event)
	// batcher runs in the background, groups records into batches and flushes batches into the database
	batching.NewBatcher(ctx, wg, repo, events, cfg.BatchMaxSize, cfg.BatchFlushTimeout)

	return &service{
		events: events,
		repo:   repo,
	}
}

func (s *service) Save(e models.Event) {
	s.events <- e
}
