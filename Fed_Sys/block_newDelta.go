package main

import (
	"fmt"
	"strconv"
	"strings"
)

type NewDelta struct {
	SourceID    int
	Iteration   int
	NoisedDelta []float64 //param with Diffential Privacy noise
	NumData     float64
	Accepted    bool
}

func (delta *NewDelta) String() string {

	return fmt.Sprintf("{Iteration:" + strconv.Itoa(delta.Iteration) + ", " + "Noise_Params:" + arrayToString(delta.NoisedDelta, ",") + "}")

}

func arrayToString(a []float64, delim string) string {
	str := "[" + strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]") + "]"
	return str

}
