package main

import (
	"fmt"

	"github.com/Rovanta/rmodel"
	"github.com/Rovanta/rmodel/brainlocal"
	"github.com/Rovanta/rmodel/processor"
)

func main() {
	bp := rModel.NewBlueprint()
	nested := bp.AddNeuron(nestedBrain)

	_, _ = bp.AddEntryLinkTo(nested)

	brain := brainlocal.BuildBrain(bp)
	_ = brain.Entry()
	brain.Wait()

	fmt.Printf("nested result: %s\n", brain.GetMemory("nested_result").(string))
}

func nestedBrain(outerBrain processor.BrainContext) error {
	bp := rModel.NewBlueprint()
	run := bp.AddNeuron(func(curBrain processor.BrainContext) error {
		_ = curBrain.SetMemory("result", fmt.Sprintf("run here neuron: %s.%s", outerBrain.GetCurrentNeuronID(), curBrain.GetCurrentNeuronID()))
		return nil
	})

	_, _ = bp.AddEntryLinkTo(run)

	brain := brainlocal.BuildBrain(bp)

	// run nested brain
	_ = brain.Entry()
	brain.Wait()
	// get nested brain result
	result := brain.GetMemory("result").(string)
	// pass nested brain result to outer brain
	_ = outerBrain.SetMemory("nested_result", result)

	return nil
}
