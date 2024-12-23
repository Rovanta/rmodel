package core

const (
	// BrainStateShutdown brain
	BrainStateShutdown BrainState = "Shutdown"
	BrainStateSleeping BrainState = "Sleeping"
	BrainStateRunning BrainState = "Running"
)

type BrainState string

type Brain interface {
	TrigLinks(links ...Link) error
	Entry() error
	EntryWithMemory(keysAndValues ...any) error

	// SetMemory set memories for brain, one key value pair is one memory.
	// memory will lazy initial util `SetMemory` or any link trig
	SetMemory(keysAndValues ...any) error
	// GetMemory get memory by key
	GetMemory(key any) any
	// ExistMemory indicates whether there is a memory in the brain
	ExistMemory(key any) bool
	// DeleteMemory delete one memory by key
	DeleteMemory(key any)
	// ClearMemory clear all memories
	ClearMemory()
	// GetState get brain state
	GetState() BrainState
	// Wait wait util brain maintainer shutdown, which means brain state is `Sleeping`
	Wait()
	// Shutdown the brain
	Shutdown()
}
