package factors

import (
	"math"
	"slices"
)

func Factors(n int) []int {
	allFactors := make([]int, 0)
	floatN := float64(n)
	var i int = 1
	for ; i < int(math.Floor(math.Sqrt(floatN))) ; i++ {
		if((n % i) == 0){
			allFactors = append(allFactors, i)
			if(n / i != i){
				allFactors = append(allFactors, n / i)
			}
		}
	}
	slices.SortFunc(allFactors, func(a,b int) int {
		return a-b
	})
	return allFactors
}