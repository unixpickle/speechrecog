package ctc

import "github.com/unixpickle/num-analysis/linalg"

// BestPath performs best path decoding on the sequence.
func BestPath(seq []linalg.Vector) []int {
	last := -1
	var res []int
	for _, vec := range seq {
		idx := maxIdx(vec)
		if idx == len(vec)-1 {
			last = -1
		} else if idx != last {
			last = idx
			res = append(res, idx)
		}
	}
	return res
}

func maxIdx(vec linalg.Vector) int {
	var maxVal float64
	var maxIdx int
	for i, x := range vec {
		if i == 0 || x >= maxVal {
			maxVal = x
			maxIdx = i
		}
	}
	return maxIdx
}
