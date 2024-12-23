package brainlite

import (
	"fmt"
	"time"

	"github.com/Rovanta/rmodel/core"
	"github.com/Rovanta/rmodel/internal/errors"
	"github.com/Rovanta/rmodel/processor"
)

func (b *BrainLite) ensureMaintainerStart() {
	if b.getState() == core.BrainStateShutdown {
		b.maintainerStart()
		b.setState(core.BrainStateSleeping)
	}

	return
}

func (b *BrainLite) maintainerStart() {
	b.logger.Info().
		Int("neuronWorkerNum", b.nWorkerNum).
		Int("neuronQueueLen", b.nQueueLen).
		Msg("brain maintainer start")
	if b.getState() != core.BrainStateShutdown {
		b.Shutdown()
	}

	// new
	b.nQueue = make(chan string, b.nQueueLen)
	b.bQueue = make(chan maintainEvent, bQueueLen)

	for i := 0; i < b.nWorkerNum; i++ {
		go b.runNeuronWorker()
	}
	go b.runBrainMaintainer()

}

func (b *BrainLite) runBrainMaintainer() {
	for msg := range b.bQueue {
		b.maintain(msg)
	}
}

func (b *BrainLite) maintain(event maintainEvent) {
	b.logger.Debug().Interface("event", event).Msg("got a maintain event")

	switch event.kind {
	case eventKindLink:
		if err := b.handleLinkEvent(event.action, event.id); err != nil {
			b.logger.Error().Err(err).Msg("handle link event error")
			return
		}
	case eventKindNeuron:
		if err := b.handleNeuronEvent(event.action, event.id); err != nil {
			b.logger.Error().Err(err).Msg("handle neuron event error")
			return
		}
	case eventKindBrain:
		if err := b.handleBrainEvent(event.action); err != nil {
			b.logger.Error().Err(err).Msg("handle brain event error")
			return
		}
	default:
		b.logger.Error().Msg("unknown maintain event kind")
		return
	}

	if event.kind != eventKindBrain {
		b.refreshState()
	}

	return
}

func (b *BrainLite) handleLinkEvent(action eventAction, linkID string) error {
	l, ok := b.links[linkID]
	if !ok {
		return errors.ErrLinkNotFound(linkID)
	}

	switch action {
	case eventActionLinkInit:
		// do nothing
	case eventActionLinkWait:
		// do nothing
	case eventActionLinkReady:
		dest, ok := b.neurons[l.spec.to]
		if !ok {
			b.logger.Debug().Any("link", b.neurons)
			return errors.ErrNeuronNotFound(l.spec.to)
		}

		// try dest neuron activate
		b.publishEvent(maintainEvent{
			kind:   eventKindNeuron,
			action: eventActionNeuronTryActivate,
			id:     dest.id,
		})

	default:
		return fmt.Errorf("unsupported link action: %s", action)
	}

	return nil
}

func (b *BrainLite) handleNeuronEvent(action eventAction, neuronID string) error {
	n, ok := b.neurons[neuronID]
	if !ok {
		return errors.ErrNeuronNotFound(neuronID)
	}

	switch action {
	case eventActionNeuronTryInactive:
		// do nothing for now, wait current neuron done and inactive
		// TODO cancel neuron process
	case eventActionNeuronTryActivate:
		return b.tryActivateNeuron(n)
	case eventActionNeuronTryCast:
		return b.neuronCast(n, false)
	case eventActionNeuronCastAnyway:
		return b.neuronCast(n, true)
	default:
		return fmt.Errorf("unsupported neuron action: %s", action)
	}

	return nil
}

func (b *BrainLite) handleBrainEvent(action eventAction) error {
	switch action {
	case eventActionBrainSleep:
		b.ForceSleep()
		return nil
	case eventActionBrainShutdown:
		b.Shutdown()
		return nil
	default:
		return fmt.Errorf("unsupported brain action: %s", action)
	}
}

func (b *BrainLite) tryActivateNeuron(n *neuron) error {
	if n.status.state == core.NeuronStateActivated {
		b.logger.Debug().Str("neuronID", n.id).Msg("neuron already activated")
		return nil
	}

	should := b.ifNeuronShouldActivate(n)
	if !should {
		b.logger.Debug().Str("neuronID", n.id).Msg("neuron should not be activated")
		return nil
	}

	// should END, send brain sleep message
	if n.id == core.EndNeuronID {
		b.logger.Info().Msg("arrival at END neuron")
		b.publishEvent(maintainEvent{
			kind:   eventKindBrain,
			action: eventActionBrainSleep,
		})
		return nil
	}

	b.publishEventActivateNeuron(n.id)

	return nil
}

