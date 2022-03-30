package main

import (
	"encoding/binary"
	"math"
	"runtime"
	"strconv"

	"github.com/DataDog/go-python3"
	"gonum.org/v1/gonum/mat"
)

const home string = "/home/zhaoxinbo"

const (
	STAKE_UNIT    = 5
	convThreshold = 0.00 //收敛变量
)

type Communication_object struct {
	id           int
	dataset      string
	numFeatures  int
	Delta        NewDelta
	blockUpdates []NewDelta
	bc           *Blockchain
}

//使用chain上最新的全局模型计算下一次更新,使用python接口
func oneEpochStep(globalW []float64) ([]float64, int, error) {

	//runtime.LockOSThread会锁定当前协程只跑在一个系统线程上，这个线程里也只跑该协程。
	runtime.LockOSThread()

	//Python 解释器不是完全线程安全的。当前线程想要安全访问Python对象的前提是获取用以支持多线程安全的全局锁。
	//没有锁，甚至多线程程序中最简单的操作都会发生问题。
	_gstate := python3.PyGILState_Ensure()

	argArray := python3.PyList_New(len(globalW))

	for i := 0; i < len(globalW); i++ {
		python3.PyList_SetItem(argArray, i, python3.PyFloat_FromDouble(globalW[i]))
	}

	var newGrad *python3.PyObject
	var pyNumData *python3.PyObject

	//args set
	args := python3.PyTuple_New(1)
	python3.PyTuple_SetItem(args, 0, argArray)

	//outLog.Println("argary", argArray)
	newGrad = pyParam.CallObject(args)
	pyNumData = pyGetDataNum.CallFunctionObjArgs()

	//get result

	goDataNum := python3.PyLong_AsLong(pyNumData)
	// 将结果转换为go字节数组
	PyByteArray := python3.PyByteArray_FromObject(newGrad)
	GoChar := python3.PyByteArray_AsString(PyByteArray)
	goByteArray := []byte(GoChar)

	// outLog.Printf("GoByte is:%s", goByteArray)
	python3.PyGILState_Release(_gstate)

	var goFloatArray []float64
	size := len(goByteArray) / 8

	for i := 0; i < size; i++ {
		currIndex := i * 8
		bits := binary.LittleEndian.Uint64(goByteArray[currIndex : currIndex+8])
		aFloat := math.Float64frombits(bits)
		goFloatArray = append(goFloatArray, aFloat)
	}

	outLog.Println("对等节点 "+strconv.Itoa(client.id)+":当前迭代轮次"+strconv.Itoa(iterationCount+1)+"的梯度计算结果(本地乘数据量后)为(长度）", len(goFloatArray), "数据量为", goDataNum)

	return goFloatArray, goDataNum, nil

}

//初始化链
func (Communication_object *Communication_object) initBlockChain(InitWeight []float64, datasetName string) {

	Communication_object.numFeatures = GetNumFeature(datasetName)
	Communication_object.dataset = datasetName
	Communication_object.bc = NewBlockchain(InitWeight, Communication_object.numFeatures) //创建创世区块区块 numfeature

}

//通过调用python的oneGradientStep函数来计算更新，并将最新的全局模型从链传递给它
func (Communication_object *Communication_object) computeUpdate(iterationCount int) {
	prevModel := Communication_object.bc.getLatestGradient()
	//outLog.Printf("Global Model:%s", prevModel)
	deltas, DataNum, err := oneEpochStep(prevModel)
	outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":梯度计算完成 ，为自己的梯度添加差分隐私\n")

	check(err)
	Communication_object.Delta = NewDelta{
		SourceID:    Communication_object.id,
		Iteration:   iterationCount,
		NoisedDelta: deltas,
		NumData:     float64(DataNum),
		Accepted:    true}

	//outLog.Printf("Deltas:%s", Communication_object.update.Delta)
}

//添加一个更新到当前迭代接收到的更新记录
func (Communication_object *Communication_object) blockAddUpdate(update NewDelta) int {

	Communication_object.blockUpdates = append(Communication_object.blockUpdates, update)
	return len(Communication_object.blockUpdates)
}

