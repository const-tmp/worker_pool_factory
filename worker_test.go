package ezwp

import (
	"fmt"
	"testing"
)

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

func Test_pool_Run(t *testing.T) {
	const n = 6
	for i := 0; i < n; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tasks, want := genCases(i)
			pool := New[int, string](Options[int, string]{
				Payload: func(i int) (string, error) {
					return fmt.Sprint(i), nil
				}}, tasks)

			got := make([]result[string], 0, len(tasks))
			for r := range pool.Run() {
				got = append(got, r)
			}
			if len(got) != len(want) {
				t.Errorf("Run() = %v, want %v", got, want)
			}
		})
	}
}