func (b *BrainLite) neuronCast(n *neuron, isCastAnyway bool) error {
	if !isCastAnyway && n.status.state != core.NeuronStateInactive {
		b.logger.Debug().
			Str("neuronID", n.id).
			Msg("neuron already active, should not cast")
		return nil
	}

	b.logger.Debug().
		Str("neuronID", n.id).
		Msg("neuron try to cast")

	var selectedGroup string
	if n.spec.selector != nil {
		selectedGroup = n.spec.selector.Select(&brainContext{
			b:               b,
			currentNeuronID: n.id,
		})
	} else {
		selectedGroup = processor.DefaultCastGroupName
	}

	selectedLinks := make(map[string]struct{})

	for _, l := range n.spec.castGroups[selectedGroup] {
		selectedLinks[l.id] = struct{}{}

		switch l.status.state {
		case core.LinkStateWait:
			l.status.state = core.LinkStateReady
			b.publishEvent(maintainEvent{
				kind:   eventKindLink,
				action: eventActionLinkReady,
				id:     l.id,
			})

		case core.LinkStateInit:
			if !isCastAnyway {
				b.logger.Debug().
					Str("neuronID", n.id).
					Str("link", l.id).
					Msg("link on init state, will not cast")
			} else {
				l.status.state = core.LinkStateReady
				b.publishEvent(maintainEvent{
					kind:   eventKindLink,
					action: eventActionLinkReady,
					id:     l.id,
				})
			}

		case core.LinkStateReady:
			if !isCastAnyway {
				b.logger.Debug().
					Str("neuronID", n.id).
					Str("link", l.id).
					Msg("link already cast, will not cast again")
			} else {
				// TODO neuron label
				go func() {
					time.Sleep(500 * time.Millisecond)
					b.publishEvent(maintainEvent{
						kind:   eventKindNeuron,
						action: eventActionNeuronCastAnyway,
						id:     n.id,
					})
				}()
			}
		}

	}

	for _, links := range n.spec.castGroups {
		for _, l := range links {
			_, found := selectedLinks[l.id]
			if found {
				continue
			}
			if !isCastAnyway {
				if l.status.state == core.LinkStateWait {
					l.status.state = core.LinkStateInit
				}
			} else {
				l.status.state = core.LinkStateWait
			}

		}
	}

	return nil
}

func (b *BrainLite) ifNeuronShouldActivate(neu *neuron) bool {
	state := b.getState()
	if state == core.BrainStateSleeping || state == core.BrainStateShutdown {
		return false
	}

	for _, links := range neu.spec.triggerGroups {
		trigLinks := make([]*link, 0)
		for _, l := range links {
			if l.status.state == core.LinkStateReady {
				trigLinks = append(trigLinks, l)
			} else {
				break
			}
		}
		if len(links) != 0 && len(trigLinks) == len(links) {
			return true
		}
	}

	return false
}

func (b *BrainLite) refreshState() {
	inactiveCnt, activateCnt := b.getNeuronCountByState()
	initCnt, waitCnt, readyCnt := b.getLinkCountByState()

	b.logger.Debug().
		Int("neuronInactive", inactiveCnt).
		Int("neuronActivated", activateCnt).
		Int("linkInit", initCnt).
		Int("linkWait", waitCnt).
		Int("linkReady", readyCnt).
		Msg("refresh brain state by count")
	// send brain sleep message
	if activateCnt+waitCnt+readyCnt == 0 {
		b.publishEvent(maintainEvent{
			kind:   eventKindBrain,
			action: eventActionBrainSleep,
			id:     b.id,
		})
	} else { // > 0, set to running
		b.setState(core.BrainStateRunning)
	}
}

func (b *BrainLite) getNeuronCountByState() (int, int) {
	var inactiveCnt, activateCnt int
	for _, neu := range b.neurons {
		switch neu.status.state {
		case core.NeuronStateInactive:
			inactiveCnt++
		case core.NeuronStateActivated:
			activateCnt++
		}
	}

	return inactiveCnt, activateCnt
}

func (b *BrainLite) getLinkCountByState() (int, int, int) {
	var initCnt, waitCnt, readyCnt int
	for _, l := range b.links {
		switch l.status.state {
		case core.LinkStateInit:
			initCnt++
		case core.LinkStateWait:
			waitCnt++
		case core.LinkStateReady:
			readyCnt++
		}
	}

	return initCnt, waitCnt, readyCnt
}

func (b *BrainLite) ForceSleep() {
	for _, l := range b.links {
		l.status.state = core.LinkStateInit
	}
	for _, neu := range b.neurons {
		neu.status.state = core.NeuronStateInactive
	}
	b.setState(core.BrainStateSleeping)
}

func (b *BrainLite) setState(state core.BrainState) {
	b.mu.Lock()
	b.state = state
	b.cond.Broadcast() // Notify all waiting goroutines
	b.mu.Unlock()
}

func (b *BrainLite) getState() core.BrainState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
