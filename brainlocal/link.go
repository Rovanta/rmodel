package brainlocal

import (
	"github.com/rModel/rModel/core"
)

type link struct {
	id     string
	spec   linkSpec
	status linkStatus
}

type linkSpec struct {
	// from neuron ID
	from string
	// to neuron ID
	to string
}

type linkStatus struct {
	state core.LinkState
	count struct {
		process int
		succeed int
		failed int
	}
}

func newLink(l core.Link) *link {
	return &link{
		id: l.GetID(),
		spec: linkSpec{
			from: l.GetSrcNeuronID(),
			to:   l.GetDestNeuronID(),
		},
		status: linkStatus{
			state: core.LinkStateInit,
		},
	}
}

func (l *link) isEntryLink() bool {
	if l.spec.from == core.EntryLinkFrom {
		return true
	}

	return false
}
