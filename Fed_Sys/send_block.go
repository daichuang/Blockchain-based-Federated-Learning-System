package main

import (
	"net"
	"net/rpc"
	"strconv"
	"time"
)

// Miner将这个迭代的块广播给所有的对等体
func sendBlock(block Block) {

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":传输本地第%d次迭代的区块\n", block.Data.Iteration)

	// 开启一个线程用于分散调用
	peerLock.Lock()

	ensureRPC.Add(len(peerAddresses))
	for _, address := range peerAddresses {
		go callRegisterBlockRPC(block, address)
	}
	//检查收敛，等待RPC调用返回并移动到新的迭代
	ensureRPC.Wait()

	peerLock.Unlock()

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":第%d次迭代对等节点间的新区块广播环节结束，使用测试集检测收敛", iterationCount)

	convergedLock.Lock()
	converged = client.checkConvergence()
	convergedLock.Unlock()

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":准备开始下一次迭代，当前迭代轮次为%d", iterationCount)

	nextCommunicationRound()

}

//RPC 把一个块送给他peer
func callRegisterBlockRPC(block Block, peerAddress net.TCPAddr) {

	defer ensureRPC.Done()
	var returnBlock Block
	c := make(chan error)

	conn, er := rpc.Dial("tcp", peerAddress.String())
	printError("rpc Dial", er)

	if er == nil {

		defer conn.Close()
		go func() { c <- conn.Call("Peer.RegisterBlock", block, &returnBlock) }()
		select {
		case err := <-c:

			outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":区块成功送达至对等节点: "+peerAddress.String()+" 当前迭代轮次为: %d", block.Data.Iteration)
			printError("传递区块时发生错误", err)

		case <-time.After(timeoutRPC):

			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + "传送区块超时 ， 在列表中删除节点")
			delete(peerAddresses, peerLookup[peerAddress.String()])

		}

	} else {

		delete(peerAddresses, peerLookup[peerAddress.String()])

		// ensureRPC <- true
		outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":对等服务端没有回应， 在列表中删除掉节点:" + peerAddress.String())

	}

}

// 检测收敛
func (Communication_object *Communication_object) checkConvergence() bool {

	//trainError := testModel(Communication_object.bc.getLatestGradient())

	//attackRate := testAttackRate(Communication_object.bc.getLatestGradient())
	//outLog.Printf(strconv.Itoa(Communication_object.id)+":Train Error is %.5f in Iteration %d",
	//	trainError, Communication_object.bc.Blocks[len(Communication_object.bc.Blocks)-1].Data.Iteration)

	//trainAcc := testModelTrainAcc(Communication_object.bc.getLatestGradient())
	testAcc := testModelTestAcc(Communication_object.bc.getLatestGradient())

	//outLog.Printf(strconv.Itoa(Communication_object.id)+":Train Acc is %.5f in Iteration %d",
	//	trainAcc, Communication_object.bc.Blocks[len(Communication_object.bc.Blocks)-1].Data.Iteration)
	outLog.Printf(strconv.Itoa(Communication_object.id)+":Test Acc is %.5f in Iteration %d",
		testAcc, Communication_object.bc.Blocks[len(Communication_object.bc.Blocks)-1].Data.Iteration)

	//outLog.Printf(strconv.Itoa(Communication_object.id)+":Attack Rate is %.5f in Iteration %d",
	//	attackRate, Communication_object.bc.Blocks[len(Communication_object.bc.Blocks)-1].Data.Iteration)

	//if trainError < convThreshold {
	//	return true
	//}

	return false
}
