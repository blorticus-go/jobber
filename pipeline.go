package jobber

type Pipeline struct {
	actions           []*PipelineAction
	indexOfNextAction int
}

func NewPipelineFromStringDescriptors(pipelineDescriptors []string, pipelineActionBasePath string) (*Pipeline, error) {
	actions := make([]*PipelineAction, len(pipelineDescriptors))

	for descriptorIndex, descriptor := range pipelineDescriptors {
		action, err := PipelineActionFromStringDescriptor(descriptor, pipelineActionBasePath)
		if err != nil {
			return nil, err
		}
		actions[descriptorIndex] = action
	}

	return &Pipeline{
		actions:           actions,
		indexOfNextAction: 0,
	}, nil
}

func (pipeline *Pipeline) NextAction() *PipelineAction {
	if pipeline.indexOfNextAction >= len(pipeline.actions) {
		return nil
	}

	p := pipeline.actions[pipeline.indexOfNextAction]
	pipeline.indexOfNextAction++

	return p
}

func (pipeline *Pipeline) Restart() *PipelineAction {
	pipeline.indexOfNextAction = 0
	return pipeline.NextAction()
}
