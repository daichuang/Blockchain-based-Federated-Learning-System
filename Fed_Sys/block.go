package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64       //时间戳
	Data          BlockData   //区块存储的实际有效信息
	PrevBlockHash []byte      //存储前一个区块的哈希
	Hash          []byte      //当前区块的哈希值
	StakeMap      map[int]int //权益字典

}

//SetHash()计算区块哈希 并 计算到区块的哈希变量
//链接前一个区块哈希、区块时间戳、区块所存储真实数据的二进制码后，对其进行哈希得到当前区块的哈希值
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, timestamp, b.Data.ToByte()}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

//创建一个普通区块
func NewBlock(data BlockData, prevBlockHash []byte, stakeMap map[int]int) *Block {

	var blockTime int64

	if len(data.Deltas) == 0 {
		blockTime = 0
	} else {
		blockTime = time.Now().Unix() //返回int64
	}

	block := &Block{blockTime, data, prevBlockHash,
		[]byte{}, stakeMap}
	block.SetHash()

	return block
}

// create a globalWW with the appropriate number of features

func GenesisBlock(InitWeight []float64, numFeatures int) *Block {

	genesisBlockData := BlockData{-1, InitWeight, make([]float64, numFeatures), []NewDelta{}}

	block := &Block{0, genesisBlockData, []byte{}, []byte{}, map[int]int{}}
	block.SetHash()
	return block

}
