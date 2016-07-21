package ctc

import (
	"math"
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
		subRes, _ := prefixSearch(sub, nil, math.Inf(-1), 0)
		res = append(res, subRes...)
	}
	return res
}

// prefixSearch performs prefix search starting from
// the first element of seq, given the existing prefix
// and the logged probabilities of that prefix with
// and without a terminating blank.
// It returns the best possible prefix and said prefix's
// probability.
func prefixSearch(seq []linalg.Vector, prefix []int, noBlankProb,
	blankProb float64) (bestSeq []int, bestProb float64) {
	if len(seq) == 0 {
		return prefix, addProbabilitiesFloat(noBlankProb, blankProb)
	}

	totalProb := addProbabilitiesFloat(noBlankProb, blankProb)

	var exts extensionList
	timeVec := seq[0]
	for i := 0; i < len(timeVec)-1; i++ {
		exts.Labels = append(exts.Labels, i)
		if len(prefix) > 0 && i == prefix[len(prefix)-1] {
			exts.Probs = append(exts.Probs, timeVec[i]+blankProb)
		} else {
			exts.Probs = append(exts.Probs, timeVec[i]+totalProb)
		}
	}

	exts.Labels = append(exts.Labels, -1)
	sameBlank := totalProb + timeVec[len(timeVec)-1]
	sameNoBlank := math.Inf(-1)
	if len(prefix) > 0 {
		last := prefix[len(prefix)-1]
		sameNoBlank = noBlankProb + timeVec[last]
	}
	exts.Probs = append(exts.Probs, addProbabilitiesFloat(sameNoBlank, sameBlank))

	sort.Sort(&exts)

	for i, addition := range exts.Labels {
		prob := exts.Probs[i]
		if i > 0 && prob < bestProb {
			continue
		}

		var s []int
		var p float64
		if addition == -1 {
			s, p = prefixSearch(seq[1:], prefix, sameNoBlank, sameBlank)
		} else {
			newPrefix := make([]int, len(prefix)+1)
			copy(newPrefix, prefix)
			newPrefix[len(prefix)] = addition
			s, p = prefixSearch(seq[1:], newPrefix, prob, math.Inf(-1))
		}
		if i == 0 || p > bestProb {
			bestProb = p
			bestSeq = s
		}
	}

	return
}

type prefixSearchResult struct {
	bestSeq       []int
	logLikelihood float64
}

type extensionList struct {
	Labels []int
	Probs  []float64
}

func (e *extensionList) Len() int {
	return len(e.Labels)
}

func (e *extensionList) Swap(i, j int) {
	e.Labels[i], e.Labels[j] = e.Labels[j], e.Labels[i]
	e.Probs[i], e.Probs[j] = e.Probs[j], e.Probs[i]
}

func (e *extensionList) Less(i, j int) bool {
	return e.Probs[i] > e.Probs[j]
}
