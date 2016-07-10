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
	var data [11]float64

	source := sliceSource{vec: []float64{1, -1, 0.5, 0.3, 0.2, 1, 0.5}, buffSize: 2}
	framedSource := framer{S: &source, Size: 3, Step: 2}

	n, err := framedSource.ReadSamples(data[:])
	if err != io.EOF {
		t.Errorf("expected EOF error, got %v", err)
	}
	expected := []float64{1, -1, 0.5, 0.5, 0.3, 0.2, 0.2, 1, 0.5, 0.5}
	if n != len(expected) {
		t.Errorf("expected %d outputs but got %d", len(expected), n)
	} else if !slicesClose(data[:len(expected)], expected) {
		t.Errorf("expected slice %v but got %v", expected, data[:len(expected)])
	}

	source = sliceSource{vec: []float64{1, -1, 0.5, 0.3, 0.2, 1}, buffSize: 2}
	framedSource = framer{S: &source, Size: 3, Step: 3}

	n, err = framedSource.ReadSamples(data[:])
	if err != io.EOF {
		t.Errorf("expected EOF error, got %v", err)
	}
	expected = []float64{1, -1, 0.5, 0.3, 0.2, 1}
	if n != len(expected) {
		t.Errorf("expected %d outputs but got %d", len(expected), n)
	} else if !slicesClose(data[:len(expected)], expected) {
		t.Errorf("expected slice %v but got %v", expected, data[:len(expected)])
	}
}

func TestrateChanger(t *testing.T) {
	var data [20]float64

	source := sliceSource{vec: []float64{1, -1, 0.5, 0.3, 0.2, 1, 0.5}, buffSize: 2}
	changer := rateChanger{S: &source, Ratio: 2 + 1e-8}

	n, err := changer.ReadSamples(data[:])
	if err != io.EOF {
		t.Errorf("expected EOF error, got %v", err)
	}
	expected := []float64{1, 0, -1, -0.25, 0.5, 0.4, 0.3, 0.25, 0.2, 0.6, 1, 0.75, 0.5}
	if n != len(expected) {
		t.Errorf("expected %d outputs but got %d", len(expected), n)
	} else if !slicesClose(data[:len(expected)], expected) {
		t.Errorf("expected slice %v but got %v", expected, data[:len(expected)])
	}

	source = sliceSource{vec: []float64{1, -1, 0.5, 0.3, 0.2, 1, 0.5}, buffSize: 2}
	changer = rateChanger{S: &source, Ratio: 0.5 + 1e-8}

	n, err = changer.ReadSamples(data[:])
	if err != io.EOF {
		t.Errorf("expected EOF error, got %v", err)
	}
	expected = []float64{1, 0.5, 0.2, 0.5}
	if n != len(expected) {
		t.Errorf("expected %d outputs but got %d", len(expected), n)
	} else if !slicesClose(data[:len(expected)], expected) {
		t.Errorf("expected slice %v but got %v", expected, data[:len(expected)])
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
