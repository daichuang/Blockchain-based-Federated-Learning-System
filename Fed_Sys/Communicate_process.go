package main

import (
	"os"
	"strconv"
	"time"
)

func nextCommunicationRound() {

	// totalUpdates = totalUpdates + len(client.bc.Blocks[len(client.bc.Blocks)-1].Data.Deltas)
	miner = false
	convergedLock.Lock()

	//Max_Communication_round = 100
	//converged = false
	// 收敛条件
	if converged {
		convergedLock.Unlock()
		time.Sleep(1000 * time.Millisecond)
		client.bc.PrintChain()
		os.Exit(1)
	} else {
		if iterationCount == MAX_ITERATIONS {

			outLog.Println("达到最大通信轮次")
			client.bc.PrintChain()
			os.Exit(1)
		}
	}

	updateLock.Lock()
	client.flushUpdates()
	updateLock.Unlock()

	convergedLock.Unlock()
	boolLock.Lock()
	peerLock.Lock()

	outLog.Printf("================== 开始新一轮通信迭代 , 轮次:%d ================== ", iterationCount+1)

	// 选择Miner
	minerId, idx := getRoles(stakeMap, numberOfNodes)

	if minerId == client.id {
		miner = true
	}

	outLog.Printf("矿工（聚合者）是 %d 号对等节点，在列表中的index是 %d \n", minerId, idx)

	for address, ID := range peerLookup {
		if minerId == ID {
			MinerAddressToConn = address
			break
		}
	}

	peerLock.Unlock()

	if miner {
		outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":身份为矿工（聚合者）. 当前迭代轮次是:%d", iterationCount+1)
		updateSent = true
		go AggregateDDL(iterationCount + 1)
	} else {
		outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":身份为计算节点. 当前迭代轮次是:%d", iterationCount+1)
		updateSent = false
		go BlockDDL(iterationCount + 1)
	}

	iterationCount++
	boolLock.Unlock()

	// portsToConnect = make([]string, len(peerPorts))
	// copy(portsToConnect, peerPorts)

}
