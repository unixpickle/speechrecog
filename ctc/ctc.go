// Package ctc implements Connectionist Temporal Coding
// for training models (typically neural networks) to
// predict output sequences.
//
// For more on CTC, check out the paper:
// ftp://ftp.idsia.ch/pub/juergen/icml2006.pdf.
package ctc

import (
	"math"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/num-analysis/linalg"
)

// LogLikelihood computes the log likelihood of the
// label given an output sequence of the logs of
// output probabilities.
//
// The last entry of each output vector is the log of
// the probability of the blank symbol.
//
// Each element in the label is an index corresponding
// to elements of the output vectors (e.g. a label 2
// corresponds to whatever symbol is represented by
// element 2 of the output vectors).
func LogLikelihood(seq []autofunc.Result, label []int) autofunc.Result {
	if len(seq) == 0 {
		if len(label) == 0 {
			return &autofunc.Variable{Vector: []float64{0}}
		} else {
			return &autofunc.Variable{Vector: []float64{math.Inf(-1)}}
		}
	}

	// positionProbs stores the log probabilities of
	// being at every position in the blank-infused
	// label, where blanks are injected at the start
	// and end of the label, and between entries.
	var positionProbs autofunc.Result

	initProbs := make(linalg.Vector, len(label)*2+1)
	initProbs[0] = 0
	for i := 1; i < len(initProbs); i++ {
		initProbs[i] = math.Inf(-1)
	}
	positionProbs = &autofunc.Variable{
		Vector: initProbs,
	}

	for _, input := range seq {
		positionProbs = autofunc.Pool(positionProbs, func(last autofunc.Result) autofunc.Result {
			resParts := make([]autofunc.Result, len(label)*2+1)
			resParts[0] = mulProbabilities(vectorEntry(last, 0), vectorEntry(input, -1))
			for i := 2; i < len(label)*2+1; i += 2 {
				resParts[i] = mulProbabilities(vectorEntry(input, -1),
					addProbabilities(vectorEntry(last, i-1), vectorEntry(last, i)))
			}
			if len(label) > 0 {
				resParts[1] = mulProbabilities(vectorEntry(input, label[0]),
					addProbabilities(vectorEntry(last, 0), vectorEntry(last, 1)))
			}
			for i := 3; i < len(label)*2+1; i += 2 {
				positionSum := addProbabilities(addProbabilities(vectorEntry(last, i),
					vectorEntry(last, i-1)), vectorEntry(last, i-2))
				labelIdx := (i - 1) / 2
				resParts[i] = mulProbabilities(vectorEntry(input, label[labelIdx]), positionSum)
			}
			return autofunc.Concat(resParts...)
		})
	}

	return addProbabilities(vectorEntry(positionProbs, -1), vectorEntry(positionProbs, -2))
}

// LogLikelihoodR is like LogLikelihood, but with
// r-operator support.
func LogLikelihoodR(seq []autofunc.RResult, label []int) autofunc.RResult {
	if len(seq) == 0 {
		if len(label) == 0 {
			return &autofunc.RVariable{
				Variable:   &autofunc.Variable{Vector: []float64{0}},
				ROutputVec: []float64{0},
			}
		} else {
			return &autofunc.RVariable{
				Variable:   &autofunc.Variable{Vector: []float64{0}},
				ROutputVec: []float64{math.Inf(-1)},
			}
		}
	}

	var positionProbs autofunc.RResult

	initProbs := make(linalg.Vector, len(label)*2+1)
	initProbs[0] = 0
	for i := 1; i < len(initProbs); i++ {
		initProbs[i] = math.Inf(-1)
	}
	positionProbs = &autofunc.RVariable{
		Variable:   &autofunc.Variable{Vector: initProbs},
		ROutputVec: make(linalg.Vector, len(initProbs)),
	}

	for _, input := range seq {
		positionProbs = autofunc.PoolR(positionProbs, func(last autofunc.RResult) autofunc.RResult {
			resParts := make([]autofunc.RResult, len(label)*2+1)
			resParts[0] = mulProbabilitiesR(vectorEntryR(last, 0), vectorEntryR(input, -1))
			for i := 2; i < len(label)*2+1; i += 2 {
				resParts[i] = mulProbabilitiesR(vectorEntryR(input, -1),
					addProbabilitiesR(vectorEntryR(last, i-1), vectorEntryR(last, i)))
			}
			if len(label) > 0 {
				resParts[1] = mulProbabilitiesR(vectorEntryR(input, label[0]),
					addProbabilitiesR(vectorEntryR(last, 0), vectorEntryR(last, 1)))
			}
			for i := 3; i < len(label)*2+1; i += 2 {
				positionSum := addProbabilitiesR(addProbabilitiesR(vectorEntryR(last, i),
					vectorEntryR(last, i-1)), vectorEntryR(last, i-2))
				labelIdx := (i - 1) / 2
				resParts[i] = mulProbabilitiesR(vectorEntryR(input, label[labelIdx]), positionSum)
			}
			return autofunc.ConcatR(resParts...)
		})
	}

	return addProbabilitiesR(vectorEntryR(positionProbs, -1), vectorEntryR(positionProbs, -2))
}

