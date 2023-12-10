package jobber

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

type Pipeline struct {
	actions           []*PipelineAction
	indexOfNextAction int
}

type PipelineActionType int

const (
	TemplatedResource PipelineActionType = iota
	ValuesTransform
	Executable
)

type PipelineAction struct {
	Type                     PipelineActionType
	Descriptor               string
	ActionFullyQualifiedPath string
}

func PipelineActionFromStringDescriptor(descriptor string, pipelineActionBasePath string) (*PipelineAction, error) {
	if descriptor == "" {
		return nil, fmt.Errorf("a pipeline action cannot be the empty string")
	}

	pr := []rune(pipelineActionBasePath)
	if pr[len(pr)-1] == '/' {
		pipelineActionBasePath = string(pr[:len(pr)-1])
	}

	pathElements := strings.Split(descriptor, "/")
	if len(pathElements) < 2 {
		return nil, fmt.Errorf("pipeline action descriptor (%s) contains only the type", descriptor)
	}

	actionFullyQualifiedPath, err := filepath.Abs(fmt.Sprintf("%s/%s", pipelineActionBasePath, descriptor))
	if err != nil {
		return nil, fmt.Errorf("cannot expand descriptor (%s) with base path (%s)", descriptor, pipelineActionBasePath)
	}

	switch pathElements[0] {
	case "resources":
		return &PipelineAction{
			Type:                     TemplatedResource,
			Descriptor:               descriptor,
			ActionFullyQualifiedPath: actionFullyQualifiedPath,
		}, nil
	case "values-transforms":
		return &PipelineAction{
			Type:                     ValuesTransform,
			Descriptor:               descriptor,
			ActionFullyQualifiedPath: actionFullyQualifiedPath,
		}, nil
	case "executables":
		return &PipelineAction{
			Type:                     Executable,
			Descriptor:               descriptor,
			ActionFullyQualifiedPath: actionFullyQualifiedPath,
		}, nil
	default:
		return nil, fmt.Errorf("pipeline type (%s) is not valid", pathElements[0])
	}
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

func (pipeline *Pipeline) Reset() {
	pipeline.indexOfNextAction = 0
}

func (action *PipelineAction) Run() error {
	switch action.Type {
	case TemplatedResource:
		return action.runTemplateResource()
	case Executable:
		return action.runExecutable()
	case ValuesTransform:
		return action.runValuesTransform()
	}

	return nil
}

func (action *PipelineAction) runTemplateResource() error {
	tmpl, err := template.ParseFiles(action.ActionFullyQualifiedPath)
	if err != nil {
		return fmt.Errorf("failed to read resource template (%s) at (%s): %s", action.Descriptor, action.ActionFullyQualifiedPath, err)
	}

	templateBuffer := new(bytes.Buffer)

	if err = tmpl.Execute(templateBuffer, nil); err != nil {
		return fmt.Errorf("failed to expand resource template (%s): %s", action.Descriptor, err)
	}

	return nil
}

func (action *PipelineAction) runExecutable() error {
	return nil
}

func (action *PipelineAction) runValuesTransform() error {
	return nil
}
