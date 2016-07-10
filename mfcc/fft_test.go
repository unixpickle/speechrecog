package mfcc

import (
	"math/rand"
	"testing"
)

const fftBenchSize = 512

func TestFFT(t *testing.T) {
	inputs := [][]float64{
		[]float64{0.517450},
		[]float64{0.517450, -0.515357},
		[]float64{0.517450, 0.591873, 0.104983, -0.512010},
		[]float64{0.517450, 0.591873, 0.104983, -0.512010, -0.037091, 0.203369,
			0.452477, -0.452457, 0.873007, 0.134188, -0.515357, 0.864060, 0.838039, 0.618038,
			-0.729226, 0.949877},
	}
	outputs := [][]float64{
		[]float64{0.517450},
		[]float64{0.0020930, 1.0328070},
		[]float64{0.70230, 0.41247, 0.54257, 1.10388},
		[]float64{3.901220, 0.598021, 0.624881, 1.641404, 2.878528, -1.558631,
			0.554137, -2.103022, -0.892656, -1.616822, -0.303837, 1.961911, 0.697998,
			-2.336823, -0.036587, -2.415035},
	}
	for i, input := range inputs {
		expected := outputs[i]
		res := fft(input)
		actual := append(res.Cos, res.Sin...)
		if len(actual) != len(expected) {
			t.Errorf("%d: len should be %d but it's %d", i, len(expected), len(actual))
		} else if !slicesClose(actual, expected) {
			t.Errorf("%d: expected %v but got %v", i, expected, actual)
		}
	}
}

func BenchmarkFFT(b *testing.B) {
	rand.Seed(123)
	inputVec := make([]float64, fftBenchSize)
	for i := range inputVec {
		inputVec[i] = rand.NormFloat64()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fft(inputVec)
	}
}
