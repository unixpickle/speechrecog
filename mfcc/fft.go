package mfcc

import "math"

// fftBins stores the dot products of a signal with a
// basis of sinusoids.
//
// Let N be len(signal), and assume it is a power of 2.
// The i-th dot product in Cos, where i is between 0 and
// N/2 inclusive, are with is cos(2*pi/N*i).
// The j-th dot product in Sin, where j is between 0 and
// N/2-1 inclusive, are with sin(2*pi/N*i).
type fftBins struct {
	Cos []float64
	Sin []float64
}

// fft computes dot products of the signal with
// various sinusoids.
func fft(signal []float64) fftBins {
	n := len(signal)
	if n == 1 {
		return fftBins{Cos: []float64{signal[0]}}
	} else if n == 2 {
		return fftBins{
			Cos: []float64{signal[0] + signal[1], signal[0] - signal[1]},
		}
	} else if n&1 != 0 {
		panic("input must be a power of 2")
	}

	halfSignal := make([]float64, len(signal)/2)
	for i := 0; i < n; i += 2 {
		halfSignal[i>>1] = signal[i]
	}
	evenBins := fft(halfSignal)

	for i := 1; i < n; i += 2 {
		halfSignal[i>>1] = signal[i]
	}
	oddBins := fft(halfSignal)

	basePeriod := 2 * math.Pi / float64(n)

	res := fftBins{
		Cos: make([]float64, n/2+1),
		Sin: make([]float64, n/2-1),
	}

	res.Cos[0] = evenBins.Cos[0] + oddBins.Cos[0]
	for i := 1; i < n/4; i++ {
		res.Cos[i] = evenBins.Cos[i] + math.Cos(basePeriod*float64(i))*oddBins.Cos[i] -
			math.Sin(basePeriod*float64(i))*oddBins.Sin[i-1]
	}
	res.Cos[n/4] = evenBins.Cos[n/4]
	for i := 1; i < n/4; i++ {
		j := n/4 - i
		res.Cos[n/4+i] = evenBins.Cos[j] - math.Cos(basePeriod*float64(j))*oddBins.Cos[j] +
			math.Sin(basePeriod*float64(j))*oddBins.Sin[j-1]
	}
	res.Cos[n/2] = evenBins.Cos[0] - oddBins.Cos[0]

	for i := 0; i < n/4-1; i++ {
		res.Sin[i] = evenBins.Sin[i] + math.Sin(basePeriod*float64(i+1))*oddBins.Cos[i+1] +
			math.Cos(basePeriod*float64(i+1))*oddBins.Sin[i]
	}
	res.Sin[n/4-1] = oddBins.Cos[len(oddBins.Cos)-1]
	for i := 0; i < n/4-1; i++ {
		j := n/4 - (i + 2)
		res.Sin[i+n/4] = -evenBins.Sin[j] + math.Sin(basePeriod*float64(j+1))*oddBins.Cos[j+1] +
			math.Cos(basePeriod*float64(j+1))*oddBins.Sin[j]
	}

	return res
}
