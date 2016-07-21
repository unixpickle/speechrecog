package ctc

import (
	"sort"

	"github.com/unixpickle/num-analysis/linalg"
)

// PrefixSearch performs prefix search decoding on
// the output sequence.
//
// Any blanks with log likelihoods greater than
// or equal to blankThresh will be treated as if
// they have a log likelihood of 0 (i.e. a
// likelihood of 1).
func PrefixSearch(seq []linalg.Vector, blankThresh float64) []int {
	var subSeqs [][]linalg.Vector
	var subSeq []linalg.Vector
	for _, x := range seq {
		if x[len(x)-1] > blankThresh {
			if len(subSeq) > 0 {
				subSeqs = append(subSeqs, subSeq)
				subSeq = nil
			}
		} else {
			subSeq = append(subSeq, x)
		}
	}
	if len(subSeq) > 0 {
		subSeqs = append(subSeqs, subSeq)
	}

	var res []int
	for _, sub := range subSeqs {
		res = append(res, fullPrefixSearch(sub)...)
	}
	return res
}

func fullPrefixSearch(seq []linalg.Vector) []int {
	m := &prefixMap{
		List:     []*prefix{&prefix{EndsWithStop: true}},
		LogProbs: []float64{0},
	}

	for _, x := range seq {
		blank := len(x) - 1
		newMap := &prefixMap{}
		for i, pref := range m.List {
			prob := m.LogProbs[i]
			if pref.EndsWithStop {
				newMap.AddLogProb(pref, prob+x[blank])
			} else {
				newMap.AddLogProb(pref, prob+x[pref.Seq[len(pref.Seq)-1]])
				blankPref := pref.Copy()
				blankPref.EndsWithStop = true
				newMap.AddLogProb(blankPref, prob+x[blank])
			}
			for symIdx, symProb := range x[:blank] {
				newPref := pref.Copy()
				newPref.EndsWithStop = false
				newPref.Seq = append(newPref.Seq, symIdx)
				newMap.AddLogProb(newPref, prob+symProb)
			}
		}
		m = newMap
	}

	// Needed to combine sequences with and without
	// terminating blanks.
	resultMap := &prefixMap{}
	for i, pref := range m.List {
		prob := m.LogProbs[i]
		resultMap.AddLogProb(&prefix{Seq: pref.Seq}, prob)
	}

	var likeliest []int
	var bestLikelihood float64
	for i, likelihood := range resultMap.LogProbs {
		if i == 0 || likelihood > bestLikelihood {
			bestLikelihood = likelihood
			likeliest = resultMap.List[i].Seq
		}
	}
	return likeliest
}

type prefix struct {
	Seq          []int
	EndsWithStop bool
}

// Compare returns -1 if p < p1, 0 if p == p1, and
// 1 if p > p1.
// The comparison system is arbitrary but well-defined.
func (p *prefix) Compare(p1 *prefix) int {
	if p.EndsWithStop && !p1.EndsWithStop {
		return 1
	} else if !p.EndsWithStop && p1.EndsWithStop {
		return -1
	}
	if len(p.Seq) < len(p1.Seq) {
		return -1
	} else if len(p.Seq) > len(p1.Seq) {
		return 1
	}
	for i, x := range p.Seq {
		if x < p1.Seq[i] {
			return -1
		} else if x > p1.Seq[i] {
			return 1
		}
	}
	return 0
}

func (p *prefix) Copy() *prefix {
	res := &prefix{
		Seq:          make([]int, len(p.Seq)),
		EndsWithStop: p.EndsWithStop,
	}
	copy(res.Seq, p.Seq)
	return res
}

type prefixMap struct {
	List     []*prefix
	LogProbs []float64
}

func (p *prefixMap) AddLogProb(prefix *prefix, logProb float64) {
	idx := sort.Search(len(p.List), func(i int) bool {
		return p.List[i].Compare(prefix) >= 0
	})
	if idx == len(p.List) {
		p.List = append(p.List, prefix)
		p.LogProbs = append(p.LogProbs, logProb)
	} else if p.List[idx].Compare(prefix) == 0 {
		p.LogProbs[idx] = addProbabilitiesFloat(p.LogProbs[idx], logProb)
	} else {
		p.List = append(p.List, nil)
		p.LogProbs = append(p.LogProbs, 0)
		copy(p.List[idx+1:], p.List[idx:])
		copy(p.LogProbs[idx+1:], p.LogProbs[idx:])
		p.List[idx] = prefix
		p.LogProbs[idx] = logProb
	}
}