//将更新创建一个区块
//Kernel
func (Communication_object *Communication_object) createBlock(iterationCount int, stakeMap map[int]int) (*Block, error) {

	//更新梯度
	weight_accumlator_mat := mat.NewDense(1, Communication_object.numFeatures, make([]float64, Communication_object.numFeatures))
	weight_accumlator_temp := mat.NewDense(1, Communication_object.numFeatures, make([]float64, Communication_object.numFeatures))
	NoisedDelta := make([]float64, Communication_object.numFeatures)

	////////////outlier Detect
	if Communication_object.blockUpdates[0].Iteration == 0 {
		//第一轮检测使用基于数据量加权的梯度和
		for _, update := range Communication_object.blockUpdates {
			all_train_data_num += update.NumData
			copy(NoisedDelta, update.NoisedDelta)
			weight_accumlator_temp = mat.NewDense(1, Communication_object.numFeatures, NoisedDelta)
			weight_accumlator_temp.MulElem(weight_accumlator_temp, GenerateGoNumData(Communication_object.numFeatures, update.NumData))
			weight_accumlator_mat.Add(weight_accumlator_mat, weight_accumlator_temp)
		}
		all_train_data_num_mat := GenerateGoNumData(Communication_object.numFeatures, all_train_data_num)
		weight_accumlator_mat.DivElem(weight_accumlator_mat, all_train_data_num_mat)
		// outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":weight_accumlator_mat", weight_accumlator_mat)

		outlier_index = byzantineTest(Communication_object.blockUpdates, matConverFloat64List(weight_accumlator_mat, Communication_object.numFeatures))

		//关键数据重制
		all_train_data_num = 0
		weight_accumlator_mat = mat.NewDense(1, Communication_object.numFeatures, make([]float64, Communication_object.numFeatures))
	} else {
		outlier_index = byzantineTest(Communication_object.blockUpdates, client.bc.getLatestSmoothed_Grade())
	}

	if len(outlier_index) == 0 {
		outLog.Println("对等节点 " + strconv.Itoa(client.id) + ":smoothed_grad_detect异常点检测完成,无节点被判定为拜占庭节点 ,全部被Accept")
		for _, update := range Communication_object.blockUpdates {
			update.Accepted = true
			NodeStack := stakeMap[update.SourceID]
			stakeMap[update.SourceID] = NodeStack + STAKE_UNIT //全部得到奖励
		}

	} else {
		outLog.Println("对等节点 "+strconv.Itoa(client.id)+":经smoothed_grad_detect异常检测得出的问题梯度节点在列表中的下标为:", outlier_index)
		outLog.Println("对等节点 " + strconv.Itoa(client.id) + ":开始排除拜占庭节点上传的梯度")
		var sig bool

		for index, update := range Communication_object.blockUpdates {
			sig = dataIn(index, outlier_index)
			if sig {
				outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":对等节点%d被判断为拜占庭节点", update.SourceID)
				Communication_object.blockUpdates[index].Accepted = false
				NodeStack := stakeMap[update.SourceID]
				stakeMap[update.SourceID] = NodeStack - STAKE_UNIT //惩罚
			} else {
				Communication_object.blockUpdates[index].Accepted = true
				NodeStack := stakeMap[update.SourceID]
				stakeMap[update.SourceID] = NodeStack + STAKE_UNIT
			}
		}
	}

	////////////根据接受结果计算本轮更新梯度
	all_train_data_num = 0
	num_update_accpted := 0

	runtime.LockOSThread()
	_gstate := python3.PyGILState_Ensure()
	PyresetVar.CallFunctionObjArgs()
	for _, update := range Communication_object.blockUpdates {
		if update.Accepted {
			num_update_accpted++
		}
	}

	for _, update := range Communication_object.blockUpdates {
		if update.Accepted {
			all_train_data_num += update.NumData

			argArray := python3.PyList_New(len(update.NoisedDelta))
			for i := 0; i < len(update.NoisedDelta); i++ {
				python3.PyList_SetItem(argArray, i, python3.PyFloat_FromDouble(update.NoisedDelta[i]))
			}
			args := python3.PyTuple_New(3)
			param1 := python3.PyLong_FromDouble(update.NumData)
			param2 := python3.PyLong_FromLong(num_update_accpted)
			python3.PyTuple_SetItem(args, 0, argArray)
			python3.PyTuple_SetItem(args, 1, param1)
			python3.PyTuple_SetItem(args, 2, param2)

			PygradCollector.CallObject(args)
			// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":returna", python3.PyLong_AsLong(returna))
			// copy(NoisedDelta, update.NoisedDelta)
			// weight_accumlator_temp = mat.NewDense(1, Communication_object.numFeatures, NoisedDelta)
			// weight_accumlator_temp.MulElem(weight_accumlator_temp, GenerateGoNumData(Communication_object.numFeatures, update.NumData))
			// weight_accumlator_mat.Add(weight_accumlator_mat, weight_accumlator_temp)
		}
	}
	// TrimmedGrad := getTrimmedGrad(num_update_accpted, byzantineNum, Communication_object.numFeatures)

	args := python3.PyTuple_New(3)
	param1 := python3.PyLong_FromLong(num_update_accpted)
	param2 := python3.PyLong_FromLong(byzantineNum)
	param3 := python3.PyLong_FromLong(Communication_object.numFeatures)
	python3.PyTuple_SetItem(args, 0, param1)
	python3.PyTuple_SetItem(args, 1, param2)
	python3.PyTuple_SetItem(args, 2, param3)
	var gradAfterTrim *python3.PyObject
	gradAfterTrim = PygradTrim.CallObject(args)

	PyByteArray := python3.PyByteArray_FromObject(gradAfterTrim)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":PyByteArray", PyByteArray)
	GoChar := python3.PyByteArray_AsString(PyByteArray)
	goByteArray := []byte(GoChar)

	var TrimmedGrad []float64
	size := len(goByteArray) / 8
	for i := 0; i < size; i++ {
		currIndex := i * 8
		bits := binary.LittleEndian.Uint64(goByteArray[currIndex : currIndex+8])
		aFloat := math.Float64frombits(bits)
		TrimmedGrad = append(TrimmedGrad, aFloat)
	}

	python3.PyGILState_Release(_gstate)

	weight_accumlator_temp = mat.NewDense(1, Communication_object.numFeatures, TrimmedGrad)
	weight_accumlator_mat.Add(weight_accumlator_mat, weight_accumlator_temp)

	outLog.Println("对等节点 "+strconv.Itoa(client.id)+":通过检验的对等节点的总数据量为", all_train_data_num)
	all_train_data_num_mat := GenerateGoNumData(Communication_object.numFeatures, all_train_data_num)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":weight_accumlator_mat_(1)", weight_accumlator_mat)
	weight_accumlator_mat.DivElem(weight_accumlator_mat, all_train_data_num_mat)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":weight_accumlator_mat(2)", weight_accumlator_mat)

	////////////FedBMUF for Golang
	//全局参数
	global_weights := make([]float64, Communication_object.numFeatures)
	global_weights = Communication_object.bc.getLatestGradient()
	global_weights_mat := mat.NewDense(1, Communication_object.numFeatures, global_weights)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":B_global_weights_mat", global_weights_mat)

	//Temp
	var Temp *mat.Dense

	// Smoothed_grad
	Smoothed_grad := Communication_object.bc.getLatestSmoothed_Grade()
	Smoothed_grad_mat := mat.NewDense(1, Communication_object.numFeatures, Smoothed_grad)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":B_Smoothed_grad_mat", Smoothed_grad_mat)

	// block_momentum ,block_lr
	block_momentum_mat := GenerateGoNumData(Communication_object.numFeatures, block_momentum)
	block_lr_mat := GenerateGoNumData(Communication_object.numFeatures, block_lr)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":block_momentum_mat", block_momentum_mat)

	//kernel
	///////////////////// (1)smoothed_grad = block_momentum * smoothed_grad + block_lr * weight_accumlator
	Smoothed_grad_mat.MulElem(Smoothed_grad_mat, block_momentum_mat)
	weight_accumlator_mat.MulElem(weight_accumlator_mat, block_lr_mat)
	Smoothed_grad_mat.Add(Smoothed_grad_mat, weight_accumlator_mat)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":Smoothed_grad_mat(1)", Smoothed_grad_mat)

	///////////////////// (2)temp.data.copy_(global_weights + smoothed_grad)
	global_weights_mat.Add(global_weights_mat, Smoothed_grad_mat)
	Temp = matCopy(global_weights_mat, Communication_object.numFeatures)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":Temp(2)", Temp)

	///////////////////// (3)use_nbm : temp.data.copy_(temp + block_momentum * smoothed_grad)
	block_momentum_mat.MulElem(block_momentum_mat, Smoothed_grad_mat) //block_momentum作为中间变量存储结果 ， 下一轮会重新设置冲量 ， 不影响
	Temp.Add(Temp, block_momentum_mat)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":Temp(3)", Temp)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":block_momentum_mat", block_momentum_mat)

	///////////////////// (4)global_weights.copy(Temp)
	global_weights_mat = matCopy(Temp, Communication_object.numFeatures)
	// outLog.Println("对等节点 "+strconv.Itoa(client.id)+":global_weights_mat(4)", global_weights_mat)

	//取出最后的结果
	finalWeights := make([]float64, Communication_object.numFeatures)
	mat.Row(finalWeights, 0, global_weights_mat)
	finalSmoothedGrad := make([]float64, Communication_object.numFeatures)
	mat.Row(finalSmoothedGrad, 0, Smoothed_grad_mat)

	updatesGathered := make([]NewDelta, len(Communication_object.blockUpdates))
	copy(updatesGathered, Communication_object.blockUpdates)

	bData := BlockData{iterationCount, finalWeights, finalSmoothedGrad, updatesGathered}
	Communication_object.bc.AddBlock(bData, stakeMap)

	newBlock := Communication_object.bc.Blocks[len(Communication_object.bc.Blocks)-1]

	//Miner reward
	NodeStack := stakeMap[client.id]
	stakeMap[client.id] = NodeStack + STAKE_UNIT*2 //全部得到奖励

	return newBlock, nil

}

