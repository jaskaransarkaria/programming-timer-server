package main

import (
	"fmt"
	"errors"
	"math"
)

func main() {
	result, err := sqrt(16)
	if err != nil {
		fmt.Println(nil)
	} else {
		fmt.Println(result)
	}
}

func sqrt(x float64) (float64, error) {
	if x < 0 {
		return  0, errors.New("Undefined for negative numbers")
	}

	return math.Sqrt(x), nil
}
