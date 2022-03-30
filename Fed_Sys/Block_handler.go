package main

import (
	"strconv"
)

// 非聚合者 接受区块 加到链上
func processBlock(block Block) {

	if iterationCount < 0 {
		return
	}

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":获取到第%d次迭代的区块信息\n", block.Data.Iteration)

	if !updateSent && block.Data.Iteration == iterationCount {
		updateSent = true
		outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":在第%d次迭代时释放计算节点", iterationCount)
	}

	//锁定，以确保迭代计数不会改变，直到我已经追加块
	outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":锁定，确保迭代计数不会改变，直到我添加区块")
	boolLock.Lock()

	outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":获得锁 ，开始处理区块")

	go addBlockToChain(block)

}

//处理接收到的区块并将其添加到链。
//当完成时移动到下一个迭代
func addBlockToChain(block Block) {

	blockChainLock.Lock()

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":为第%d次迭代添加区块 , 我目前处于第%d次迭代中\n",
		block.Data.Iteration, iterationCount)

	err := client.addBlock(block)

	blockChainLock.Unlock()

	if (block.Data.Iteration == iterationCount) && (err == nil) {

		// 更新权益列表
		if len(block.StakeMap) > 0 {
			stakeMap = block.StakeMap
			outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":新追加的区块中所携带新的权益列表为: %v", stakeMap)
		}

		if len(block.Data.Deltas) != 0 && updateSent && iterationCount >= 0 {

			if !miner {

				outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":为通道释放‘区块接收’信号量 ")
				blockReceived <- true
			}

		}

		boolLock.Unlock()

		outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":释放区块添加互斥锁 , 并将我收到的新区块告诉我的对等节点")
		go sendBlock(block) // client start next iter

	} else {

		boolLock.Unlock()
		outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":释放区块添加互斥锁")

	}

}
