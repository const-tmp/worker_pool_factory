package worker_pool_factory

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var nWorkers = []uint{0, 1, 2}
var taskQueueSize = []uint{0, 1, 2}
var resultQueueSize = []uint{0, 1, 2}

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

func genCases(n int) ([]int, []Result[string]) {
	var (
		t []int
		r []Result[string]
	)
	for i := 0; i < pow(10, n); i++ {
		t = append(t, i)
		r = append(r, Result[string]{Result: fmt.Sprint(i), Error: nil})
	}
	return t, r
}

func TestRunTasks(t *testing.T) {
	for _, nWorker := range nWorkers {
		for _, tqs := range taskQueueSize {
			for _, rqs := range resultQueueSize {
				for i := 0; i < n; i++ {
					name := fmt.Sprintf("run Tasks %d %d %d %d ", nWorker, tqs, rqs, i)
					t.Run(name, func(t *testing.T) {
						tasks, want := genCases(i)

						var tc chan int
						if tqs == 0 {
							tc = make(chan int)
						} else {
							tc = make(chan int, tqs)
						}

						var rc chan Result[string]
						if rqs == 0 {
							rc = make(chan Result[string])
						} else {
							rc = make(chan Result[string], rqs)
						}

						wp := WorkerPoolFactory[int, string]{
							Payload: func(i int) (string, error) {
								return fmt.Sprint(i), nil
							},
							NWorkers:     nWorker,
							Tasks:        tc,
							Results:      rc,
							LoggerPrefix: name,
						}
						wp.Run()
						go func() {
							for _, task := range tasks {
								tc <- task
							}
							close(tc)
						}()

						got := make([]Result[string], 0, len(tasks))
						for r := range rc {
							t.Log("got", r)
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
