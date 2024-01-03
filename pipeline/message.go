package pipeline

import (
	"bytes"

	"github.com/blorticus-go/jobber/wrapped"
)

type ActionMessageType int

const (
	ActionCompleted ActionMessageType = iota
	ResourceCreatedSuccessfully
	ResourceCreationFailed
	TemplateExpandedSuccessfully
	TemplateExpansionFailed
	ResourceYamlParseFailed
	ExecutionCompletedSuccessfully
	ExecutionFailed
	VariablesTransformCompletedSuccessfully
	VariablesTransformFailed
	WaitingForJobCompletion
	WaitingForPodRunningState
)

type ActionExecution struct {
	StdoutBuffer *bytes.Buffer
	StderrBuffer *bytes.Buffer
}

type ActionTemplateExpansion struct {
	ExpandedTemplateBuffer *bytes.Buffer
}

type ActionMessage struct {
	Type      ActionMessageType
	Resource  wrapped.Resource
	Template  *ActionTemplateExpansion
	Execution *ActionExecution
	Variables *Variables
	Error     error
}

func NewActionCompletedMessage() *ActionMessage {
	return &ActionMessage{
		Type: ActionCompleted,
	}
}

func NewResourceCreatedSuccessfully(resource wrapped.Resource) *ActionMessage {
	return &ActionMessage{
		Type:     ResourceCreatedSuccessfully,
		Resource: resource,
	}
}

func NewResourceCreationFailed(resource wrapped.Resource, err error) *ActionMessage {
	return &ActionMessage{
		Type:     ResourceCreationFailed,
		Resource: resource,
		Error:    err,
	}
}

func NewTemplateExpandedSuccessfully(templateBuffer *bytes.Buffer) *ActionMessage {
	return &ActionMessage{
		Type: ResourceCreationFailed,
		Template: &ActionTemplateExpansion{
			ExpandedTemplateBuffer: templateBuffer,
		},
	}
}

func NewTemplateExpansionFailed(err error) *ActionMessage {
	return &ActionMessage{
		Type:  ResourceCreationFailed,
		Error: err,
	}
}

func NewResourceYamlParseFailed(err error) *ActionMessage {
	return &ActionMessage{
		Type:  ResourceCreationFailed,
		Error: err,
	}
}

func NewExecutionCompletedSuccessfully(stdoutBuffer *bytes.Buffer, stderrBuffer *bytes.Buffer) *ActionMessage {
	return &ActionMessage{
		Type: ResourceCreationFailed,
		Execution: &ActionExecution{
			StdoutBuffer: stdoutBuffer,
			StderrBuffer: stderrBuffer,
		},
	}
}

func NewExecutionFailed(stdoutBuffer, stderrBuffer *bytes.Buffer, err error) *ActionMessage {
	return &ActionMessage{
		Type: ResourceCreationFailed,
		Execution: &ActionExecution{
			StdoutBuffer: stdoutBuffer,
			StderrBuffer: stderrBuffer,
		},
		Error: err,
	}
}

func NewVariablesTransformCompletedSuccessfully(outputVariables *Variables, stdoutBuffer, stderrBuffer *bytes.Buffer) *ActionMessage {
	return &ActionMessage{
		Type: ResourceCreationFailed,
		Execution: &ActionExecution{
			StdoutBuffer: stdoutBuffer,
			StderrBuffer: stderrBuffer,
		},
		Variables: outputVariables,
	}
}

func NewVariablesTransformFailed(stdoutBuffer, stderrBuffer *bytes.Buffer, err error) *ActionMessage {
	return &ActionMessage{
		Type: ResourceCreationFailed,
		Execution: &ActionExecution{
			StdoutBuffer: stdoutBuffer,
			StderrBuffer: stderrBuffer,
		},
		Error: err,
	}
}

func NewWaitingForJobCompletion(resource wrapped.Resource) *ActionMessage {
	return &ActionMessage{
		Type:     ResourceCreationFailed,
		Resource: resource,
	}
}

func NewWaitingForPodRunningState(resource wrapped.Resource) *ActionMessage {
	return &ActionMessage{
		Type:     ResourceCreationFailed,
		Resource: resource,
	}
}
