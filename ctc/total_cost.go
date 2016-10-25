package ctc

import (
	"runtime"
	"sort"
	"sync"

	"github.com/unixpickle/autofunc/seqfunc"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/sgd"
)

// TotalCost returns total CTC cost of a network on
// a batch of samples.
//
// The maxGos argument specifies the maximum number
// of goroutines to run batches on simultaneously.
// If it is 0, GOMAXPROCS is used.
func TotalCost(f seqfunc.RFunc, s sgd.SampleSet, maxBatch, maxGos int) float64 {
	if maxGos == 0 {
		maxGos = runtime.GOMAXPROCS(0)
	}

	s = s.Copy()
	sortSampleSet(s)

	subBatches := make(chan sgd.SampleSet, s.Len()/maxBatch+1)
	for i := 0; i < s.Len(); i += maxBatch {
		bs := maxBatch
		if bs > s.Len()-i {
			bs = s.Len() - i
		}
		subBatches <- s.Subset(i, i+bs)
	}
	close(subBatches)

	var wg sync.WaitGroup
	costChan := make(chan float64, 0)
	for i := 0; i < maxGos; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range subBatches {
				costChan <- costForBatch(f, batch)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(costChan)
	}()

	var sum float64
	for c := range costChan {
		sum += c
	}
	return sum
}

func costForBatch(f seqfunc.RFunc, s sgd.SampleSet) float64 {
	inputVecs := make([][]linalg.Vector, s.Len())
	for i := 0; i < s.Len(); i++ {
		sample := s.GetSample(i).(Sample)
		inputVecs[i] = sample.Input
	}

	outputs := f.ApplySeqs(seqfunc.ConstResult(inputVecs))

	var sum float64
	for i, outSeq := range outputs.OutputSeqs() {
		seqVars := sequenceToVars(outSeq)
		label := s.GetSample(i).(Sample).Label
		sum += LogLikelihood(varsToResults(seqVars), label).Output()[0]
	}

	return -sum
}

// sortSampleSet sorts samples so that the longest
// sequences come first.
func sortSampleSet(s sgd.SampleSet) {
	sort.Sort(sampleSorter{s})
}

type sampleSorter struct {
	s sgd.SampleSet
}

func (s sampleSorter) Len() int {
	return s.s.Len()
}

func (s sampleSorter) Swap(i, j int) {
	s.s.Swap(i, j)
}

func (s sampleSorter) Less(i, j int) bool {
	item1 := s.s.GetSample(i).(Sample)
	item2 := s.s.GetSample(j).(Sample)
	return len(item1.Input) > len(item2.Input)
}
