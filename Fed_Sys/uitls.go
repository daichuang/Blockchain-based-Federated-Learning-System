package main

import (
	"errors"
	"math"

	"gonum.org/v1/gonum/mat"
)

//utils function
func GenerateGoNumData(size int, value float64) *mat.Dense {
	float_List := make([]float64, size)

	for i := 0; i < size; i++ {
		float_List[i] = value
	}

	ret := mat.NewDense(1, size, float_List)
	return ret
}

func matCopy(a *mat.Dense, size int) *mat.Dense {

	temp := make([]float64, size)
	mat.Row(temp, 0, a)
	ret := mat.NewDense(1, size, temp)

	return ret

}

//cosine similarity
func Cosine(a []float64, b []float64) (cosine float64, err error) {
	count := 0
	length_a := len(a)
	length_b := len(b)
	if length_a > length_b {
		count = length_a
	} else {
		count = length_b
	}
	sumA := 0.0
	s1 := 0.0
	s2 := 0.0
	for k := 0; k < count; k++ {
		if k >= length_a {
			s2 += math.Pow(b[k], 2)
			continue
		}
		if k >= length_b {
			s1 += math.Pow(a[k], 2)
			continue
		}
		sumA += a[k] * b[k]
		s1 += math.Pow(a[k], 2)
		s2 += math.Pow(b[k], 2)
	}
	if s1 == 0 || s2 == 0 {
		return 0.0, errors.New("Vectors should not be null (all zeros)")
	}
	return sumA / (math.Sqrt(s1) * math.Sqrt(s2)), nil
}

func StructquickSort(blockUpdate []NewDelta, start, end int) {
	if start < end {
		i, j := start, end
		key := blockUpdate[(start+end)/2].SourceID
		for i <= j {
			for blockUpdate[i].SourceID < key {
				i++
			}
			for blockUpdate[j].SourceID > key {
				j--
			}
			if i <= j {
				blockUpdate[i], blockUpdate[j] = blockUpdate[j], blockUpdate[i]
				i++
				j--
			}
		}

		if start < j {
			StructquickSort(blockUpdate, start, j)
		}
		if end > i {
			StructquickSort(blockUpdate, i, end)
		}
	}
}

func matConverFloat64List(befort *mat.Dense, size int) []float64 {
	res := make([]float64, size)
	mat.Row(res, 0, befort)
	return res
}

func BlockDeltaDivDataNum(Num *mat.Dense, Delta []float64) []float64 {
	Temp := mat.NewDense(1, len(Delta), Delta)
	Temp_copy := matCopy(Temp, len(Delta)) //防止原Delta发生改变
	Temp_copy.DivElem(Temp_copy, Num)
	res := matConverFloat64List(Temp_copy, len(Delta))

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
