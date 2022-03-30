package main

import (
	"net"
	"strconv"
)

type Peer int

func (s *Peer) RegisterPeer(peerAddress net.TCPAddr, chain *Blockchain) error {

	outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ": 对等客户端 " + peerAddress.String() + " 请求注册并获取最新blockchain信息")
	peerLock.Lock()
	peerAddresses[peerLookup[peerAddress.String()]] = peerAddress

	stakeMap[peerLookup[peerAddress.String()]] = DEFAULT_STAKE
	peerLock.Unlock()

	// if I am first node (index:0) and I am waiting for a peer to join (iterationCount < 0) then send signal that I have at least one peer.
	if myPort == strconv.Itoa(basePort) && iterationCount < 0 {
		networkBootstrapped <- true
	}
	*chain = *client.bc

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+": 通知对等客户端 "+peerAddress.String()+" 目前本地区块链长度为 %d", len(chain.Blocks))

	return nil
}

func (s *Peer) RegisterBlock(block Block, returnBlock *Block) error {

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":获得第%d轮迭代中来自对等节点到新的区块信息\n", block.Data.Iteration)

	*returnBlock = block

	go processBlock(block)

	return nil

}

func (s *Peer) RegisterDelta(delta NewDelta, signal *bool) error {

	outLog.Printf("对等节点 "+strconv.Itoa(client.id)+":我作为聚合者，得到客户端聚合请求, 所收到梯度的迭代轮次为 %d\n", delta.Iteration)

	*signal = false
	go processDelta(delta)

	return nil

}
