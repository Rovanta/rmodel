package brainlite

import (
	"fmt"

	"github.com/Rovanta/rmodel/core"
	"github.com/Rovanta/rmodel/internal/errors"
)

func (b *BrainLite) publishEventActivateNeuron(neuronID string) {
	if b.getState() == core.BrainStateShutdown || b.nQueue == nil {
		return
	}
	b.logger.Debug().Interface("neuronID", neuronID).Msg("publish activate neuron event")

	b.nQueue <- neuronID
}

func (b *BrainLite) runNeuronWorker() {
	for neuronID := range b.nQueue {
		neu, ok := b.neurons[neuronID]
		if !ok {
			b.logger.Error().Str("neuronID", neuronID).Msg("neuron not found")
			continue
		}

		err := b.activateNeuron(neu)
		if err != nil {
			b.logger.Error().Err(err).Str("neuronID", neuronID).Msg("activate neuron error")
		}
	}
}

func (b *BrainLite) activateNeuron(neu *neuron) error {
	if neu == nil {
		return errors.ErrNeuronNotFound("nil")
	}

	b.logger.Debug().Interface("neuronID", neu.id).Msg("start activate neuron")
	neu.status.state = core.NeuronStateActivated
	// in-link set init
	for _, links := range neu.spec.triggerGroups {
		for _, l := range links {
			l.status.state = core.LinkStateInit
		}
	}

	// out-link set wait
	for _, links := range neu.spec.castGroups {
		for _, l := range links {
			l.status.state = core.LinkStateWait
		}
	}

	neu.status.count.process++
	// block process
	err := neu.spec.processor.Process(&brainContext{
		b:               b,
		currentNeuronID: neu.id,
	})
	neu.status.state = core.NeuronStateInactive
	if err != nil {
		neu.status.count.failed++
		return fmt.Errorf("process neuron error: %w", err)
	}

	// SucceedCount++
	neu.status.count.succeed++

	// cast
	b.publishEvent(maintainEvent{
		kind:   eventKindNeuron,
		action: eventActionNeuronTryCast,
		id:     neu.id,
	})

	return nil
}