//链上填块
func (Communication_object *Communication_object) addBlock(newBlock Block) error {

	// if already exists don't create/replace it
	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":尝试为第%d次迭代添加新的区块", newBlock.Data.Iteration)

	// 全同步更新  ，  必为 nil
	if Communication_object.bc.getBlock(newBlock.Data.Iteration) != nil {

		better := Communication_object.evaluateBlockQuality(newBlock)
		if !better {
			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":已经有区块了，且之前添加的区块为正确区块")
			return blockExistsError
		} else {
			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":已经有区块了，但是需要为此轮次添加的区块较之前的区块更佳")
			Communication_object.replaceBlock(newBlock, newBlock.Data.Iteration)
			return nil
		}

	} else {

		outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":为本轮次添加区块成功！")
		Communication_object.bc.AddBlockMsg(newBlock)
		return nil

	}

}

//在每一轮训练之前,清空梯度记录
func (Communication_object *Communication_object) flushUpdates() {

	Communication_object.blockUpdates = Communication_object.blockUpdates[:0]
}

//检查本轮迭代是不是已经添加过区块
func (Communication_object *Communication_object) hasBlock(iterationCount int) bool {

	if Communication_object.bc.getBlock(iterationCount) != nil {
		return true
	} else {
		return false
	}

}

