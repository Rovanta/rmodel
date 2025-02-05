package processor

type Processor interface {
	Process(ctx BrainContext) error
	Clone() Processor
}

func NewFuncProcessor(processFn func(ctx BrainContext) error) *FuncProcessor {
	return &FuncProcessor{
		processFn: processFn,
	}
}

type FuncProcessor struct {
	processFn func(ctx BrainContext) error
}

func (p *FuncProcessor) Process(ctx BrainContext) error {
	// TODO wrap process func, wrap timeout, retry
	return p.processFn(ctx)
}

func (p *FuncProcessor) Clone() Processor {
	return &FuncProcessor{
		processFn: p.processFn,
	}
}

type EmptyProcessor struct{}

func (p *EmptyProcessor) Process(ctx BrainContext) error {
	return nil
}

func (p *EmptyProcessor) Clone() Processor {
	return &EmptyProcessor{}
}
