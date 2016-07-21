package ctc

import (
	"testing"

	"github.com/unixpickle/num-analysis/linalg"
)

func TestPrefixSearch(t *testing.T) {
	var seqs = [][]linalg.Vector{
		{
			{-9.21034037197618, -0.000100005000333347},
			{-0.105360515657826, -2.302585092994046},
			{-9.21034037197618, -0.000100005000333347},
			{-0.105360515657826, -2.302585092994046},
			{-9.21034037197618, -0.000100005000333347},
			{-9.21034037197618, -0.000100005000333347},
		},
		{
			{-1.38155105579643e+01, -1.38155105579643e+01, -2.00000199994916e-06},
			// The first label is not more likely, but
			// after both timesteps it has a 0.64% chance
			// of being seen in at least one of the two
			// timesteps.
			{-0.916290731874155, -13.815510557964274, -0.510827290434046},
			{-0.916290731874155, -13.815510557964274, -0.510827290434046},
			{-1.38155105579643e+01, -1.38155105579643e+01, -2.00000199994916e-06},
			{-1.609437912434100, -0.693147180559945, -1.203972804325936},
		},
		{
			{-1.38155105579643e+01, -1.38155105579643e+01, -2.00000199994916e-06},
			{-0.916290731874155, -13.815510557964274, -0.510827290434046},
			{-1.38155105579643e+01, -1.38155105579643e+01, -2.00000199994916e-06},
			{-1.609437912434100, -0.693147180559945, -1.203972804325936},
		},
		{
			{-0.916290731874155, -13.815510557964274, -0.510827290434046},
			{-1.38155105579643e+01, -1.38155105579643e+01, -2.00000199994916e-06},
			{-1.609437912434100, -0.693147180559945, -1.203972804325936},
		},
	}
	var outputs = [][]int{
		{0, 0},
		{0, 1},
		{1},
		{1},
	}
	var threshes = []float64{-1e-2, -1e-3, -1e-6, -1e-10}
	for _, thresh := range threshes {
		for i, seq := range seqs {
			actual := PrefixSearch(seq, thresh)
			expected := outputs[i]
			if !labelingsEqual(actual, expected) {
				t.Errorf("thresh %f: seq %d: expected %v got %v", thresh, i,
					expected, actual)
			}
		}
	}
}

func labelingsEqual(l1, l2 []int) bool {
	if len(l1) != len(l2) {
		return false
	}
	for i, x := range l1 {
		if x != l2[i] {
			return false
		}
	}
	return true
}
