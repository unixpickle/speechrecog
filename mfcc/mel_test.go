package mfcc

import (
	"math"
	"testing"
)

func TestMelBin(t *testing.T) {
	powers := []float64{1, 2, 3, 4, 5, 6, 7}
	bin := melBin{startIdx: 1, middleIdx: 3, endIdx: 6}
	res := bin.Apply(powers)
	expected := 3*0.5 + 4 + 5*2.0/3 + 6*1.0/3
	if math.Abs(res-expected) > 1e-5 {
		t.Errorf("expected %f got %f", expected, res)
	}
}

func TestMelBinner(t *testing.T) {
	binner := newMelBinner(8, 16, 2, 2, 8)
	actual := binner.Apply(fftBins{
		Cos: []float64{2, 3, 4, 5, 6},
		Sin: []float64{0, 0, 0},
	})
	expected := []float64{4.0 * 4 / 8, 5.0 * 5 / 8}
	if len(actual) != len(expected) {
		t.Errorf("expected len %d got len %d", len(expected), len(actual))
	} else if !slicesClose(actual, expected) {
		t.Errorf("expected %v got %v", expected, actual)
	}
}
