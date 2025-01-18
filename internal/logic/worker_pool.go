package logic

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/Gregmus2/sync-service/internal/adapters"
	"github.com/Gregmus2/sync-service/internal/common"
	"github.com/Gregmus2/sync-service/internal/interceptors"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
)

type workerPool struct {
	in     chan job
	repo   adapters.Repository
	logger *logrus.Entry
}

type job struct {
	stream  proto.SyncService_SyncDataServer
	wg      *sync.WaitGroup
	groupID string
}

func NewWorkerPool(cfg *common.Config, repo adapters.Repository, logger *logrus.Entry) WorkerPool {
	in := make(chan job, cfg.WorkerPoolBuffer)
	pool := &workerPool{
		in:     in,
		repo:   repo,
		logger: logger,
	}

	for i := 0; i < cfg.Workers; i++ {
		go pool.worker(in)
	}

	return pool
}

func (wp workerPool) Add(stream proto.SyncService_SyncDataServer, groupID string) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	wp.in <- job{
		stream:  stream,
		wg:      &wg,
		groupID: groupID,
	}

	return &wg
}

func (wp workerPool) worker(in chan job) {
	for j := range in {
		deviceToken := j.stream.Context().Value(interceptors.ContextDeviceToken).(string)

		for {
			operations, err := j.stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				wp.logger.WithError(err).Error("failed to receive data")

				break
			}

			err = wp.repo.InsertData(deviceToken, j.groupID, operations.Operations)
			if err != nil {
				wp.logger.WithError(err).Error("failed to insert data")

				break
			}
		}

		j.wg.Done()
	}
}
