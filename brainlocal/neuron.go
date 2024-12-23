package brainlocal

import (
	"github.com/rModel/rModel/core"
	"github.com/rModel/rModel/internal/utils"
	"github.com/rModel/rModel/processor"
)

type neuron struct {
	id     string
	labels map[string]string
	spec   neuronSpec
	status neuronStatus
}

type neuronSpec struct {
	processor processor.Processor
	triggerGroups map[string][]*link
	castGroups map[string][]*link
	selector processor.Selector
}

type neuronStatus struct {
	state core.NeuronState
	count struct {
		process int
		succeed int
		failed  int
	}
}

func newNeuron(n core.Neuron, linkMap map[string]*link) *neuron {
	neu := &neuron{
		id:     n.GetID(),
		labels: utils.LabelsDeepCopy(n.GetLabels()),
		spec: neuronSpec{
			processor:     n.GetProcessor(),
			selector:      n.GetSelector(),
			triggerGroups: make(map[string][]*link),
			castGroups:    make(map[string][]*link),
		},
		status: neuronStatus{
			state: core.NeuronStateInactive,
		},
	}

	for gName, links := range n.ListTriggerGroups() {
		neu.spec.triggerGroups[gName] = make([]*link, len(links))
		for i, linkID := range links {
			neu.spec.triggerGroups[gName][i] = linkMap[linkID]
		}
	}

	for gName, links := range n.ListCastGroups() {
		neu.spec.castGroups[gName] = make([]*link, len(links))
		for i, linkID := range links {
			neu.spec.castGroups[gName][i] = linkMap[linkID]
		}
	}

	return neu
}
