package worker_pool_factory

import (
	"context"
	"log"
	"sync"
)

func New[T, R any](w *WorkerPoolFactory[T, R]) *workerPool[T, R] {
	wp := workerPool[T, R]{
		wg: new(sync.WaitGroup),
		f:  w.Payload,
	}

	if w.Logger == nil {
		w.Logger = DefaultLogger(w.LoggerPrefix)
	}
	wp.logger = w.Logger

	if w.TaskQueueSize == 0 {
		wp.tasks = make(chan T)
	} else {
		wp.tasks = make(chan T, w.TaskQueueSize)
	}

	if w.NWorkers == 0 {
		w.NWorkers = 1
	}
	wp.n = w.NWorkers

	if w.ResultQueueSize == 0 {
		wp.results = make(chan result[R], w.NWorkers)
	} else {
		wp.results = make(chan result[R], w.ResultQueueSize)
	}

	var ctx context.Context

	if w.TaskCtx == nil {
		ctx = context.Background()
	} else {
		ctx = w.TaskCtx
	}
	wp.taskCtx, wp.cancelTaskCtx = context.WithCancel(ctx)

	if w.ResultCtx == nil {
		ctx = context.Background()
	} else {
		ctx = w.ResultCtx
	}
	wp.resultCtx, wp.cancelResultCtx = context.WithCancel(ctx)

	wp.logger.Printf("created worker pool: %+v", wp)

	return &wp
}

// workerPool is a generic struct for concurrent task execution.
type workerPool[T any, R any] struct {
	f               PayloadFunc[T, R]
	tasks           chan T
	taskCtx         context.Context
	cancelTaskCtx   context.CancelFunc
	results         chan result[R]
	resultCtx       context.Context
	cancelResultCtx context.CancelFunc
	n               uint
	logger          *log.Logger
	wg              *sync.WaitGroup
}

// TaskChan returns chan with tasks.
func (w *workerPool[T, R]) TaskChan() chan<- T {
	return w.tasks
}

// CloseTaskChan closes tasks chan.
func (w *workerPool[T, R]) CloseTaskChan() {
	w.logger.Println("closing task chan")
	close(w.tasks)
}

func (w *workerPool[T, R]) TaskContext() context.Context {
	return w.taskCtx
}

func (w *workerPool[T, R]) CancelTaskContext() {
	w.logger.Println("cancelling task context")
	w.cancelTaskCtx()
}

// ResultChan returns chan with results.
func (w *workerPool[T, R]) ResultChan() <-chan result[R] {
	return w.results
}

// CloseResultChan closes tasks chan.
func (w *workerPool[T, R]) CloseResultChan() {
	w.logger.Println("closing result chan")
	close(w.results)
}

func (w *workerPool[T, R]) ResultContext() context.Context {
	return w.resultCtx
}

func (w *workerPool[T, R]) CancelResultContext() {
	w.logger.Println("cancelling result context")
	w.cancelResultCtx()
}

// result is passed through chan.
type result[R any] struct {
	Result R
	Error  error
}

// workerFunc retrieves tasks from tasks queue, executes f, push result in result, until tasks chan is closed.
func (w *workerPool[T, R]) workerFunc() {
	for {
		//w.logger.Println("worker wait for event")
		select {
		case task := <-w.tasks:
			//w.logger.Println("running task:", task)
			r, err := w.f(task)
			//w.logger.Println("returning result:", task)
			w.results <- result[R]{Result: r, Error: err}
			//w.logger.Println("task done:", task)
			w.wg.Done()
		case <-w.TaskContext().Done():
			w.logger.Println("task context done")
			return
		}
	}
}

// RunWithContext starts workerPool and returns result queue.
func (w *workerPool[T, R]) RunWithContext() <-chan result[R] {
	for i := uint(0); i < w.n; i++ {
		w.logger.Println("starting worker", i)
		go w.workerFunc()
	}
	w.closeResultsAfterDone()
	return w.ResultChan()
}

func (w *workerPool[T, R]) RunTasks(tasks []T) <-chan result[R] {
	res := w.RunWithContext()
	w.AddTasks(tasks)
	w.WaitForTasks()
	return res
}

func (w *workerPool[T, R]) AddTask(task T) {
	w.AddTasks([]T{task})
}

func (w *workerPool[T, R]) AddTasks(tasks []T) {
	w.wg.Add(len(tasks))
	go func() {
		for _, task := range tasks {
			//w.logger.Println("adding task:", task)
			w.TaskChan() <- task
			//w.logger.Println("task added:", task)
		}
	}()
}

func (w *workerPool[T, R]) WaitForTasks() {
	go func() {
		w.wg.Wait()
		w.CancelResultContext()
		w.CancelTaskContext()
	}()
}

func (w *workerPool[T, R]) closeResultsAfterDone() {
	go func() {
		w.logger.Println("waiting for result context cancel")
		<-w.ResultContext().Done()
		w.logger.Println("result context cancelled")
		w.CloseResultChan()
	}()
}
