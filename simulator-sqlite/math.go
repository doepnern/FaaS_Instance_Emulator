package main

import (
	"math"
	"time"

	rng "github.com/leesper/go_rng"
)

func LognormalGenerator(loc, scale, s float64) float64 {
	lrng := rng.NewLognormalGenerator(time.Now().UnixNano())
	res := lrng.Lognormal(math.Log(scale), s) + loc
	return res
}
