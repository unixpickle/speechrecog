package mfcc

import (
	"math/rand"
	"testing"
)

const (
	dctBenchSignalSize = 26
	dctBenchBinCount   = 13
)

func TestDCT(t *testing.T) {
	inputs := [][]float64{
		[]float64{1, 2, 3, 4, 5, 6, 7, 8},
		[]float64{1, -2, 3, -4, 5, -6},
	}
	ns := []int{8, 3}
	outputs := [][]float64{
		[]float64{36.000000000000000, -12.884646045410273, -0.000000000000003,
			-1.346909601807877, 0.000000000000000, -0.401805807471995,
			-0.000000000000031, -0.101404645519244},
		[]float64{-3.00000000000000, 3.62346663143529, -3.46410161513775},
	}
	for i, input := range inputs {
		actual := dct(input, ns[i])
		expected := outputs[i]
		if len(actual) != len(expected) {
			t.Errorf("%d: expected len %d got len %d", i, len(expected), len(actual))
		} else if !slicesClose(actual, expected) {
			t.Errorf("%d: expected %v got %v", i, expected, actual)
		}
	}
}

func BenchmarkDCT(b *testing.B) {
	rand.Seed(123)
	input := make([]float64, dctBenchSignalSize)
	for i := range input {
		input[i] = rand.NormFloat64()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dct(input, dctBenchBinCount)
	}
}
