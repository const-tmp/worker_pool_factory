package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/h1ght1me/worker_pool_factory"
)

func randomTicker(minMs, maxMs int) <-chan struct{} {
	rand.Seed(time.Now().UnixNano())
	c := make(chan struct{})
	go func() {
		for {
			time.Sleep(time.Duration(rand.Intn(maxMs-minMs)+minMs) * time.Millisecond)
			c <- struct{}{}
		}
	}()
	return c
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wp := worker_pool_factory.WorkerPoolFactory[int, string]{
		Payload: func(i int) (string, error) {
			fmt.Println("payload running", i)
			return fmt.Sprint(i), nil
		},
		NWorkers:        1,
		TaskQueueSize:   0,
		TaskCtx:         ctx,
		ResultQueueSize: 0,
		ResultCtx:       ctx,
		Logger:          nil,
		LoggerPrefix:    "task list: ",
	}

	pool := wp.Create()

	results := pool.RunWithContext()

	go func() {
		ticker := randomTicker(100, 1000)
		var i int
		for {
			select {
			case <-ctx.Done():
				fmt.Println("producer ctx done")
				return
			case <-ticker:
				i++
				fmt.Println("putting task:", i)
				pool.AddTask(i)
			}
		}
	}()

	for r := range results {
		fmt.Println(r)
	}
}
