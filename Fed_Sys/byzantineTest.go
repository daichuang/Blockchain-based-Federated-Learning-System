package main

import (
	"runtime"
	"strconv"

	"github.com/DataDog/go-python3"
)

func byzantineTest(BlockDelta []NewDelta, Smoothed_grad []float64) []int {

	CosineSimilarutyList := []float64{}
	StructquickSort(BlockDelta, 0, len(BlockDelta)-1)
	// for _, Delta := range BlockDelta {
	// 	outLog.Println(Delta.SourceID)
	// }
	//得到按节点排序之后，每个节点上传梯度与Smoothed_grad之间的余弦相似度
	for _, Delta := range BlockDelta {
		CosineSimilarity, err := Cosine(BlockDeltaDivDataNum(GenerateGoNumData(len(Delta.NoisedDelta), Delta.NumData), Delta.NoisedDelta), Smoothed_grad)
		// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":CosineSimilarity:", CosineSimilarity)
		if err == nil {
			CosineSimilarutyList = append(CosineSimilarutyList, CosineSimilarity)

		}
	}
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":将节点顺序排列并计算后得到的余弦相似度列表为:", CosineSimilarutyList)

	// grubbs 异常检测

	runtime.LockOSThread()
	_gstate := python3.PyGILState_Ensure()

	CosineSimilarutyArray := python3.PyList_New(len(CosineSimilarutyList))

	for i := 0; i < len(CosineSimilarutyList); i++ {
		python3.PyList_SetItem(CosineSimilarutyArray, i, python3.PyFloat_FromDouble(CosineSimilarutyList[i]))
	}

	outLog.Println("对等节点 "+strconv.Itoa(client.id)+":将节点顺序排列并计算后得到的余弦相似度列表为:", CosineSimilarutyList)

	var List_index *python3.PyObject

	args := python3.PyTuple_New(2)
	param := python3.PyLong_FromLong(byzantineNum)
	python3.PyTuple_SetItem(args, 0, CosineSimilarutyArray)
	python3.PyTuple_SetItem(args, 1, param)

	List_index = pyOutlier.CallObject(args)

	// 将结果转换为go字节数组
	PyByteArray := python3.PyByteArray_FromObject(List_index)
	GoChar := python3.PyByteArray_AsString(PyByteArray)
	goByteArray := []byte(GoChar)

	python3.PyGILState_Release(_gstate)

	var goIntArray []int
	for _, value := range goByteArray {
		goIntArray = append(goIntArray, int(value))
	}

	return goIntArray

}
