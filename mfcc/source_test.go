package mfcc

import (
	"io"
	"math"
	"testing"
)

type sliceSource struct {
	vec      []float64
	idx      int
	buffSize int
}

func (s *sliceSource) ReadSamples(slice []float64) (n int, err error) {
	if len(slice) > s.buffSize {
		slice = slice[:s.buffSize]
	}
	n = copy(slice, s.vec[s.idx:])
	if n == 0 {
		err = io.EOF
	}
	s.idx += n
	return
}

func TestFramer(t *testing.T) {
	source := sliceSource{vec: []float64{1, -1, 0.5, 0.3, 0.2, 1, 0.5}, buffSize: 2}
	framedSource := Framer{S: &source, Size: 3, Step: 2}

	var data [11]float64
	n, err := framedSource.ReadSamples(data[:])
	if err != io.EOF {
		t.Errorf("expected EOF error, got %v", err)
	}
	expected := []float64{1, -1, 0.5, 0.5, 0.3, 0.2, 0.2, 1, 0.5, 0.5}
	if n != 10 {
		t.Errorf("expected 10 outputs but got %d", n)
	} else if !slicesClose(data[:10], expected) {
		t.Errorf("expected slice %v but got %v", expected, data[:10])
	}

	source = sliceSource{vec: []float64{1, -1, 0.5, 0.3, 0.2, 1}, buffSize: 2}
	framedSource = Framer{S: &source, Size: 3, Step: 3}

	n, err = framedSource.ReadSamples(data[:])
	if err != io.EOF {
		t.Errorf("expected EOF error, got %v", err)
	}
	expected = []float64{1, -1, 0.5, 0.3, 0.2, 1}
	if n != 6 {
		t.Errorf("expected 6 outputs but got %d", n)
	} else if !slicesClose(data[:6], expected) {
		t.Errorf("expected slice %v but got %v", expected, data[:6])
	}
}

func slicesClose(s1, s2 []float64) bool {
	for i, x := range s1 {
		if math.Abs(s2[i]-x) > 1e-5 {
			return false
		}
	}
	return true
}
