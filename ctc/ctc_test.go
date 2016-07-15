package ctc

import (
	"math"
	"math/rand"
	"testing"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/num-analysis/linalg"
)

const (
	testSymbolCount = 5
	testPrecision   = 1e-5

	benchLabelLen    = 50
	benchSeqLen      = 500
	benchSymbolCount = 30
)

func TestLogLikelihoodOutputs(t *testing.T) {
	for i := 0; i < 10; i++ {
		labelLen := 5 + rand.Intn(5)
		seqLen := labelLen + rand.Intn(5)
		label := make([]int, labelLen)
		for i := range label {
			label[i] = rand.Intn(testSymbolCount)
		}
		seq, resSeq, rresSeq := createTestSequence(seqLen, testSymbolCount)
		expected := exactLikelihood(seq, label, -1)
		actual := math.Exp(LogLikelihood(resSeq, label).Output()[0])
		rActual := math.Exp(LogLikelihoodR(rresSeq, label).Output()[0])
		if math.Abs(actual-expected)/math.Abs(expected) > testPrecision {
			t.Errorf("LogLikelihood gave log(%e) but expected log(%e)",
				actual, expected)
		}
		if math.Abs(rActual-expected)/math.Abs(expected) > testPrecision {
			t.Errorf("LogLikelihoodR gave log(%e) but expected log(%e)",
				rActual, expected)
		}
	}
}

func BenchmarkLogLikelihood(b *testing.B) {
	label := make([]int, benchLabelLen)
	for i := range label {
		label[i] = rand.Intn(testSymbolCount)
	}
	_, resSeq, _ := createTestSequence(benchSeqLen, benchSymbolCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogLikelihood(resSeq, label)
	}
}

func createTestSequence(seqLen, symCount int) (seq []linalg.Vector,
	res []autofunc.Result, rres []autofunc.RResult) {
	res = make([]autofunc.Result, seqLen)
	rres = make([]autofunc.RResult, seqLen)
	seq = make([]linalg.Vector, seqLen)
	for i := range seq {
		seq[i] = make(linalg.Vector, symCount+1)
		var probSum float64
		for j := range seq[i] {
			seq[i][j] = rand.Float64()
			probSum += seq[i][j]
		}
		for j := range seq[i] {
			seq[i][j] /= probSum
		}
		logVec := make(linalg.Vector, len(seq[i]))
		res[i] = &autofunc.Variable{
			Vector: logVec,
		}
		for j := range logVec {
			logVec[j] = math.Log(seq[i][j])
		}
		rres[i] = &autofunc.RVariable{
			Variable:   res[i].(*autofunc.Variable),
			ROutputVec: make(linalg.Vector, len(logVec)),
		}
	}
	return
}

func exactLikelihood(seq []linalg.Vector, label []int, lastSymbol int) float64 {
	if len(seq) == 0 {
		if len(label) == 0 {
			return 1
		} else {
			return 0
		}
	}

	next := seq[0]
	blank := len(next) - 1

	var res float64
	res += next[blank] * exactLikelihood(seq[1:], label, -1)
	if lastSymbol >= 0 {
		res += next[lastSymbol] * exactLikelihood(seq[1:], label, lastSymbol)
	}
	if len(label) > 0 {
		res += next[label[0]] * exactLikelihood(seq[1:], label[1:], label[0])
	}
	return res
}
