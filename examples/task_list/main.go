package main

import (
	"fmt"
	"github.com/h1ght1me/worker_pool_factory"
)

func main() {
	wp := worker_pool_factory.WorkerPoolFactory[int, string]{
		Payload: func(i int) (string, error) {
			fmt.Println("payload running", i)
			return fmt.Sprint(i), nil
		},
		NWorkers:        1,
		TaskQueueSize:   0,
		TaskCtx:         nil,
		ResultQueueSize: 0,
		ResultCtx:       nil,
		Logger:          nil,
		LoggerPrefix:    "task list: ",
	}

	pool := wp.Create()

	var tasks []int
	for i := 0; i < 5; i++ {
		tasks = append(tasks, i)
	}

	results := pool.RunTasks(tasks)

	for r := range results {
		fmt.Println("result:", r)
	}
}
