// package main

// import (
// 	"fmt"
// 	"strconv"

// 	"github.com/sbinet/go-python"
// 	//"go.dedis.ch/kyber/v3/pairing/bn256"
// 	//"os"
// 	//"bufio"
// 	//"encoding/json"
// )

// var (
// 	pyLogModule *python.PyObject
// 	pyTest      *python.PyObject
// )

// var pyTestModule *python.PyObject
// var pyTrainFunc *python.PyObject

// func init() {
// 	err := python.Initialize()
// 	if err != nil {
// 		panic(err.Error())
// 	}
// }
// func main() {
// 	Init_Pytorch("mnist")
// 	// TestPath("creditcard")//

// 	// var w = []float64{-0.0091, -0.0051, 0.0028, 0.0001, 0.0021, 0.0006, -0.0008, 0.0045, 0.0039, 0.0033, 0.0061, 0.0016, -0.0053, -0.0054, -0.0031, -0.0035, -0.0052, -0.0046, -0.0031, -0.0053, -0.0004, -0.0013, -0.0003, -0.0003, 0.0068}
// 	// fmt.Println(testTrainErr(w))

// }

// func testTrainErr(weights []float64) (res float64) {
// 	m := python.PyImport_ImportModule("sys")
// 	sysPath := m.GetAttrString("path")
// 	python.PyList_Insert(sysPath, 0, python.PyString_FromString("/home/zhaoxinbo/BFL/ML/code"))
// 	fmt.Println(sysPath)

// 	pyTestModule = python.PyImport_ImportModule("logistic_model_test")
// 	if pyTestModule == nil {
// 		exc, val, tb := python.PyErr_Fetch()
// 		fmt.Printf("exc=%v\nval=%v\ntb=%v\n", exc, val, tb)
// 		panic(fmt.Errorf("could not call function"))
// 	}

// 	pyTrainFunc = pyTestModule.GetAttrString("train_error")
// 	if pyTrainFunc == nil {
// 		exc, val, tb := python.PyErr_Fetch()
// 		fmt.Printf("exc=%v\nval=%v\ntb=%v\n", exc, val, tb)
// 		panic(fmt.Errorf("could not call function"))
// 	}

// 	argArray := python.PyList_New(len(weights))

// 	for i := 0; i < len(weights); i++ {
// 		python.PyList_SetItem(argArray, i, python.PyFloat_FromDouble(weights[i]))
// 	}
// 	fmt.Println(argArray)

// 	pyTrainResult := pyTrainFunc.CallFunction(argArray)
// 	res = python.PyFloat_AsDouble(pyTrainResult)
// 	return

// }

// func TestPath(datasetName string) {
// 	m := python.PyImport_ImportModule("sys")
// 	sysPath := m.GetAttrString("path")

// 	python.PyList_Insert(sysPath, 0, python.PyString_FromString("/home/zhaoxinbo/BFL/ML/code"))
// 	fmt.Println(sysPath)

// 	pyLogModule = python.PyImport_ImportModule("logistic_model")
// 	if pyLogModule == nil {
// 		exc, val, tb := python.PyErr_Fetch()
// 		fmt.Printf("exc=%v\nval=%v\ntb=%v\n", exc, val, tb)
// 		panic(fmt.Errorf("could not call function1"))
// 	}
// 	pyTest = pyLogModule.GetAttrString("test")
// 	if pyTest == nil {
// 		exc, val, tb := python.PyErr_Fetch()
// 		fmt.Printf("exc=%v\nval=%v\ntb=%v\n", exc, val, tb)
// 		panic(fmt.Errorf("could not call function2"))
// 	}

// 	pyNumFeatures := pyTest.CallFunction(python.PyString_FromString(datasetName))
// 	if pyNumFeatures == nil {
// 		exc, val, tb := python.PyErr_Fetch()
// 		fmt.Printf("exc=%v\nval=%v\ntb=%v\n", exc, val, tb)
// 		panic(fmt.Errorf("could not call function3"))
// 	}
// 	numFeatures := python.PyInt_AsLong(pyNumFeatures)
// 	fmt.Println(numFeatures)

// }

// func Init_Pytorch(datasetName string) {

// 	m := python.PyImport_ImportModule("sys")
// 	sysPath := m.GetAttrString("path")

// 	python.PyList_Insert(sysPath, 0, python.PyString_FromString("/home/zhaoxinbo/BFL/ML/Pytorch"))
// 	fmt.Println(sysPath)

// 	pyLogModule = python.PyImport_ImportModule("client_obj")
// 	pyTest = pyLogModule.GetAttrString("init")
// 	if pyTest == nil {
// 		fmt.Println("Error importing function")
// 	}

// 	pyNumFeatures := pyTest.CallFunction(python.PyString_FromString(datasetName), python.PyString_FromString(datasetName+strconv.Itoa((0))), python.PyInt_FromLong(0), python.PyInt_FromLong(25))
// 	if pyNumFeatures == nil {
// 		exc, val, tb := python.PyErr_Fetch()
// 		fmt.Printf("exc=%v\nval=%v\ntb=%v\n", exc, val, tb)
// 		panic(fmt.Errorf("could not call function"))
// 	}
// 	numFeatures := python.PyInt_AsLong(pyNumFeatures)
// 	fmt.Println(numFeatures)

// }

// //func pyInit(datasetName string, dataFile string, epsilon float64)  {
// //
// //	m := python.PyImport_ImportModule("sys")
// //	sysPath := m.GetAttrString("path")
// //
// //
// //
// //	python.PyList_Insert(sysPath, 0, python.PyString_FromString("/home/zhaoxinbo/Biscotti/ML/code"))
// //	fmt.Println(sysPath)
// //
// //	pyLogModule = python.PyImport_ImportModule("logistic_model")
// //	pyTest = pyLogModule.GetAttrString("init")
// //	if pyTest == nil {
// //		fmt.Println("Error importing function")
// //	}
// //
// //	pyNumFeatures := pyTest.CallFunction(python.PyString_FromString(datasetName),python.PyString_FromString(datasetName),python.PyInt_FromLong(0),python.PyInt_FromLong(25))
// //        if pyNumFeatures == nil {
// //		exc, val, tb := python.PyErr_Fetch()
// //		fmt.Printf("exc=%v\nval=%v\ntb=%v\n", exc, val, tb)
// //		panic(fmt.Errorf("could not call function"))
// //	}
// //	numFeatures := python.PyInt_AsLong(pyNumFeatures)
// //	fmt.Println(numFeatures)
// //
// //
// //}
