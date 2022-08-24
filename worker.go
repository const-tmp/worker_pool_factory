package worker_pool_factory

import (
	"context"
	"log"
)

// PayloadFunc is a generic payload func type
type PayloadFunc[T, R any] func(T) (R, error)

type WorkerPoolFactory[T, R any] struct {

	// Payload is func doing the job.
	Payload PayloadFunc[T, R]

	// NWorkers is 1 by default.
	NWorkers uint

	// TaskQueueSize is task chan buffer size, TaskQueueSize by default.
	TaskQueueSize uint

	// ResultQueueSize is result chan buffer size, NWorkers by default.
	ResultQueueSize uint

	// Logger defaults to DefaultLogger(LoggerPrefix)
	Logger       *log.Logger
	LoggerPrefix string

	// You can provide TaskCtx & ResultCtx for cancelling from outside
	TaskCtx   context.Context
	ResultCtx context.Context
}

func (w *WorkerPoolFactory[T, R]) Create() *workerPool[T, R] {
	return New[T, R](w)
}
