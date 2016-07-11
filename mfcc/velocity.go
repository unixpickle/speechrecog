package mfcc

// AddVelocities generates a CoeffSource which wraps
// c and augments every vector of coefficients with
// an additional vector of coefficient velocities.
//
// For example, for input coefficients [a,b,c], the
// resulting source would produce coefficients
// [a,b,c,da,db,dc] where d stands for derivative.
func AddVelocities(c CoeffSource) CoeffSource {
	return &velocitySource{
		Wrapped: c,
	}
}

type velocitySource struct {
	Wrapped CoeffSource

	last      []float64
	lastLast  []float64
	doneError error
}

func (v *velocitySource) NextCoeffs() ([]float64, error) {
	if v.doneError != nil {
		return nil, v.doneError
	}

	if v.last == nil {
		v.lastLast, v.doneError = v.Wrapped.NextCoeffs()
		if v.doneError != nil {
			return nil, v.doneError
		}
		v.last, v.doneError = v.Wrapped.NextCoeffs()
		if v.doneError != nil {
			augmented := make([]float64, len(v.lastLast)*2)
			copy(augmented, v.lastLast)
			return augmented, nil
		}
		res := make([]float64, len(v.lastLast)*2)
		copy(res, v.lastLast)
		for i, x := range v.lastLast {
			res[i+len(v.lastLast)] = v.last[i] - x
		}
		return res, nil
	}

	var next []float64
	next, v.doneError = v.Wrapped.NextCoeffs()
	if v.doneError != nil {
		res := make([]float64, len(v.last)*2)
		copy(res, v.last)
		for i, x := range v.lastLast {
			res[i+len(v.last)] = v.last[i] - x
		}
		return res, nil
	}

	midpointRes := make([]float64, len(v.last)*2)
	copy(midpointRes, v.last)
	for i, x := range v.lastLast {
		midpointRes[i+len(v.last)] = (next[i] - x) / 2
	}

	v.lastLast = v.last
	v.last = next

	return midpointRes, nil
}
