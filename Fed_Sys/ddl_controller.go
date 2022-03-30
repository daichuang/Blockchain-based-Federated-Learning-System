package main

import (
	"os"
	"strconv"
	"time"
)

// Miner聚合时间限制函数
func AggregateDDL(timer int) {

	select {

	case <-allUpdatesReceived:
		outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":在第%d迭代中收到所有更新，我在第%d迭代轮次，准备创建区块并发送",
			timer, iterationCount)

	case <-time.After(timeoutUpdate):
		outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":超时，没有在第%d迭代中收到预期的更新数量，准备创建不完整区块并发送 ", iterationCount)
	}

	if timer == iterationCount {

		if len(client.blockUpdates) > 0 {

			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":请求互斥访问区块链")
			blockChainLock.Lock()

			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":成功获得互斥访问区块链锁")
			blockToSend, err := client.createBlock(iterationCount, stakeMap) //需进一步实现

			blockChainLock.Unlock()
			printError("迭代轮次: "+strconv.Itoa(iterationCount), err)

			if err == nil {
				sendBlock(*blockToSend)
			}

		} else {
			outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":没有收到来自对等节点的更新，注销 ，Timer 轮次为 %d ，我的轮次为%d ", timer, iterationCount)
			os.Exit(1)
		}

	}
}

//非聚合者接受区块函数
func BlockDDL(timer int) {
	select {
	case <-blockReceived:
		if timer == iterationCount {
			outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":‘区块接收’信号量释放， 接收到的第 %d 次迭代的区块已经成功添加至链上", iterationCount)
		}

	case <-time.After(timeoutBlock):

		if timer == iterationCount {

			outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":超时，没有收到区块，在第%d次迭代追加空区块", timer)
			blockChainLock.Lock()
			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":请求互斥访问区块链 success (blockDDL)")
			blockToSend, err := client.createBlock(iterationCount, stakeMap)
			blockChainLock.Unlock()
			printError("迭代轮次: "+strconv.Itoa(iterationCount), err)
			if err == nil {
				sendBlock(*blockToSend)
			}

		}

	}

}
