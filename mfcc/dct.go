package mfcc

import "math"

// dct computes the first n bins of the discrete cosine
// transform of a signal.
func dct(signal []float64, n int) []float64 {
	res := make([]float64, n)
	baseFreq := math.Pi / float64(len(signal))
	for k := 0; k < n; k++ {
		initArg := baseFreq * float64(k) * 0.5
		curCos := math.Cos(initArg)
		curSin := math.Sin(initArg)

		// Double angle formulas to avoid more sin and cos.
		addCos := curCos*curCos - curSin*curSin
		addSin := 2 * curCos * curSin

		for _, x := range signal {
			res[k] += x * curCos
			// Angle sum formulas are a lot faster than
			// recomputing sines and cosines.
			curCos, curSin = curCos*addCos-curSin*addSin, curCos*addSin+addCos*curSin
		}
	}
	return res
}
