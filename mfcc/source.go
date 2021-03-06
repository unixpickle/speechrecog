package mfcc

import "io"

// A Source is a place from which audio sample data can
// be read.
// This interface is very similar to io.Reader, except
// that it deals with samples instead of bytes.
type Source interface {
	ReadSamples(s []float64) (n int, err error)
}

// A SliceSource is a Source which returns pre-determined
// samples from a slice.
type SliceSource struct {
	Slice []float64

	// Offset is the current offset into the slice.
	// This will be increased as samples are read.
	Offset int
}

func (s *SliceSource) ReadSamples(out []float64) (n int, err error) {
	n = copy(out, s.Slice[s.Offset:])
	s.Offset += n
	if n < len(out) {
		err = io.EOF
	}
	return
}

// A framer is a Source that wraps another Source and
// generates overlapping windows of sample data.
//
// For instance, if Size is set to 200 and Step is
// set to 100, then samples 0-199 from the wrapped
// source are returned, followed by samples 100-299,
// followed by 200-399, etc.
//
// The Step must not be greater than the Size.
//
// The last frame returned by a framer may be partial,
// i.e. it may be less than Size samples.
type framer struct {
	S Source

	Size int
	Step int

	doneError         error
	curCache          []float64
	nextCache         []float64
	outWindowProgress int
}

func (f *framer) ReadSamples(s []float64) (n int, err error) {
	if f.doneError != nil {
		return 0, f.doneError
	}
	for i := range s {
		var noSample bool
		s[i], noSample, err = f.readSample()
		if noSample {
			break
		}
		n++
		if err != nil {
			break
		}
	}
	if err != nil {
		f.doneError = err
	}
	return
}

func (f *framer) readSample() (sample float64, noSample bool, err error) {
	if len(f.curCache) > 0 {
		sample = f.curCache[0]
		f.curCache = f.curCache[1:]
	} else {
		var s [1]float64
		for {
			var n int
			n, err = f.S.ReadSamples(s[:])
			if n == 1 {
				sample = s[0]
				break
			} else if err != nil {
				return 0, true, err
			}
		}
	}
	if f.outWindowProgress >= f.Step {
		f.nextCache = append(f.nextCache, sample)
	}
	f.outWindowProgress++
	if f.outWindowProgress == f.Size {
		f.outWindowProgress = 0
		f.curCache = f.nextCache
		f.nextCache = nil
	}
	return
}

// A rateChanger changes the sample rate of a Source.
//
// The Ratio argument determines the ratio of the new
// sample rate to the old one.
// For example, a Ratio of 2.5 would turn the sample
// rate 22050 to the rate 55125.
type rateChanger struct {
	S     Source
	Ratio float64

	doneError  error
	started    bool
	lastSample float64
	nextSample float64
	midpart    float64
}

func (r *rateChanger) ReadSamples(s []float64) (n int, err error) {
	if r.doneError != nil {
		return 0, r.doneError
	}
	for i := range s {
		var noSample bool
		s[i], noSample, err = r.readSample()
		if noSample {
			break
		}
		n++
		if err != nil {
			break
		}
	}
	if err != nil {
		r.doneError = err
	}
	return
}

func (r *rateChanger) readSample() (sample float64, noSample bool, err error) {
	if !r.started {
		noSample, err = r.start()
		if noSample {
			return
		}
	}

	if r.midpart > 1 {
		readCount := int(r.midpart)
		for i := 0; i < readCount; i++ {
			noSample, err = r.readNext()
		}
		if noSample {
			return
		}
		r.midpart -= float64(readCount)
	}

	sample = r.lastSample*(1-r.midpart) + r.nextSample*r.midpart
	r.midpart += 1 / r.Ratio
	return
}

func (r *rateChanger) start() (noSample bool, err error) {
	var samples [2]float64
	var n, gotten int
	for gotten < 2 {
		n, err = r.S.ReadSamples(samples[gotten:])
		gotten += n
		if err != nil {
			break
		}
	}
	if gotten < 2 {
		return true, err
	}
	r.lastSample = samples[0]
	r.nextSample = samples[1]
	r.started = true
	return
}

func (r *rateChanger) readNext() (noSample bool, err error) {
	var samples [1]float64
	var n int
	for {
		n, err = r.S.ReadSamples(samples[:])
		if n == 1 {
			break
		} else if err != nil {
			return true, err
		}
	}
	r.lastSample = r.nextSample
	r.nextSample = samples[0]
	return
}
