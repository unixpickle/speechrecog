package ctc

import (
	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/autofunc/seqfunc"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/neuralnet"
)

// A Sample is a labeled training sample.
type Sample struct {
	Input []linalg.Vector
	Label []int
}

// RGradienter computes gradients for sgd.SampleSets
// full of Samples.
type RGradienter struct {
	SeqFunc seqfunc.RFunc
	Learner sgd.Learner

	// MaxConcurrency is the maximum number of goroutines
	// to use simultaneously.
	MaxConcurrency int

	// MaxSubBatch is the maximum batch size to pass to
	// the SeqFunc in one call.
	MaxSubBatch int

	helper *neuralnet.GradHelper
}

func (r *RGradienter) Gradient(s sgd.SampleSet) autofunc.Gradient {
	s = s.Copy()
	sortSampleSet(s)
	return r.makeHelper().Gradient(s)
}

func (r *RGradienter) RGradient(v autofunc.RVector, s sgd.SampleSet) (autofunc.Gradient,
	autofunc.RGradient) {
	s = s.Copy()
	sortSampleSet(s)
	return r.makeHelper().RGradient(v, s)
}

func (r *RGradienter) makeHelper() *neuralnet.GradHelper {
	if r.helper != nil {
		r.helper.MaxConcurrency = r.MaxConcurrency
		r.helper.MaxSubBatch = r.MaxSubBatch
		return r.helper
	}
	return &neuralnet.GradHelper{
		MaxConcurrency: r.MaxConcurrency,
		MaxSubBatch:    r.MaxSubBatch,
		Learner:        r.Learner,
		CompGrad:       r.compGrad,
		CompRGrad:      r.compRGrad,
	}
}

func (r *RGradienter) compGrad(g autofunc.Gradient, s sgd.SampleSet) {
	inputVars := make([][]*autofunc.Variable, s.Len())
	for i := 0; i < s.Len(); i++ {
		sample := s.GetSample(i).(Sample)
		inputVars[i] = sequenceToVars(sample.Input)
	}

	outputs := r.SeqFunc.ApplySeqs(seqfunc.VarResult(inputVars))

	var upstream [][]linalg.Vector
	for i, outSeq := range outputs.OutputSeqs() {
		seqVars := sequenceToVars(outSeq)
		grad := autofunc.NewGradient(seqVars)
		label := s.GetSample(i).(Sample).Label
		cost := autofunc.Scale(LogLikelihood(varsToResults(seqVars), label), -1)
		cost.PropagateGradient(linalg.Vector{1}, grad)

		upstreamSeq := make([]linalg.Vector, len(seqVars))
		for i, variable := range seqVars {
			upstreamSeq[i] = grad[variable]
		}
		upstream = append(upstream, upstreamSeq)
	}

	outputs.PropagateGradient(upstream, g)
}

func (r *RGradienter) compRGrad(rv autofunc.RVector, rg autofunc.RGradient,
	g autofunc.Gradient, s sgd.SampleSet) {
	inputVars := make([][]*autofunc.Variable, s.Len())
	for i := 0; i < s.Len(); i++ {
		sample := s.GetSample(i).(Sample)
		inputVars[i] = sequenceToVars(sample.Input)
	}

	outputs := r.SeqFunc.ApplySeqsR(rv, seqfunc.VarRResult(rv, inputVars))

	var upstream [][]linalg.Vector
	var upstreamR [][]linalg.Vector
	for i, outSeq := range outputs.OutputSeqs() {
		seqRVars := sequenceToRVars(outSeq, outputs.ROutputSeqs()[i])
		params := varsInRVars(seqRVars)
		grad := autofunc.NewGradient(params)
		rgrad := autofunc.NewRGradient(params)
		label := s.GetSample(i).(Sample).Label
		cost := autofunc.ScaleR(LogLikelihoodR(rvarsToRResults(seqRVars), label), -1)
		cost.PropagateRGradient(linalg.Vector{1}, linalg.Vector{0}, rgrad, grad)

		upstreamSeq := make([]linalg.Vector, len(params))
		upstreamSeqR := make([]linalg.Vector, len(params))
		for i, variable := range params {
			upstreamSeq[i] = grad[variable]
			upstreamSeqR[i] = rgrad[variable]
		}
		upstream = append(upstream, upstreamSeq)
		upstreamR = append(upstreamR, upstreamSeqR)
	}

	outputs.PropagateRGradient(upstream, upstreamR, rg, g)
}

func sequenceToVars(seq []linalg.Vector) []*autofunc.Variable {
	res := make([]*autofunc.Variable, len(seq))
	for i, vec := range seq {
		res[i] = &autofunc.Variable{Vector: vec}
	}
	return res
}

func sequenceToRVars(seq, seqR []linalg.Vector) []*autofunc.RVariable {
	if seqR == nil && len(seq) > 0 {
		seqR = make([]linalg.Vector, len(seq))
		zeroVec := make(linalg.Vector, len(seq[0]))
		for i := range seqR {
			seqR[i] = zeroVec
		}
	}
	res := make([]*autofunc.RVariable, len(seq))
	for i, vec := range seq {
		res[i] = &autofunc.RVariable{
			Variable:   &autofunc.Variable{Vector: vec},
			ROutputVec: seqR[i],
		}
	}
	return res
}

func varsToResults(vars []*autofunc.Variable) []autofunc.Result {
	res := make([]autofunc.Result, len(vars))
	for i, v := range vars {
		res[i] = v
	}
	return res
}

func rvarsToRResults(vars []*autofunc.RVariable) []autofunc.RResult {
	res := make([]autofunc.RResult, len(vars))
	for i, v := range vars {
		res[i] = v
	}
	return res
}

func varsInRVars(vars []*autofunc.RVariable) []*autofunc.Variable {
	res := make([]*autofunc.Variable, len(vars))
	for i, x := range vars {
		res[i] = x.Variable
	}
	return res
}
