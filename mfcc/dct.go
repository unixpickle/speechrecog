package mfcc

import "math"

// dct computes the first n bins of the discrete cosine
// transform of a signal.
func dct(signal []float64, n int) []float64 {
	res := make([]float64, n)
	baseFreq := math.Pi / float64(len(signal))
	for k := 0; k < n; k++ {
		for i, x := range signal {
			res[k] += x * math.Cos(baseFreq*float64(k)*(float64(i)+0.5))
		}
	}
	return res
}
