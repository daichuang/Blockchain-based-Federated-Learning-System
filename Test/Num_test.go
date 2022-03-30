package main

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

func main() {

	a := mat.NewDense(1, 3, []float64{3.0, 3.0, 3.0})
	b := mat.NewDense(1, 3, []float64{3.0, 3.0, 3.0})

	fmt.Println(a)
	fmt.Println(b)
	a.Add(a, b)
	fmt.Println(a)

}
