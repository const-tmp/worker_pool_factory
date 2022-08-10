package main

import (
	"fmt"
	"github.com/h1ght1me/ezwp"
)

func main() {
	wp := ezwp.New[int, string](
		ezwp.Options[int, string]{
			Payload: func(i int) (string, error) {
				return fmt.Sprint(i), nil
			},
		},
		[]int{1, 2, 3, 4, 5},
	)

	for r := range wp.Run() {
		fmt.Println(r)
	}
}
