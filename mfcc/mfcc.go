// Package mfcc can compute Mel-frequency cepstrum
// coefficients from raw sample data.
//
// For more information about MFCC, see Wikipedia:
// https://en.wikipedia.org/wiki/Mel-frequency_cepstrum
package mfcc

import (
	"math"
	"time"
)

const (
	DefaultWindow  = time.Millisecond * 20
	DefaultOverlap = time.Millisecond * 10

	DefaultFFTSize   = 512
	DefaultLowFreq   = 300
	DefaultHighFreq  = 8000
	DefaultMelCount  = 26
	DefaultKeepCount = 13
)

// Options stores all of the configuration options for
// computing MFCCs.
type Options struct {
	// Window is the amount of time represented in each
	// MFCC frame.
	// If this is 0, DefaultWindow is used.
	Window time.Duration

	// Overlap is the amount of overlapping time between
	// adjacent windows.
	// If this is 0, DefaultOverlap is used.
	Overlap time.Duration

	// DisableOverlap can be set to disable overlap.
	// If this is set to true, Overlap is ignored.
	DisableOverlap bool

	// FFTSize is the number of FFT bins to compute for
	// each window.
	// This must be a power of 2.
	// If this is 0, DefaultFFTSize is used.
	//
	// It may be noted that the FFT size influences the
	// upsampling/downsampling behavior of the converter.
	FFTSize int

	// LowFreq is the minimum frequency for Mel banks.
	// If this is 0, DefaultLowFreq is used.
	LowFreq float64

	// HighFreq is the maximum frequency for Mel banks.
	// If this is 0, DefaultHighFreq is used.
	// In practice, this may be bounded by the FFT window
	// size.
	HighFreq float64

	// MelCount is the number of Mel banks to compute.
	// If this is 0, DefaultMelCount is used.
	MelCount int

	// KeepCount is the number of MFCCs to keep after the
	// discrete cosine transform is complete.
	// If this is 0, DefaultKeepCount is used.
	KeepCount int
}

// CoeffSource computes MFCCs (or augmented MFCCs) from an
// underlying audio source.
type CoeffSource interface {
	// NextCoeffs returns the next batch of coefficients,
	// or an error if the underlying Source ended with one.
	//
	// This will never return a non-nil batch along with
	// an error.
	NextCoeffs() ([]float64, error)
}

// MFCC generates a CoeffSource that computes the MFCCs
// of the given Source.
//
// After source returns its first error, the last window
// will be padded with zeroes and used to compute a final
// batch of MFCCs before returning the error.
func MFCC(source Source, sampleRate int, options *Options) CoeffSource {
	if options == nil {
		options = &Options{}
	}

	windowTime := options.Window
	if windowTime == 0 {
		windowTime = DefaultWindow
	}
	windowSeconds := float64(windowTime) / float64(time.Second)

	fftSize := intOrDefault(options.FFTSize, DefaultFFTSize)
	newSampleRate := int(float64(fftSize)/windowSeconds + 0.5)

	overlapTime := options.Overlap
	if options.DisableOverlap {
		overlapTime = 0
	} else if overlapTime == 0 {
		overlapTime = DefaultOverlap
	}
	overlapSeconds := float64(overlapTime) / float64(time.Second)
	overlapSamples := int(overlapSeconds*float64(newSampleRate) + 0.5)
	if overlapSamples >= fftSize {
		overlapSamples = fftSize - 1
	}

	binCount := intOrDefault(options.MelCount, DefaultMelCount)
	minFreq := floatOrDefault(options.LowFreq, DefaultLowFreq)
	maxFreq := floatOrDefault(options.HighFreq, DefaultHighFreq)

	return &coeffChan{
		windowedSource: &framer{
			S: &rateChanger{
				S:     source,
				Ratio: float64(newSampleRate) / float64(sampleRate),
			},
			Size: fftSize,
			Step: fftSize - overlapSamples,
		},
		windowSize: fftSize,
		binner:     newMelBinner(fftSize, newSampleRate, binCount, minFreq, maxFreq),
		keepCount:  intOrDefault(options.KeepCount, DefaultKeepCount),
	}
}

type coeffChan struct {
	windowedSource Source
	windowSize     int
	binner         melBinner
	keepCount      int

	doneError error
}

func (c *coeffChan) NextCoeffs() ([]float64, error) {
	if c.doneError != nil {
		return nil, c.doneError
	}

	buf := make([]float64, c.windowSize)
	var have int
	for have < len(buf) && c.doneError == nil {
		n, err := c.windowedSource.ReadSamples(buf[:have])
		if err != nil {
			c.doneError = err
		}
		have += n
	}
	if have == 0 && c.doneError != nil {
		return nil, c.doneError
	}

	// ReadSamples can use the buffer as scratch space,
	// just like io.Reader.
	for i := have; i < len(buf); i++ {
		buf[i] = 0
	}

	banks := c.binner.Apply(fft(buf))
	for i, x := range banks {
		banks[i] = math.Log(x)
	}
	return dct(banks, c.keepCount), nil
}

func intOrDefault(val, def int) int {
	if val == 0 {
		return def
	} else {
		return val
	}
}

func floatOrDefault(val, def float64) float64 {
	if val == 0 {
		return def
	} else {
		return val
	}
}
