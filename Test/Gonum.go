package main

import (
	"math/rand"

	"gonum.org/v1/gonum/mat"
)

func main() {

	rand.Seed(7)
	a := rand.Int63n(10000000)
	print(a)
	// a := mat.NewDense(1, 3, []float64{6.0, 6.0, 6.0})
	// c := []float64{3, 3, 3}
	// fmt.Println(a)
	// fmt.Println(c)

	// res := BlockDeltaDivDataNum2(a, c)
	// fmt.Println("----")
	// fmt.Println(a)
	// fmt.Println(c)
	// fmt.Println(res)

	// var sig bool
	// for index, _ := range []int{0, 3, 0, 3} {
	// 	fmt.Println(index)

	// 	sig = dataIn(index, []int{2})
	// 	fmt.Println(sig)
	// }

}

func matCopy(a *mat.Dense, size int) *mat.Dense {

	temp := make([]float64, size)
	mat.Row(temp, 0, a)
	ret := mat.NewDense(1, size, temp)

	return ret

}

func BlockDeltaDivDataNum(Num *mat.Dense, Delta []float64) []float64 {
	Temp := mat.NewDense(1, len(Delta), Delta)
	Temp_copy := matCopy(Temp, len(Delta)) //防止原Delta发生改变
	Temp_copy.DivElem(Temp_copy, Num)
	res := matConverFloat64List(Temp_copy, len(Delta))

	return res
}

func BlockDeltaDivDataNum2(Num *mat.Dense, Delta []float64) []float64 {
	Temp := mat.NewDense(1, len(Delta), Delta)
	Temp.DivElem(Temp, Num)
	res := matConverFloat64List(Temp, len(Delta))

	return res
}

func matConverFloat64List(befort *mat.Dense, size int) []float64 {
	res := make([]float64, size)
	mat.Row(res, 0, befort)
	return res
}

func dataIn(ind int, list []int) bool {
	var sig bool
	for _, value := range list {
		if ind == value {
			sig = true
			break
		}
	}
	return sig
}
