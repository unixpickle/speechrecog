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
//
// The result is only valid so long as the label slice
// is not changed by the caller.
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

	for _, inputRes := range seq {
		input := inputRes.Output()
		last := positionProbs.Output()
		newProbs := make(linalg.Vector, len(positionProbs.Output()))
		newProbs[0] = last[0] + input[len(input)-1]
		for i := 2; i < len(label)*2+1; i += 2 {
			newProbs[i] = addProbabilitiesFloat(last[i-1], last[i]) +
				input[len(input)-1]
		}
		if len(label) > 0 {
			newProbs[1] = addProbabilitiesFloat(last[0], last[1]) +
				input[label[0]]
		}
		for i := 3; i < len(label)*2+1; i += 2 {
			positionSum := addProbabilitiesFloat(last[i],
				addProbabilitiesFloat(last[i-2], last[i-1]))
			labelIdx := (i - 1) / 2
			newProbs[i] = input[label[labelIdx]] + positionSum
		}
		positionProbs = &logLikelihoodStep{
			OutputVec: newProbs,
			LastProbs: positionProbs,
			SeqIn:     inputRes,
			Label:     label,
		}
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

type logLikelihoodStep struct {
	OutputVec linalg.Vector
	LastProbs autofunc.Result
	SeqIn     autofunc.Result
	Label     []int
}

func (l *logLikelihoodStep) Output() linalg.Vector {
	return l.OutputVec
}

func (l *logLikelihoodStep) Constant(g autofunc.Gradient) bool {
	return l.SeqIn.Constant(g) && l.LastProbs.Constant(g)
}

func (l *logLikelihoodStep) PropagateGradient(upstream linalg.Vector, g autofunc.Gradient) {
	if l.Constant(g) {
		return
	}

	last := l.LastProbs.Output()
	input := l.SeqIn.Output()

	lastGrad := make(linalg.Vector, len(last))
	inputGrad := make(linalg.Vector, len(input))

	lastGrad[0] = upstream[0]
	inputGrad[len(inputGrad)-1] = upstream[0]

	for i := 2; i < len(l.Label)*2+1; i += 2 {
		inputGrad[len(inputGrad)-1] += upstream[i]
		da, db := productSumPartials(last[i-1], last[i], upstream[i])
		lastGrad[i-1] += da
		lastGrad[i] += db
	}
	if len(l.Label) > 0 {
		inputGrad[l.Label[0]] += upstream[1]
		da, db := productSumPartials(last[0], last[1], upstream[1])
		lastGrad[0] += da
		lastGrad[1] += db
	}
	for i := 3; i < len(l.Label)*2+1; i += 2 {
		labelIdx := (i - 1) / 2
		inputGrad[l.Label[labelIdx]] += upstream[i]
		a := addProbabilitiesFloat(last[i-2], last[i-1])
		b := last[i]
		da, db := productSumPartials(a, b, upstream[i])
		lastGrad[i] += db
		da, db = productSumPartials(last[i-2], last[i-1], da)
		lastGrad[i-2] += da
		lastGrad[i-1] += db
	}

	if !l.LastProbs.Constant(g) {
		l.LastProbs.PropagateGradient(lastGrad, g)
	}
	if !l.SeqIn.Constant(g) {
		l.SeqIn.PropagateGradient(inputGrad, g)
	}
}

func productSumPartials(a, b, upstream float64) (da, db float64) {
	aExp := math.Exp(a)
	bExp := math.Exp(b)
	denom := aExp + bExp
	da = upstream * aExp / denom
	db = upstream * bExp / denom
	return
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

func addProbabilitiesFloat(a, b float64) float64 {
	if math.IsInf(a, -1) {
		return b
	} else if math.IsInf(b, -1) {
		return a
	}
	normalizer := math.Max(a, b)
	exp1 := math.Exp(a - normalizer)
	exp2 := math.Exp(b - normalizer)
	return math.Log(exp1+exp2) + normalizer
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
