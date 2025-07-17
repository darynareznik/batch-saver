package service

import (
	"batch-saver/internal/models"
	"context"
	"github.com/rs/zerolog/log"
	"sync"
)

type writerPool struct {
	wg *sync.WaitGroup

	repo repository
	pool chan struct{}
}

func NewWriterPool(wg *sync.WaitGroup, repo repository, maxConcurrency int) *writerPool {
	return &writerPool{
		wg:   wg,
		repo: repo,
		pool: make(chan struct{}, maxConcurrency),
	}
}

func (p *writerPool) Save(ctx context.Context, e []models.Event) error {
	// adhere to max concurrent writes
	p.pool <- struct{}{}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		if err := p.repo.Save(ctx, e); err != nil {
			log.Err(err).Msg("Error saving events")
		}
		<-p.pool
	}()

	return nil
}