// vectorEntry returns a Result for the i-th entry in
// an autofunc.Result.
// If i is negative, then the length of the vector is
// added to it.
func vectorEntry(vec autofunc.Result, i int) autofunc.Result {
	if i < 0 {
		i += len(vec.Output())
	}
	return autofunc.Slice(vec, i, i+1)
}

func vectorEntryR(vec autofunc.RResult, i int) autofunc.RResult {
	if i < 0 {
		i += len(vec.Output())
	}
	return autofunc.SliceR(vec, i, i+1)
}

// mulProbabilities multiplies two probabilities given
// their logarithms and returns the new log probability.
func mulProbabilities(p1, p2 autofunc.Result) autofunc.Result {
	if math.IsInf(p1.Output()[0], -1) {
		return p1
	} else if math.IsInf(p2.Output()[0], -1) {
		return p2
	}
	return autofunc.Add(p1, p2)
}

func mulProbabilitiesR(p1, p2 autofunc.RResult) autofunc.RResult {
	if math.IsInf(p1.Output()[0], -1) {
		return p1
	} else if math.IsInf(p2.Output()[0], -1) {
		return p2
	}
	return autofunc.AddR(p1, p2)
}

// addProbabilities adds two probabilities given their
// logarithms and returns the new log probability.
func addProbabilities(p1, p2 autofunc.Result) autofunc.Result {
	if math.IsInf(p1.Output()[0], -1) {
		return p2
	} else if math.IsInf(p2.Output()[0], -1) {
		return p1
	}
	normalizer := math.Max(p1.Output()[0], p2.Output()[0])
	offset1 := autofunc.AddScaler(p1, -normalizer)
	offset2 := autofunc.AddScaler(p2, -normalizer)
	exp := autofunc.Exp{}
	exp1 := exp.Apply(offset1)
	exp2 := exp.Apply(offset2)
	sumLog := autofunc.Log{}.Apply(autofunc.Add(exp1, exp2))
	return autofunc.AddScaler(sumLog, normalizer)
}

func addProbabilitiesR(p1, p2 autofunc.RResult) autofunc.RResult {
	if math.IsInf(p1.Output()[0], -1) {
		return p2
	} else if math.IsInf(p2.Output()[0], -1) {
		return p1
	}
	normalizer := math.Max(p1.Output()[0], p2.Output()[0])
	offset1 := autofunc.AddScalerR(p1, -normalizer)
	offset2 := autofunc.AddScalerR(p2, -normalizer)
	exp := autofunc.Exp{}
	exp1 := exp.ApplyR(autofunc.RVector{}, offset1)
	exp2 := exp.ApplyR(autofunc.RVector{}, offset2)
	sumLog := autofunc.Log{}.ApplyR(autofunc.RVector{}, autofunc.AddR(exp1, exp2))
	return autofunc.AddScalerR(sumLog, normalizer)
}
