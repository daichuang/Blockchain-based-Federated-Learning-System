package main

import "fmt"

type NewDelta struct {
	SourceID int
}

func quickSort(blockUpdate []NewDelta, start, end int) {
	if start < end {
		i, j := start, end
		key := blockUpdate[(start+end)/2].SourceID
		for i <= j {
			for blockUpdate[i].SourceID < key {
				i++
			}
			for blockUpdate[j].SourceID > key {
				j--
			}
			if i <= j {
				blockUpdate[i], blockUpdate[j] = blockUpdate[j], blockUpdate[i]
				i++
				j--
			}
		}

		if start < j {
			quickSort(blockUpdate, start, j)
		}
		if end > i {
			quickSort(blockUpdate, i, end)
		}
	}
}

func main() {
	u1 := NewDelta{
		SourceID: 4,
	}
	u2 := NewDelta{
		SourceID: 3,
	}
	u3 := NewDelta{
		SourceID: 1,
	}
	u4 := NewDelta{
		SourceID: 2,
	}

	list := []NewDelta{}
	list = append(list, u1)
	list = append(list, u2)
	list = append(list, u3)
	list = append(list, u4)
	fmt.Println(list)

	quickSort(list, 0, len(list)-1)
	fmt.Println(list)

}
