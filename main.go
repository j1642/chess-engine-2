package main

import (
	"engine2/pieces"
	"fmt"
	"time"
)

func main() {
	// 180ms for array of arrays linear search, 205ms for array of maps,
	// 190ms for array of arrays binary search.
	start := time.Now()
	for i := 0; i < 10000; i++ {
		pieces.MakeKnightBBs()
	}
	fmt.Println(time.Now().Sub(start))
}
