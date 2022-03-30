package main

import (
	"fmt"
	// "encoding/binary"
	"bytes"
	"encoding/gob"
)

type BlockData struct {
	Iteration      int
	GlobalW        []float64
	Smoothed_Grade []float64
	Deltas         []NewDelta
}

//BlockData的构造函数
func NewBlockData(iteration int, globalW []float64, smoothed_Grade []float64, delta []NewDelta) *BlockData {
	blockData := &BlockData{iteration, globalW, smoothed_Grade, delta}
	return blockData
}

//BlockData转换为String
func (blockdata BlockData) String() string {
	return fmt.Sprintf("迭代轮次: %d, 全局参数: %s,全局更新: %s,对等节点上传梯度: %s", blockdata.Iteration, arrayToString(blockdata.GlobalW, ","), arrayToString(blockdata.Smoothed_Grade, ","), arrayToStringUpdate(blockdata.Deltas, ","))
}

//BlockData转换为字符串数组
func (blockdata BlockData) ToByte() []byte {

	var blockDataBytes bytes.Buffer
	enc := gob.NewEncoder(&blockDataBytes)
	err := enc.Encode(blockdata)
	if err != nil {
		fmt.Println("encode error:", err)
	}

	return blockDataBytes.Bytes()
}

func arrayToStringUpdate(a []NewDelta, delim string) string {

	updates := "["
	numUpdates := len(a)
	for i := 0; i < numUpdates; i++ {
		updates += a[i].String()
		if i != numUpdates-1 {
			updates += " " + delim
		}
	}
	updates += "]"
	return updates
}
