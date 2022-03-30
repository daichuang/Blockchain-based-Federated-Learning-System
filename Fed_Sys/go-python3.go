package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"

	"github.com/DataDog/go-python3"
)

var (
	pyModule       *python3.PyObject
	pyCreateClient *python3.PyObject
	pyNum          *python3.PyObject
	pyGetInitParam *python3.PyObject

	pyParam      *python3.PyObject
	pyGetDataNum *python3.PyObject
	pyTestErr    *python3.PyObject
	pyTestAcc    *python3.PyObject
	pyNoise      *python3.PyObject
	pyOutlier    *python3.PyObject

	PyresetVar      *python3.PyObject
	PygradCollector *python3.PyObject
	PygradTrim      *python3.PyObject

	blockExistsError = errors.New("区块已经存在")
)

const (
	epoch      = 1
	batch_size = 50
	ModulePath = home + "/BFL/python3File/DeepLearning"
)

//初始化go-python接口
func pyInit() {

	python3.Py_Initialize()
	if !python3.Py_IsInitialized() {
		fmt.Println("Error initializing the python interpreter")
		os.Exit(1)
	}
	m := python3.PyImport_ImportModule("sys")
	if m == nil {
		fmt.Println("import error")
	}
	sysPath := m.GetAttrString("path")
	if sysPath == nil {
		fmt.Println("get path error")
	}

	python3.PyList_Insert(sysPath, 0, python3.PyUnicode_FromString(ModulePath))
	//Py3 Module
	pyModule = python3.PyImport_ImportModule("client_obj")
	//Py3 Func
	pyNum = pyModule.GetAttrString("getNumFeature")
	pyParam = pyModule.GetAttrString("GetLocalParam")
	pyGetDataNum = pyModule.GetAttrString("GetLocalDataNum")
	pyNoise = pyModule.GetAttrString("getNoise")
	pyTestErr = pyModule.GetAttrString("getTestErr")
	pyTestAcc = pyModule.GetAttrString("getTestAcc")
	pyGetInitParam = pyModule.GetAttrString("getInitWeight")
	pyCreateClient = pyModule.GetAttrString("CreateClient")
	pyOutlier = pyModule.GetAttrString("smoothedGrag_detect")

	//trimed_func
	PyresetVar = pyModule.GetAttrString("resetVar")
	PygradCollector = pyModule.GetAttrString("gradCollector")
	PygradTrim = pyModule.GetAttrString("gradTrim")
}

func CreateClient(datasetName string, dataFile string, disable_dp bool, epsilon float64) {

	args := python3.PyTuple_New(6)
	param1 := python3.PyUnicode_FromString(datasetName)
	param2 := python3.PyUnicode_FromString(dataFile)
	param3 := python3.PyLong_FromLong(batch_size)
	param4 := python3.PyLong_FromLong(epoch)
	var param5 *python3.PyObject
	if disable_dp {
		param5 = python3.PyBool_FromLong(1)
	} else {
		param5 = python3.PyBool_FromLong(0)
	}
	param6 := python3.PyLong_FromDouble(epsilon)

	python3.PyTuple_SetItem(args, 0, param1)
	python3.PyTuple_SetItem(args, 1, param2)
	python3.PyTuple_SetItem(args, 2, param3)
	python3.PyTuple_SetItem(args, 3, param4)
	python3.PyTuple_SetItem(args, 4, param5)
	python3.PyTuple_SetItem(args, 5, param6)

	pyCreateClient.CallObject(args)
}

//初始化go-python接口
func GetNumFeature(datasetName string) int {

	args := python3.PyTuple_New(1)
	param1 := python3.PyUnicode_FromString(datasetName)
	python3.PyTuple_SetItem(args, 0, param1)

	var pyNumFeatures *python3.PyObject

	pyNumFeatures = pyNum.CallObject(args)

	if pyNumFeatures == nil {
		exc, val, tb := python3.PyErr_Fetch()
		fmt.Printf("exc=%v\nval=%v\ntb=%v\n", exc, val, tb)
		panic(fmt.Errorf("could not call function"))
	}
	numFeatures := python3.PyLong_AsLong(pyNumFeatures)

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+": 成功拉取到模型. 模型参数量为: %d\n", numFeatures)

	return numFeatures
}

func getInitWeight() []float64 {

	//runtime.LockOSThread会锁定当前协程只跑在一个系统线程上，这个线程里也只跑该协程。
	runtime.LockOSThread()

	//Python3 解释器不是完全线程安全的。当前线程想要安全访问Python对象的前提是获取用以支持多线程安全的全局锁。
	//没有锁，甚至多线程程序中最简单的操作都会发生问题。
	_gstate := python3.PyGILState_Ensure()

	var InitWeight *python3.PyObject
	InitWeight = pyGetInitParam.CallFunctionObjArgs()

	// 将结果转换为go字节数组
	PyByteArray := python3.PyByteArray_FromObject(InitWeight)
	GoChar := python3.PyByteArray_AsString(PyByteArray)
	goByteArray := []byte(GoChar)

	python3.PyGILState_Release(_gstate)

	var goFloatArray []float64
	size := len(goByteArray) / 8

	for i := 0; i < size; i++ {
		currIndex := i * 8
		bits := binary.LittleEndian.Uint64(goByteArray[currIndex : currIndex+8])
		aFloat := math.Float64frombits(bits)
		goFloatArray = append(goFloatArray, aFloat)
	}
	return goFloatArray
}

// func getTrimmedGrad(num_update_accpted int, byzantineNum int, numFeatures int) []float64 {
// 	runtime.LockOSThread()
// 	_gstate := python3.PyGILState_Ensure()
// 	args := python3.PyTuple_New(3)
// 	param1 := python3.PyLong_FromLong(num_update_accpted)
// 	param2 := python3.PyLong_FromLong(byzantineNum)
// 	param3 := python3.PyLong_FromLong(numFeatures)
// 	python3.PyTuple_SetItem(args, 0, param1)
// 	python3.PyTuple_SetItem(args, 1, param2)
// 	python3.PyTuple_SetItem(args, 2, param3)
// 	var gradAfterTrim *python3.PyObject
// 	gradAfterTrim = PygradTrim.CallObject(args)

// 	PyByteArray := python3.PyByteArray_FromObject(gradAfterTrim)
// 	GoChar := python3.PyByteArray_AsString(PyByteArray)
// 	goByteArray := []byte(GoChar)

// 	python3.PyGILState_Release(_gstate)

// 	var goFloatArray []float64
// 	size := len(goByteArray) / 8
// 	for i := 0; i < size; i++ {
// 		currIndex := i * 8
// 		bits := binary.LittleEndian.Uint64(goByteArray[currIndex : currIndex+8])
// 		aFloat := math.Float64frombits(bits)
// 		goFloatArray = append(goFloatArray, aFloat)
// 	}
// 	return goFloatArray
// }
