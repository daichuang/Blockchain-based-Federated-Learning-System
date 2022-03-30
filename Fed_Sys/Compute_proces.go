package main

import (
	"net/rpc"
	"strconv"
	"time"
)

// 验证和聚合者在每轮中不贡献梯度
func ComputeProcess(ports []string) {

	for {

		if miner {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		boolLock.Lock()

		//聚合者在每轮中不贡献梯度
		if !updateSent {

			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":开始计算本轮的梯度更新\n")
			client.computeUpdate(iterationCount)

			//梯度交给聚合者
			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":计算完成，将梯度交给聚合者\n")
			sendDeltaToMiner(MinerAddressToConn) //送梯度

			if iterationCount == client.Delta.Iteration {
				updateSent = true
			}

		}

		if updateSent {
			boolLock.Unlock()
			time.Sleep(1000 * time.Millisecond)
		}

	}
}

func sendDeltaToMiner(MinerAdress string) {

	var signal bool
	c := make(chan error)

	conn, err := rpc.Dial("tcp", MinerAdress)

	for err != nil {

		conn, err = rpc.Dial("tcp", MinerAdress)
		printError("暂时无法连接至Miner，等待重新尝试", err)
		time.Sleep(1000 * time.Millisecond)
	}

	if err == nil {
		defer conn.Close()
		outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":成功连接至聚合者，当前迭代轮次为:%d\n", client.Delta.Iteration)
		go func() { c <- conn.Call("Peer.RegisterDelta", client.Delta, &signal) }()

		select {

		case err := <-c:
			printError("发送梯度出现错误", err)
			if err == nil {
				outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":已经将第%d次迭代的梯度送至聚合者处\n", client.Delta.Iteration)
			}

		case <-time.After(timeoutRPC):
			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":远程过程调用超时")

		}

	}

}
