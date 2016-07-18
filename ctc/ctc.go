// Package ctc implements Connectionist Temporal
// Classification for training models (typically
// neural networks) to predict output sequences.
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
		for i := 1; i < len(label)*2+1; i += 2 {
			labelIdx := (i - 1) / 2
			positionSum := addProbabilitiesFloat(last[i], last[i-1])
			if labelIdx > 0 && label[labelIdx-1] != label[labelIdx] {
				positionSum = addProbabilitiesFloat(positionSum,
					last[i-2])
			}
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

	for _, inputRes := range seq {
		input := inputRes.Output()
		last := positionProbs.Output()
		inputR := inputRes.ROutput()
		lastR := positionProbs.ROutput()
		newProbs := make(linalg.Vector, len(positionProbs.Output()))
		newProbsR := make(linalg.Vector, len(positionProbs.Output()))
		newProbs[0] = last[0] + input[len(input)-1]
		newProbsR[0] = lastR[0] + inputR[len(input)-1]
		for i := 2; i < len(label)*2+1; i += 2 {
			newProbs[i], newProbsR[i] = addProbabilitiesFloatR(last[i-1], lastR[i-1],
				last[i], lastR[i])
			newProbs[i] += input[len(input)-1]
			newProbsR[i] += inputR[len(input)-1]
		}
		for i := 1; i < len(label)*2+1; i += 2 {
			labelIdx := (i - 1) / 2
			posSum, posSumR := addProbabilitiesFloatR(last[i-1], lastR[i-1],
				last[i], lastR[i])
			if labelIdx > 0 && label[labelIdx-1] != label[labelIdx] {
				posSum, posSumR = addProbabilitiesFloatR(last[i-2], lastR[i-2],
					posSum, posSumR)
			}
			newProbs[i] = input[label[labelIdx]] + posSum
			newProbsR[i] = inputR[label[labelIdx]] + posSumR
		}
		positionProbs = &logLikelihoodRStep{
			OutputVec:  newProbs,
			ROutputVec: newProbsR,
			LastProbs:  positionProbs,
			SeqIn:      inputRes,
			Label:      label,
		}
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
	for i := 1; i < len(l.Label)*2+1; i += 2 {
		labelIdx := (i - 1) / 2
		inputGrad[l.Label[labelIdx]] += upstream[i]
		if labelIdx > 0 && l.Label[labelIdx-1] != l.Label[labelIdx] {
			a := addProbabilitiesFloat(last[i-2], last[i-1])
			b := last[i]
			da, db := productSumPartials(a, b, upstream[i])
			lastGrad[i] += db
			da, db = productSumPartials(last[i-2], last[i-1], da)
			lastGrad[i-2] += da
			lastGrad[i-1] += db
		} else {
			da, db := productSumPartials(last[i-1], last[i], upstream[i])
			lastGrad[i-1] += da
			lastGrad[i] += db
		}
	}

	if !l.LastProbs.Constant(g) {
		l.LastProbs.PropagateGradient(lastGrad, g)
	}
	if !l.SeqIn.Constant(g) {
		l.SeqIn.PropagateGradient(inputGrad, g)
	}
}

type logLikelihoodRStep struct {
	OutputVec  linalg.Vector
	ROutputVec linalg.Vector
	LastProbs  autofunc.RResult
	SeqIn      autofunc.RResult
	Label      []int
}

func (l *logLikelihoodRStep) Output() linalg.Vector {
	return l.OutputVec
}

func (l *logLikelihoodRStep) ROutput() linalg.Vector {
	return l.ROutputVec
}

func (l *logLikelihoodRStep) Constant(rg autofunc.RGradient, g autofunc.Gradient) bool {
	return l.SeqIn.Constant(rg, g) && l.LastProbs.Constant(rg, g)
}

func (l *logLikelihoodRStep) PropagateRGradient(upstream, upstreamR linalg.Vector,
	rg autofunc.RGradient, g autofunc.Gradient) {
	if l.Constant(rg, g) {
		return
	}

	last := l.LastProbs.Output()
	lastR := l.LastProbs.ROutput()
	input := l.SeqIn.Output()

	lastGrad := make(linalg.Vector, len(last))
	lastGradR := make(linalg.Vector, len(last))
	inputGrad := make(linalg.Vector, len(input))
	inputGradR := make(linalg.Vector, len(input))

	lastGrad[0] = upstream[0]
	lastGradR[0] = upstreamR[0]
	inputGrad[len(inputGrad)-1] = upstream[0]
	inputGradR[len(inputGrad)-1] = upstreamR[0]

	for i := 2; i < len(l.Label)*2+1; i += 2 {
		inputGrad[len(inputGrad)-1] += upstream[i]
		inputGradR[len(inputGrad)-1] += upstreamR[i]
		da, daR, db, dbR := productSumPartialsR(last[i-1], lastR[i-1], last[i], lastR[i],
			upstream[i], upstreamR[i])
		lastGrad[i-1] += da
		lastGrad[i] += db
		lastGradR[i-1] += daR
		lastGradR[i] += dbR
	}
	for i := 1; i < len(l.Label)*2+1; i += 2 {
		labelIdx := (i - 1) / 2
		inputGrad[l.Label[labelIdx]] += upstream[i]
		inputGradR[l.Label[labelIdx]] += upstreamR[i]
		if labelIdx > 0 && l.Label[labelIdx-1] != l.Label[labelIdx] {
			a, aR := addProbabilitiesFloatR(last[i-2], lastR[i-2], last[i-1], lastR[i-1])
			b, bR := last[i], lastR[i]
			da, daR, db, dbR := productSumPartialsR(a, aR, b, bR, upstream[i], upstreamR[i])
			lastGrad[i] += db
			lastGradR[i] += dbR
			da, daR, db, dbR = productSumPartialsR(last[i-2], lastR[i-2], last[i-1],
				lastR[i-1], da, daR)
			lastGrad[i-2] += da
			lastGrad[i-1] += db
			lastGradR[i-2] += daR
			lastGradR[i-1] += dbR
		} else {
			da, daR, db, dbR := productSumPartialsR(last[i-1], lastR[i-1], last[i],
				lastR[i], upstream[i], upstreamR[i])
			lastGrad[i-1] += da
			lastGradR[i-1] += daR
			lastGrad[i] += db
			lastGradR[i] += dbR
		}
	}

	if !l.LastProbs.Constant(rg, g) {
		l.LastProbs.PropagateRGradient(lastGrad, lastGradR, rg, g)
	}
	if !l.SeqIn.Constant(rg, g) {
		l.SeqIn.PropagateRGradient(inputGrad, inputGradR, rg, g)
	}
}

func productSumPartials(a, b, upstream float64) (da, db float64) {
	if math.IsInf(a, -1) && math.IsInf(b, -1) {
		return
	}
	denomLog := addProbabilitiesFloat(a, b)
	daLog := a - denomLog
	dbLog := b - denomLog
	da = upstream * math.Exp(daLog)
	db = upstream * math.Exp(dbLog)
	return
}

func productSumPartialsR(a, aR, b, bR, upstream, upstreamR float64) (da, daR, db, dbR float64) {
	if math.IsInf(a, -1) && math.IsInf(b, -1) {
		return
	}
	denomLog, denomLogR := addProbabilitiesFloatR(a, aR, b, bR)
	daExp := math.Exp(a - denomLog)
	dbExp := math.Exp(b - denomLog)
	da = upstream * daExp
	db = upstream * dbExp
	daR = upstreamR*daExp + upstream*daExp*(aR-denomLogR)
	dbR = upstreamR*dbExp + upstream*dbExp*(bR-denomLogR)
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

func addProbabilitiesFloatR(a, aR, b, bR float64) (res, resR float64) {
	if math.IsInf(a, -1) {
		return b, bR
	} else if math.IsInf(b, -1) {
		return a, aR
	}
	normalizer := math.Max(a, b)
	exp1 := math.Exp(a - normalizer)
	exp2 := math.Exp(b - normalizer)
	res = math.Log(exp1+exp2) + normalizer
	resR = (exp1*aR + exp2*bR) / (exp1 + exp2)
	return
}
