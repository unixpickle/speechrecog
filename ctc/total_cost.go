package ctc

import (
	"runtime"
	"sync"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/rnn"
)

// TotalCost returns total CTC cost of a network on
// a batch of samples.
//
// The maxGos argument specifies the maximum number
// of goroutines to run batches on simultaneously.
// If it is 0, GOMAXPROCS is used.
func TotalCost(f rnn.SeqFunc, s sgd.SampleSet, maxBatch, maxGos int) float64 {
	if maxGos == 0 {
		maxGos = runtime.GOMAXPROCS(0)
	}

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

func costForBatch(f rnn.SeqFunc, s sgd.SampleSet) float64 {
	inputVars := make([][]autofunc.Result, s.Len())
	for i := 0; i < s.Len(); i++ {
		sample := s.GetSample(i).(Sample)
		inputVars[i] = varsToResults(sequenceToVars(sample.Input))
	}

	outputs := f.BatchSeqs(inputVars)

	var sum float64
	for i, outSeq := range outputs.OutputSeqs() {
		seqVars := sequenceToVars(outSeq)
		label := s.GetSample(i).(Sample).Label
		sum += LogLikelihood(varsToResults(seqVars), label).Output()[0]
	}

	return -sum
}
