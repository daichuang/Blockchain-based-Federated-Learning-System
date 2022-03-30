package main

import "math/rand"

//基于stakeMap，节点获得的彩票与其所持股份成比例
func getRoles(stakeMap map[int]int, totalNodes int) (int, int) {

	lottery := []int{}

	// 按股份大小加入同等的角色号码
	for nodeid := 0; nodeid < totalNodes; nodeid++ {
		stake := stakeMap[nodeid]
		for i := 0; i < stake; i++ {
			lottery = append(lottery, nodeid)
		}
	}

	var winner int

	outLog.Printf("各对等节点股份占比列表总长度为 %d \n", len(lottery))
	// outLog.Printf("股份列表为 %d \n", lottery)
	rand.Seed(int64(len(stakeMap)))
	winnerIdx := int(rand.Int63n(int64(len(stakeMap)*256)) % int64(len(lottery)))

	winner = lottery[winnerIdx]

	// outLog.Printf("矿工（聚合者）是 %d 号对等节点，在列表中的index是 %d \n", winner, winnerIdx)

	return winner, winnerIdx
}
