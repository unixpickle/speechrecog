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
	temp := make([]float64, len(signal))
	signalCopy := make([]float64, len(signal))
	copy(signalCopy, signal)

	basePeriod := 2 * math.Pi / float64(len(signal))
	sines := make([]float64, len(signal)/4)
	cosines := make([]float64, len(signal)/4)
	for i := range cosines {
		if i < 2 || i%100 == 0 {
			cosines[i] = math.Cos(basePeriod * float64(i))
			sines[i] = math.Sin(basePeriod * float64(i))
		} else {
			// Angle sum formulas for cosine and sine.
			cosines[i] = cosines[i-1]*cosines[1] - sines[i-1]*sines[1]
			sines[i] = sines[i-1]*cosines[1] + cosines[i-1]*sines[1]
		}
	}

	return destructiveFFT(signalCopy, temp, sines, cosines, 0)
}

func destructiveFFT(signal []float64, temp []float64, sines, cosines []float64,
	depth uint) fftBins {
	n := len(signal)
	if n == 1 {
		return fftBins{Cos: []float64{signal[0]}}
	} else if n == 2 {
		return fftBins{
			Cos: []float64{signal[0] + signal[1], signal[0] - signal[1]},
		}
	} else if n == 4 {
		return fftBins{
			Cos: []float64{
				signal[0] + signal[1] + signal[2] + signal[3],
				signal[0] - signal[2],
				signal[0] - signal[1] + signal[2] - signal[3],
			},
			Sin: []float64{
				signal[1] - signal[3],
			},
		}
	} else if n&1 != 0 {
		panic("input must be a power of 2")
	}

	evenSignal := temp[:n/2]
	oddSignal := temp[n/2:]
	for i := 0; i < n; i += 2 {
		evenSignal[i>>1] = signal[i]
		oddSignal[i>>1] = signal[i+1]
	}
	evenBins := destructiveFFT(evenSignal, signal[:n/2], sines, cosines, depth+1)
	oddBins := destructiveFFT(oddSignal, signal[n/2:], sines, cosines, depth+1)

	res := fftBins{
		Cos: temp[:n/2+1],
		Sin: temp[n/2+1:],
	}

	res.Cos[0] = evenBins.Cos[0] + oddBins.Cos[0]
	res.Cos[n/2] = evenBins.Cos[0] - oddBins.Cos[0]
	res.Cos[n/4] = evenBins.Cos[n/4]
	for i := 1; i < n/4; i++ {
		oddPart := cosines[i<<depth]*oddBins.Cos[i] -
			sines[i<<depth]*oddBins.Sin[i-1]
		res.Cos[i] = evenBins.Cos[i] + oddPart
		res.Cos[n/2-i] = evenBins.Cos[i] - oddPart
	}

	res.Sin[n/4-1] = oddBins.Cos[len(oddBins.Cos)-1]
	for i := 0; i < n/4-1; i++ {
		oddPart := sines[(i+1)<<depth]*oddBins.Cos[i+1] +
			cosines[(i+1)<<depth]*oddBins.Sin[i]
		res.Sin[i] = evenBins.Sin[i] + oddPart
		res.Sin[len(res.Sin)-(i+1)] = -evenBins.Sin[i] + oddPart
	}

	return res
}
