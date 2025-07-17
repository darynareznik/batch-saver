package service

import (
	"batch-saver/internal/models"
	"batch-saver/internal/service/mock"
	"context"
	"go.uber.org/mock/gomock"
	"sync"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	tt := map[string]struct {
		isContextCanceled bool
		events            []models.Event
		expectedBatches   [][]models.Event
		batchSize         int
		flushTimeout      time.Duration
		sleepAfterSend    time.Duration
	}{
		"flush_on_batch_size": {
			events: []models.Event{
				{
					ID:      "1",
					GroupID: "1",
					Data:    []byte("data1"),
				},
				{
					ID:      "2",
					GroupID: "1",
					Data:    []byte("data2"),
				},
				{
					ID:      "3",
					GroupID: "1",
					Data:    []byte("data3"),
				},
			},
			expectedBatches: [][]models.Event{
				{
					{
						ID:      "1",
						GroupID: "1",
						Data:    []byte("data1"),
					},
					{
						ID:      "2",
						GroupID: "1",
						Data:    []byte("data2"),
					},
					{
						ID:      "3",
						GroupID: "1",
						Data:    []byte("data3"),
					},
				},
			},
			batchSize:      3,
			flushTimeout:   50 * time.Millisecond,
			sleepAfterSend: 30 * time.Millisecond,
		},
		"flush_multiple_groups": {
			events: []models.Event{
				{
					ID:      "1",
					GroupID: "1",
					Data:    []byte("data1"),
				},
				{
					ID:      "2",
					GroupID: "2",
					Data:    []byte("data2"),
				},
				{
					ID:      "3",
					GroupID: "2",
					Data:    []byte("data3"),
				},
				{
					ID:      "4",
					GroupID: "1",
					Data:    []byte("data4"),
				},
			},
			expectedBatches: [][]models.Event{
				{
					{
						ID:      "1",
						GroupID: "1",
						Data:    []byte("data1"),
					},
					{
						ID:      "4",
						GroupID: "1",
						Data:    []byte("data4"),
					},
				},
				{
					{
						ID:      "2",
						GroupID: "2",
						Data:    []byte("data2"),
					},
					{
						ID:      "3",
						GroupID: "2",
						Data:    []byte("data3"),
					},
				},
			},
			batchSize:      2,
			flushTimeout:   50 * time.Millisecond,
			sleepAfterSend: 30 * time.Millisecond,
		},
		"flush_on_timeout": {
			events: []models.Event{
				{
					ID:      "1",
					GroupID: "1",
					Data:    []byte("data1"),
				},
				{
					ID:      "2",
					GroupID: "1",
					Data:    []byte("data2"),
				},
			},
			expectedBatches: [][]models.Event{
				{
					{
						ID:      "1",
						GroupID: "1",
						Data:    []byte("data1"),
					},
					{
						ID:      "2",
						GroupID: "1",
						Data:    []byte("data2"),
					},
				},
			},
			batchSize:      3,
			flushTimeout:   50 * time.Millisecond,
			sleepAfterSend: 100 * time.Millisecond,
		},
		"flush_on_shutdown": {
			isContextCanceled: true,
			events: []models.Event{
				{
					ID:      "1",
					GroupID: "1",
					Data:    []byte("data1"),
				},
			},
			expectedBatches: [][]models.Event{
				{
					{
						ID:      "1",
						GroupID: "1",
						Data:    []byte("data1"),
					},
				},
			},
			batchSize:      3,
			flushTimeout:   50 * time.Millisecond,
			sleepAfterSend: 30 * time.Millisecond,
		},
	}

	for testName, testCase := range tt {
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			wg := new(sync.WaitGroup)

			repo := mock.NewMockrepository(ctrl)
			for _, batch := range testCase.expectedBatches {
				repo.EXPECT().
					Save(gomock.Eq(context.Background()), gomock.Eq(batch)).
					Times(1).
					Return(nil)
			}

			s := NewService(ctx, wg, repo, Config{
				BatchMaxSize:      testCase.batchSize,
				BatchFlushTimeout: testCase.flushTimeout,
			})

			send(s, testCase.events)

			if testCase.isContextCanceled {
				cancel()
			}

			time.Sleep(testCase.sleepAfterSend)
		})
	}
}

func TestService_NoOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	wg := new(sync.WaitGroup)

	repo := mock.NewMockrepository(ctrl)
	repo.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Times(0)

	NewService(ctx, wg, repo, Config{
		BatchMaxSize:      3,
		BatchFlushTimeout: 50 * time.Millisecond,
	})

	time.Sleep(30 * time.Millisecond)
}

func send(s *service, events []models.Event) {
	for _, e := range events {
		s.Save(e)
	}
}
