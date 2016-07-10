package mfcc

import "math"

type melBin struct {
	startIdx  int
	middleIdx int
	endIdx    int
}

func (m melBin) Apply(powers []float64) float64 {
	var res float64
	for i := m.startIdx + 1; i < m.middleIdx; i++ {
		dist := float64(i-m.startIdx) / float64(m.middleIdx-m.startIdx)
		res += dist * powers[i]
	}
	for i := m.middleIdx; i < m.endIdx; i++ {
		dist := float64(i-m.middleIdx) / float64(m.endIdx-m.middleIdx)
		res += (1 - dist) * powers[i]
	}
	return res
}

type melBinner []melBin

func newMelBinner(fftSize, sampleRate, binCount int, minFreq, maxFreq float64) melBinner {
	if hardMax := float64(sampleRate) / 2; maxFreq > hardMax {
		maxFreq = hardMax
	}

	minMels, maxMels := hertzToMels(minFreq), hertzToMels(maxFreq)

	points := make([]float64, binCount+2)
	points[0] = minMels
	for i := 1; i <= binCount; i++ {
		points[i] = minMels + float64(i)*(maxMels-minMels)/float64(binCount+1)
	}
	points[binCount+1] = maxMels

	fftPoints := make([]int, len(points))
	for i, m := range points {
		fftPoints[i] = hertzToBin(melsToHertz(m), fftSize, sampleRate)
	}

	res := make(melBinner, binCount)
	for i := range res {
		res[i] = melBin{
			startIdx:  fftPoints[i],
			middleIdx: fftPoints[i+1],
			endIdx:    fftPoints[i+2],
		}
	}
	return res
}

func (m melBinner) Apply(f fftBins) []float64 {
	powers := f.powerSpectrum()
	res := make([]float64, len(m))
	for i, b := range m {
		res[i] = b.Apply(powers)
	}
	return res
}

func hertzToMels(h float64) float64 {
	return 1125.0 * math.Log(1+h/700)
}

func melsToHertz(m float64) float64 {
	return 700 * (math.Exp(m/1125) - 1)
}

func hertzToBin(h float64, fftSize, sampleRate int) int {
	freqScale := float64(sampleRate) / float64(fftSize)
	bin := h / freqScale
	bin = math.Min(bin, float64(fftSize)/2)
	floorFreq := math.Floor(bin) * freqScale
	ceilFreq := math.Ceil(bin) * freqScale
	if math.Abs(floorFreq-h) < math.Abs(ceilFreq-h) {
		return int(bin)
	} else {
		return int(math.Ceil(bin))
	}
}
