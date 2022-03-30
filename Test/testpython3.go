package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/DataDog/go-python3"
)

var (
	pyLogModule *python3.PyObject
	pyTest      *python3.PyObject
)

func main() {
	py3Init()
}

func py3Init() {
	python3.Py_Initialize()
	if !python3.Py_IsInitialized() {
		fmt.Println("Error initializing the python interpreter")
		os.Exit(1)
	}

	m := python3.PyImport_ImportModule("sys")
	sysPath := m.GetAttrString("path")

	python3.PyList_Insert(sysPath, 0, python3.PyUnicode_FromString("/home/zhaoxinbo/BFL/ML/Pytorch"))

	pyLogModule = python3.PyImport_ImportModule("client_obj")
	pyTest = pyLogModule.GetAttrString("Test")
	if pyTest == nil {
		panic("Error importing function")
	}

	//args set

	argArray := python3.PyList_New(3)
	for i := 0; i < 3; i++ {
		python3.PyList_SetItem(argArray, i, python3.PyFloat_FromDouble(0.666))
	}

	args := python3.PyTuple_New(1)
	// param1 := python3.PyUnicode_FromString("creditcard")
	// param2 := python3.PyFloat_FromDouble(0.666)

	python3.PyTuple_SetItem(args, 0, argArray)
	_gstate := python3.PyGILState_Ensure()

	//python3.PyTuple_SetItem(args, 1, paramAge)
	//paramAge := python3.PyLong_FromGoInt(99)
	var result *python3.PyObject
	result = pyTest.CallObject(args)
	if !(result != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
	}

	// Convert the resulting array to a go byte array

	goBytes := python3.PyByteArray_FromObject(result)
	GoChar := python3.PyByteArray_AsString(goBytes)
	goByteArray := []byte(GoChar)

	// outLog.Printf("GoByte is:%s", goByteArray)
	python3.PyGILState_Release(_gstate)

	var goFloatArray []float64
	size := len(GoChar) / 8

	for i := 0; i < size; i++ {
		currIndex := i * 8
		bits := binary.LittleEndian.Uint64(goByteArray[currIndex : currIndex+8])
		aFloat := math.Float64frombits(bits)
		goFloatArray = append(goFloatArray, aFloat)
	}

	fmt.Println("the result from python is ", goFloatArray)
	python3.Py_Finalize()

}
