package worker_pool_factory

import (
	"context"
	"log"
	"sync"
)

// PayloadFunc is a generic payload func type
type PayloadFunc[T, R any] func(T) (R, error)

// Result is passed through results chan.
type Result[R any] struct {
	Result R
	Error  error
}

type WorkerPoolFactory[T, R any] struct {

	// Payload is func doing the job.
	Payload PayloadFunc[T, R]

	// NWorkers is 1 by default.
	NWorkers uint

	// Tasks is channel for tasks.
	//Workers stop when Tasks closed.
	Tasks <-chan T

	// Results closes when workers stopped.
	Results chan<- Result[R]

	// Logger defaults to DefaultLogger(LoggerPrefix)
	Logger       *log.Logger
	LoggerPrefix string

	// You can provide Ctx for cancelling from outside
	Ctx context.Context

	init bool
	wg   *sync.WaitGroup
}

func (w *WorkerPoolFactory[T, R]) Init() {
	if w.Logger == nil {
		w.Logger = DefaultLogger(w.LoggerPrefix)
	}
	w.Logger.Println("initializing worker pool")
	if w.Payload == nil {
		panic("no Payload provided")
	}
	w.wg = new(sync.WaitGroup)
	if w.Tasks == nil {
		panic("Tasks channel is not  provided")
	}
	if w.Results == nil {
		panic("Results channel is not  provided")
	}
	if w.NWorkers == 0 {
		w.NWorkers = 1
	}
	if w.Ctx == nil {
		w.Ctx = context.Background()
	}

	w.init = true
}

func (w WorkerPoolFactory[T, R]) Run() {
	if !w.init {
		w.Init()
	}
	w.startWorkers()
	go w.closeResultsAfterFinish()
}

func (w WorkerPoolFactory[T, R]) closeResultsAfterFinish() {
	w.wg.Wait()
	close(w.Results)
}

func (w WorkerPoolFactory[T, R]) worker() {
Loop:
	for {
		select {
		case <-w.Ctx.Done():
			w.Logger.Println("cancelling worker")
			break Loop
		case task, ok := <-w.Tasks:
			if !ok {
				w.Logger.Println("task chan is closed")
				break Loop
			}
			w.Logger.Println("running task:", task)
			r, err := w.Payload(task)
			//w.Logger.Println("returning Result:", task)
			w.Results <- Result[R]{Result: r, Error: err}
			w.Logger.Println("task done:", task)
		}
	}
	w.wg.Done()
}

func (w WorkerPoolFactory[T, R]) startWorkers() {
	w.Logger.Println("starting workers")
	w.wg.Add(int(w.NWorkers))
	for i := uint(0); i < w.NWorkers; i++ {
		w.Logger.Println("starting worker", i)
		go w.worker()
	}
}
