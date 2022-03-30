package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

// go routine  监听传入的RPC调用
func messageListener(peerServer *rpc.Server, port string) {

	l, e := net.Listen("tcp", myPrivateIP+port)
	handleError("listen error", e)
	defer l.Close()

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+": 节点已启动. 监听端口 %s\n", port)

	for {
		conn, _ := l.Accept()
		//ServeConn在单个连接上执行server。ServeConn会阻塞，服务该连接直到客户端挂起。
		go peerServer.ServeConn(conn)
	}

}

//节点在进入系统时向所有其他节点宣布自己
func announceToNetwork(peerList []net.TCPAddr) {

	outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ": 向网络中其他节点宣布自己")
	myAddress, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(myIP+myPort))
	handleError("Resolve own address", err)

	for _, address := range peerList {
		outLog.Printf("对等节点 "+strconv.Itoa(client.id)+": 开始远程调用，对等节点的目标地址 %s", address.String())
		callRegisterPeerRPC(*myAddress, address)
	}

	//如果没找到邻居节点
	if len(peerAddresses) == 0 {
		outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ": 对等节点列表为空")
		os.Exit(1)
	}

	//outLog.Printf("对等节点 " + strconv.Itoa(client.id)+": Bootstrapped Network. Calling signal（networkBootstrapped）")
	networkBootstrapped <- true //释放 main goroutine
	outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ": 网络初始化结束， < -- signal")

}

//与登记节点进行远程调用，向其他节点宣布自己 ，洪泛
func callRegisterPeerRPC(myAddress net.TCPAddr, peerAddress net.TCPAddr) {

	// var chain Blockchain
	chain := Blockchain{[]*Block{}}
	c := make(chan error)

	conn, err := rpc.Dial("tcp", peerAddress.String())
	printError("对等节点暂时不在线 ", err)

	// 等待 node0 服务器 上线监听

	if (peerLookup[peerAddress.String()] == 0) && err != nil {

		outLog.Printf("对等节点 " + strconv.Itoa(client.id) + "等待对等节点 0 上线")

		for err != nil {
			time.Sleep(1000 * time.Millisecond)
			conn, err = rpc.Dial("tcp", peerAddress.String())
			printError("对等节点已掉线 ", err)
		}

	}

	if err == nil {

		outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ": 远程过程调用成功")

		defer conn.Close()
		//outLog.Printf("client " +strconv.Itoa(client.id)+": Calling RPC:"+ peerAddress.String())

		go func() {
			c <- conn.Call("Peer.RegisterPeer", myAddress, &chain) // 收到其他Peer节点的链信息
		}()

		outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ": 获取对等服务端 " + peerAddress.String() + " 的最新blockchain状态")

		select {

		case err = <-c:

			if err == nil {

				outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ": 向对等节点宣布自身，并获取得到区块链最新信息状态")

				//添加我的peer
				peerLock.Lock()
				peerAddresses[peerLookup[peerAddress.String()]] = peerAddress
				peerLock.Unlock()

				outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":收到的blockchain长度为 : %d", len(chain.Blocks))
				outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":本地blockchain长度为 : %d", len(client.bc.Blocks))

				//检查我现在的链是不是最长链 不是就换成我邻居节点告诉我的
				if len(chain.Blocks) > len(client.bc.Blocks) {
					boolLock.Lock()
					iterationCount = client.replaceChain(chain)
					if len(client.bc.Blocks)-2 != iterationCount {
						outLog.Printf("对等节点 "+strconv.Itoa(client.id)+"迭代轮次为 : %d", iterationCount)
						outLog.Printf("对等节点 "+strconv.Itoa(client.id)+"区块长度为 : %d", len(client.bc.Blocks))
						outLog.Printf("对等节点 " + strconv.Itoa(client.id) + "生成区块数量和迭代轮次不一致")
						os.Exit(1)
					}
					outLog.Printf("对等节点 " + strconv.Itoa(client.id) + "得到最新的链，且迭代轮次为 " + strconv.Itoa(iterationCount))
					boolLock.Unlock()
				}

			}

			// use err and result
		case <-time.After(timeoutPeer):

			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ": 得不到对等节点的回应: " + peerAddress.String())
		}
	}

}

func handleError(s string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s, err = %s\n", s, err.Error())
		os.Exit(1)
	}
}

func printError(msg string, e error) {
	if e != nil {
		errLog.Printf("%s, 错误信息为 %s ", msg, e.Error())
	}
}
