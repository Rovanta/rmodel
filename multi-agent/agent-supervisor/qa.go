package main

import (
	"github.com/Rovanta/rmodel/community/processor/go_code_tester"
	"github.com/Rovanta/rmodel/processor"
)

const (
	memKeyGoTestResult = "go_test_result"
)

func QAProcess(b processor.BrainContext) error {
	p := go_code_tester.NewProcessor().WithTestCodeKeep(true)
	if err := p.Process(b); err != nil {
		return err
	}

	if err := b.SetMemory(memKeyFeedback, b.GetCurrentNeuronID()); err != nil {
		return err
	}

	return nil
}
