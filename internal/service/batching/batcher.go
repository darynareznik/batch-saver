package batching

import (
	"batch-saver/internal/models"
	"context"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type repository interface {
	Save(context.Context, []models.Event) error
}

type batcher struct {
	events       <-chan models.Event
	batchSize    int
	flushTimeout time.Duration
	flushFn      func(events []models.Event)

	sync.RWMutex
	batches map[string]*batch
}

func NewBatcher(ctx context.Context, wg *sync.WaitGroup, repo repository, events <-chan models.Event, batchSize int, flushTimeout time.Duration) *batcher {
	flushFn := func(e []models.Event) {
		if len(e) == 0 {
			return
		}
		repo.Save(context.Background(), e)
	}

	b := &batcher{
		events:       events,
		batchSize:    batchSize,
		flushTimeout: flushTimeout,
		flushFn:      flushFn,
		batches:      make(map[string]*batch),
	}

	wg.Add(1)
	go b.run(ctx, wg)

	return b
}

func (b *batcher) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			cancel()
			return
		case e := <-b.events:
			bat, ok := b.getBatch(e.GroupID)
			if !ok {
				b.addBatch(e.GroupID, newBatch(cancelCtx, wg, b.batchSize, b.flushTimeout, b.flushFn, func() { b.deleteBatch(e.GroupID) }))
			}

			bat, _ = b.getBatch(e.GroupID)
			bat.e <- e
		}
	}
}

func (b *batcher) getBatch(id string) (*batch, bool) {
	b.RLock()
	defer b.RUnlock()

	bat, ok := b.batches[id]
	return bat, ok
}

func (b *batcher) addBatch(id string, bat *batch) {
	b.Lock()
	defer b.Unlock()

	b.batches[id] = bat
	return
}

func (b *batcher) deleteBatch(id string) {
	b.Lock()
	defer b.Unlock()

	delete(b.batches, id)
}

type batch struct {
	buf          []models.Event
	maxSize      int
	flushTimeout time.Duration
	e            chan models.Event
	flush        func([]models.Event)
	cleanup      func()
}

func newBatch(ctx context.Context, wg *sync.WaitGroup, size int, flushTimeout time.Duration, flushFn func(events []models.Event), cleanupFn func()) *batch {
	b := &batch{
		buf:          make([]models.Event, 0, size),
		maxSize:      size,
		flushTimeout: flushTimeout,
		e:            make(chan models.Event),
		flush:        flushFn,
		cleanup:      cleanupFn,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		timer := time.NewTimer(flushTimeout)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				b.cleanup()
				b.flush(b.buf)
				return
			case e := <-b.e:
				b.buf = append(b.buf, e)
				if len(b.buf) == b.maxSize {
					log.Debug().
						Str("reason", "size").
						Any("batch", b.buf).
						Msg("Flushing batch")
					b.cleanup()
					b.flush(b.buf)
					return
				}

				if !timer.Stop() {
					<-timer.C // drain
				}
				timer.Reset(b.flushTimeout)

			case <-timer.C:
				log.Debug().
					Str("reason", "timeout").
					Any("batch", b.buf).
					Msg("Flushing batch")
				b.cleanup()
				b.flush(b.buf)
				return
			}
		}
	}()

	return b
}
