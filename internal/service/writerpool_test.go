package service

import (
	"batch-saver/internal/models"
	"batch-saver/internal/service/mock"
	"context"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"sync"
	"testing"
	"time"
)

func TestWriterPool(t *testing.T) {
	tt := map[string]struct {
		numBatches     int
		maxConcurrency int
	}{
		"one": {
			numBatches:     5,
			maxConcurrency: 1,
		},
		"two": {
			numBatches:     3,
			maxConcurrency: 2,
		},
	}

	for testName, testCase := range tt {
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var mu sync.Mutex
			active := 0
			maxActive := 0
			wg := new(sync.WaitGroup)

			repo := mock.NewMockrepository(ctrl)
			repo.EXPECT().
				Save(gomock.Any(), gomock.Any()).
				Times(testCase.numBatches).
				DoAndReturn(func(_ context.Context, _ []models.Event) error {
					mu.Lock()
					active++
					if active > maxActive {
						maxActive = active
					}
					mu.Unlock()

					time.Sleep(50 * time.Millisecond)

					mu.Lock()
					active--
					mu.Unlock()
					return nil
				})

			pool := NewWriterPool(wg, repo, testCase.maxConcurrency)
			for i := 0; i < testCase.numBatches; i++ {
				pool.Save(context.Background(), []models.Event{{ID: "1"}})
			}
			wg.Wait()

			require.LessOrEqual(t, maxActive, testCase.maxConcurrency)
		})
	}
}
