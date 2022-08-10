package ezwp

import (
	"log"
	"sync"
)

const defaultNWorkers = 1

type Payload[T any, R any] func(T) (R, error)

type Options[T any, R any] struct {
	// Payload is func doing the job.
	Payload Payload[T, R]

	// NWorkers is 1 by default.
	NWorkers int

	// ResultQueueSize is result chan buffer size, NWorkers by default.
	ResultQueueSize int
}

// New is a constructor for pool.
func New[T any, R any](opt Options[T, R], tasks []T) *pool[T, R] {
	w := pool[T, R]{
		wg:    new(sync.WaitGroup),
		tasks: make(chan T, len(tasks)),
	}

	if opt.NWorkers == 0 {
		w.n = defaultNWorkers
	} else {
		w.n = opt.NWorkers
	}

	if opt.ResultQueueSize == 0 {
		w.results = make(chan result[R], w.n)
	} else {
		w.results = make(chan result[R], opt.ResultQueueSize)
	}

	if opt.Payload == nil {
		w.logger.Fatal("payload required")
	} else {
		w.f = opt.Payload
	}

	w.addTasksAndClose(tasks...)

	return &w
}

// Run starts pool and returns result queue.
func (w *pool[T, R]) Run() <-chan result[R] {
	for i := 0; i < w.n; i++ {
		go w.worker()
	}

	go func() {
		w.wg.Wait()
		close(w.results)
	}()

	return w.results
}

// addTasksAndClose pushes tasks to queue and closes task queue.
func (w *pool[T, R]) addTasksAndClose(tasks ...T) {
	w.addTasks(tasks...)
	w.noMoreTasks()
}

// addTasks pushes tasks to queue.
func (w *pool[T, R]) addTasks(tasks ...T) {
	w.wg.Add(len(tasks))
	for _, task := range tasks {
		w.tasks <- task
	}
}

// noMoreTasks closes tasks chan.
func (w *pool[T, R]) noMoreTasks() {
	close(w.tasks)
}

// pool is a generic struct for concurrent task execution.
type pool[T any, R any] struct {
	logger  *log.Logger
	wg      *sync.WaitGroup
	tasks   chan T
	n       int
	results chan result[R]
	f       func(T) (R, error)
}

// result is passed through chan.
type result[R any] struct {
	Result R
	Error  error
}

// worker retrieves tasks from tasks queue, executes f, push result in result, until tasks chan is closed.
func (w *pool[T, R]) worker() {
	for t := range w.tasks {
		r, err := w.f(t)
		w.results <- result[R]{Result: r, Error: err}
		w.wg.Done()
	}
}
