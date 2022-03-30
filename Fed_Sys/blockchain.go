package main

import (
	"fmt"
	"os"
)

type Blockchain struct {
	Blocks []*Block
}

//AddBlock 给区块链添加一个块
func (bc *Blockchain) AddBlock(data BlockData, stakeMap map[int]int) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1] //取最后一个区块
	newBlock := NewBlock(data, prevBlock.Hash, stakeMap)
	bc.Blocks = append(bc.Blocks, newBlock)
}

func NewGenesisBlock(InitWeight []float64, numFeatures int) *Block {
	return GenesisBlock(InitWeight, numFeatures)
}

func NewBlockchain(InitWeight []float64, numFeatures int) *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock(InitWeight, numFeatures)}}
}

func (bc *Blockchain) getLatestBlock() *Block {
	return bc.Blocks[len(bc.Blocks)-1]
}

//最后一个块的全局梯度
func (bc *Blockchain) getLatestGradient() []float64 {

	prevBlock := bc.Blocks[len(bc.Blocks)-1] //取最后一个块
	gradient := make([]float64, len(prevBlock.Data.GlobalW))
	copy(gradient, prevBlock.Data.GlobalW)
	return gradient
}

//最后一个块的Smooth_grad
func (bc *Blockchain) getLatestSmoothed_Grade() []float64 {

	prevBlock := bc.Blocks[len(bc.Blocks)-1] //取最后一个块
	Smoothed_Grade := make([]float64, len(prevBlock.Data.Smoothed_Grade))
	copy(Smoothed_Grade, prevBlock.Data.Smoothed_Grade)
	return Smoothed_Grade
}

func (bc *Blockchain) PrintChain() {

	for _, block := range bc.Blocks {
		outLog.Printf("前节点哈希: %x\n", block.PrevBlockHash)
		outLog.Printf("区块数据: %s\n", block.Data.String())
		outLog.Printf("本节点哈希: %x\n", block.Hash)
		// fmt.Printf("Stake: %v\n", block.StakeMap)
		fmt.Println()
	}

}

//获得最后一个块的Hash值
func (bc *Blockchain) getLatestBlockHash() []byte {
	outLog.Printf("最后一个块的hash值为: %x", bc.Blocks[len(bc.Blocks)-1].Hash)
	return bc.Blocks[len(bc.Blocks)-1].Hash
}

func (bc *Blockchain) verifyBlock(block Block) bool {

	return true
}

func (bc *Blockchain) AddBlockMsg(newBlock Block) {

	appendBlock := &Block{Timestamp: newBlock.Timestamp,
		Data:          newBlock.Data,
		PrevBlockHash: newBlock.PrevBlockHash,
		Hash:          newBlock.Hash,
		StakeMap:      newBlock.StakeMap}

	bc.Blocks = append(bc.Blocks, appendBlock)
}

//检测其他节点是否已经添加block
func (bc *Blockchain) getBlock(iterationCount int) *Block {

	if len(bc.Blocks) >= (iterationCount + 2) {
		if bc.Blocks[iterationCount+1].Data.Iteration != iterationCount {

			outLog.Printf("已追加用于多个迭代的块")
			bc.PrintChain()
			os.Exit(1)
		}
		return bc.Blocks[iterationCount+1]

	} else {
		return nil
	}

}
