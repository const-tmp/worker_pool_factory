package worker_pool_factory

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var nWorkers = []uint{0, 1, 5}
var taskQueueSize = []uint{0, 1, 5}
var resultQueueSize = []uint{0, 1, 5}

const n = 4

func pow(x, y int) int {
	r := 1
	for i := 0; i < y; i++ {
		r *= x
	}
	return r
}

func TestPow(t *testing.T) {
	tests := []struct{ x, y, want int }{
		{0, 0, 1},
		{1, 0, 1},
		{0, 1, 0},
		{1, 1, 1},
		{2, 1, 2},
		{2, 2, 4},
		{10, 1, 10},
		{10, 2, 100},
		{10, 3, 1000},
	}
	for _, tc := range tests {
		res := pow(tc.x, tc.y)
		if res != tc.want {
			t.Errorf("x=%d y=%d want=%d got=%d", tc.x, tc.y, tc.want, res)
		}
	}
}

func genCases(n int) ([]int, []result[string]) {
	var (
		t []int
		r []result[string]
	)
	for i := 0; i < pow(10, n); i++ {
		t = append(t, i)
		r = append(r, result[string]{Result: fmt.Sprint(i), Error: nil})
	}
	return t, r
}

func TestRunTasks(t *testing.T) {
	for _, nWorker := range nWorkers {
		for _, tqs := range taskQueueSize {
			for _, rqs := range resultQueueSize {
				for i := 0; i < n; i++ {
					name := fmt.Sprintf("run tasks %d %d %d %d ", nWorker, tqs, rqs, i)
					t.Run(name, func(t *testing.T) {
						tasks, want := genCases(i)

						wp := WorkerPoolFactory[int, string]{
							Payload: func(i int) (string, error) {
								return fmt.Sprint(i), nil
							},
							NWorkers:        nWorker,
							TaskQueueSize:   tqs,
							ResultQueueSize: rqs,
							LoggerPrefix:    name,
						}
						pool := wp.Create()

						got := make([]result[string], 0, len(tasks))
						for r := range pool.RunTasks(tasks) {
							got = append(got, r)
						}
						if len(got) != len(want) {
							t.Errorf("RunWithContext() = %v, want %v", got, want)
						}
					})
				}
			}
		}
	}
}

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

func TestRunWithContext(t *testing.T) {
	for _, nWorker := range nWorkers {
		for _, tqs := range taskQueueSize {
			for _, rqs := range resultQueueSize {
				name := fmt.Sprintf("run with context %d %d %d ", nWorker, tqs, rqs)
				t.Run(name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()

					f := WorkerPoolFactory[int, string]{
						Payload: func(i int) (string, error) {
							return fmt.Sprint(i), nil
						},
						NWorkers:        nWorker,
						TaskQueueSize:   tqs,
						TaskCtx:         ctx,
						ResultQueueSize: rqs,
						ResultCtx:       ctx,
						LoggerPrefix:    name,
					}
					wp := f.Create()
					results := wp.RunWithContext()
					go func() {
						ticker := randomTicker(50, 300)
						var i int
						for {
							select {
							case <-ctx.Done():
								t.Log("producer ctx done")
								return
							case <-ticker:
								i++
								t.Log("putting task:", i)
								wp.AddTask(i)
							}
						}
					}()
					for r := range results {
						t.Log("result:", r.Result)
					}
				})
			}
		}
	}
}

func TestChanClosing(t *testing.T) {
	num := 10
	ci := make(chan int, num)
	for i := 0; i < num; i++ {
		ci <- i
	}
	close(ci)
	for {
		i, ok := <-ci
		if !ok {
			t.Log("chan closed")
			break
		}
		t.Log(i, ok)
	}
}