//检查获得到的旧区块是不是比我现在的更好
func (Communication_object *Communication_object) evaluateBlockQuality(block Block) bool {

	myBlock := Communication_object.bc.getBlock(block.Data.Iteration)
	previousBlock := Communication_object.bc.getBlock(block.Data.Iteration - 1)

	// hash 检测
	if string(block.PrevBlockHash[:]) != string(previousBlock.Hash[:]) {
		outLog.Println("不一致哈希. 区块中记录前一个区块哈希值为:", block.PrevBlockHash[:], ",前一个区块的哈希为:", previousBlock.Hash[:])
		return false
	} else if len(block.Data.Deltas) == 0 || len(myBlock.Data.Deltas) != 0 {
		return false
	}

	return true

}

//替换区块
func (Communication_object *Communication_object) replaceBlock(block Block, iterationCount int) {

	*Communication_object.bc.Blocks[iterationCount+1] = block

}

//替换区块链
func (Communication_object *Communication_object) replaceChain(chain Blockchain) int {
	*Communication_object.bc = chain
	return chain.Blocks[len(chain.Blocks)-1].Data.Iteration //最后一个被挖出的块 已经是第几次迭代
}

//模型Test acc准确率检测
func testModelTestAcc(weights []float64) float64 {

	outLog.Println("对等节点 "+strconv.Itoa(client.id)+":需要计算测试集准确率的参数长度为", len(weights))
	runtime.LockOSThread()

	_gstate := python3.PyGILState_Ensure()

	argArray := python3.PyList_New(len(weights))

	for i := 0; i < len(weights); i++ {
		python3.PyList_SetItem(argArray, i, python3.PyFloat_FromDouble(weights[i]))
	}

	var TestAcc float64

	args := python3.PyTuple_New(1)
	python3.PyTuple_SetItem(args, 0, argArray)
	pyTestResult := pyTestAcc.CallObject(args)

	TestAcc = python3.PyFloat_AsDouble(pyTestResult)

	python3.PyGILState_Release(_gstate)

	return TestAcc

}

//检查错误
func check(e error) {

	if e != nil {
		panic(e)
	}

}
