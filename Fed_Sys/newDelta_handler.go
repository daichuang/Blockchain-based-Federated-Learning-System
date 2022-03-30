package main

import (
	"strconv"
)

func processDelta(newdelta NewDelta) {

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":处理来自%d号对等节点提交的第%d次迭代梯度的更新 , 本地目前迭代轮次为: %d\n", newdelta.SourceID, newdelta.Iteration, iterationCount)

	if iterationCount == newdelta.Iteration {

		updateLock.Lock()
		numberOfDelta := client.blockAddUpdate(newdelta)
		updateLock.Unlock()

		outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":作为聚合者, 我期待收到全部%d个节点更新, 目前得到%d个", numComputeNode, numberOfDelta)

		if numberOfDelta == numComputeNode {
			outLog.Printf("对等节点 "+strconv.Itoa(client.id)+": 收到第%d次迭代全部节点的更新", iterationCount)
			allUpdatesReceived <- true
		}

	}

}
