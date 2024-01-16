package jobber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v3"
)

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

type PipelineActionOutcome struct {
	Variables       *PipelineVariables
	OutputBuffer    *bytes.Buffer
	StderrBuffer    *bytes.Buffer
	CreatedResource *GenericK8sResource
	Error           error
}

type PipelineExecutionEnvironment struct {
	EnvironmentalVariables map[string]string
	flattedString          []string
}

func (e *PipelineExecutionEnvironment) ToFlattenedStrings() []string {
	if len(e.flattedString) == 0 && len(e.EnvironmentalVariables) > 0 {
		s := make([]string, 0, len(e.EnvironmentalVariables))
		for k, v := range e.EnvironmentalVariables {
			s = append(s, fmt.Sprintf("%s=%s", k, v))
		}

		e.flattedString = s
	}

	return e.flattedString
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

type ActionEventType int

const (
	TemplateExpanded ActionEventType = iota
	ResourceCreated
	JobCompleted
	PodMovedToRunningState
	ExecutionSuccessful
	ValuesTransformCompleted
	ActionCompletedSuccessfully
	AnErrorOccurred
)

type ActionEvent struct {
	Type                   ActionEventType
	Error                  error
	ExpandedTemplateBuffer *bytes.Buffer
	StdoutBuffer           *bytes.Buffer
	StderrBuffer           *bytes.Buffer
	AffectedResource       *GenericK8sResource
}

func (action *PipelineAction) Run(pipelineVariables *PipelineVariables, executionEnvironment *PipelineExecutionEnvironment, client *Client, eventChannel chan<- *ActionEvent) {
	switch action.Type {
	case TemplatedResource:
		action.runTemplatedResource(pipelineVariables, client, eventChannel)
	case Executable:
		action.runExecutable(pipelineVariables, executionEnvironment, eventChannel)
	case ValuesTransform:
		action.runValuesTransform(pipelineVariables, eventChannel)
	}
}

var yamlDocumentSplitPattern = regexp.MustCompile(`(?m)^---$`)
var emptyYamlDocumentMatch = regexp.MustCompile(`(?s)^\s*$`)

func (action *PipelineAction) runTemplatedResource(pipelineVariables *PipelineVariables, client *Client, eventChannel chan<- *ActionEvent) {
	tmpl, err := template.New(filepath.Base(action.ActionFullyQualifiedPath)).Funcs(sprig.FuncMap()).Funcs(JobberTemplateFunctions()).ParseFiles(action.ActionFullyQualifiedPath)
	if err != nil {
		eventChannel <- &ActionEvent{
			Type:  AnErrorOccurred,
			Error: fmt.Errorf("failed to read resource template (%s): %s", action.ActionFullyQualifiedPath, err),
		}
		return
	}

	templateBuffer := new(bytes.Buffer)

	if err = tmpl.Execute(templateBuffer, pipelineVariables); err != nil {
		eventChannel <- &ActionEvent{
			Type:  AnErrorOccurred,
			Error: fmt.Errorf("failed to expand resource template (%s): %s", action.ActionFullyQualifiedPath, err),
		}
		return
	}

	eventChannel <- &ActionEvent{
		Type:                   TemplateExpanded,
		ExpandedTemplateBuffer: templateBuffer,
	}

	yamlDocuments := yamlDocumentSplitPattern.Split(templateBuffer.String(), -1)

	yamlDocumentsThatAreNotEmpty := make([]string, 0, len(yamlDocuments))
	for _, yamlDocument := range yamlDocuments {
		if !emptyYamlDocumentMatch.MatchString(yamlDocument) {
			yamlDocumentsThatAreNotEmpty = append(yamlDocumentsThatAreNotEmpty, yamlDocument)
		}
	}

	for _, yamlDocumentString := range yamlDocumentsThatAreNotEmpty {
		decoder := yaml.NewDecoder(strings.NewReader(yamlDocumentString))
		decodedYaml := make(map[string]any)

		if err = decoder.Decode(decodedYaml); err != nil {
			eventChannel <- &ActionEvent{
				Type:  AnErrorOccurred,
				Error: fmt.Errorf("failed to decode yaml from template (%s): %s", action.ActionFullyQualifiedPath, err),
			}
			return
		}

		if len(decodedYaml) > 0 {
			resource, err := NewGenericK8sResourceFromUnstructuredMap(decodedYaml, client)
			if err != nil {
				eventChannel <- &ActionEvent{
					Type:  AnErrorOccurred,
					Error: fmt.Errorf("decoded yaml from template (%s) does not describe a Kubernetes resource: %s", action.ActionFullyQualifiedPath, err),
				}
				return
			}

			if resource.NamespaceName() == "" {
				resource.SetNamespace(pipelineVariables.Runtime.DefaultNamespace.Name)
			}

			if err := resource.Create(); err != nil {
				eventChannel <- &ActionEvent{
					Type:  AnErrorOccurred,
					Error: fmt.Errorf("failed to create resource: %s", err),
				}
				return
			}

			eventChannel <- &ActionEvent{
				Type:             ResourceCreated,
				AffectedResource: resource,
			}

			switch resource.GvkString() {
			case "v1/Pod":
				if err = resource.AsAPod().WaitForRunningState(60 * time.Second); err != nil {
					if err == ErrorTimeExceeded {
						err = fmt.Errorf("timed out waiting for Running state")
					}
					eventChannel <- &ActionEvent{
						Type:             AnErrorOccurred,
						Error:            err,
						AffectedResource: resource,
					}
					return
				}

				eventChannel <- &ActionEvent{
					Type:             PodMovedToRunningState,
					AffectedResource: resource,
				}
			case "batch/v1/Job":
				if err = resource.AsAJob().WaitForCompletion(); err != nil {
					eventChannel <- &ActionEvent{
						Type:             AnErrorOccurred,
						Error:            err,
						AffectedResource: resource,
					}
					return
				}
			}

			pipelineVariables.Runtime.Add(resource)
		}
	}

	eventChannel <- &ActionEvent{
		Type: ActionCompletedSuccessfully,
	}
}

func (action *PipelineAction) runExecutable(pipelineVariables *PipelineVariables, executionEnvironment *PipelineExecutionEnvironment, eventChannel chan<- *ActionEvent) {
	cmdStdout := new(bytes.Buffer)
	cmdStderr := new(bytes.Buffer)

	cmd := exec.Command(action.ActionFullyQualifiedPath)
	cmd.Stdout = cmdStdout
	cmd.Stderr = cmdStderr

	cmd.Env = executionEnvironment.ToFlattenedStrings()

	jsonBytes, err := json.Marshal(pipelineVariables)
	if err != nil {
		eventChannel <- &ActionEvent{
			Type:  AnErrorOccurred,
			Error: fmt.Errorf("failed to marshall variables to json: %s", err),
		}
		return
	}

	stdinWritePipe, err := cmd.StdinPipe()
	if err != nil {
		eventChannel <- &ActionEvent{
			Type:  AnErrorOccurred,
			Error: fmt.Errorf("could not connect stdin pipe: %s", err),
		}
		return
	}

	go func() {
		defer stdinWritePipe.Close()
		stdinWritePipe.Write(jsonBytes)
	}()

	if err := cmd.Run(); err != nil {
		eventChannel <- &ActionEvent{
			Type:         AnErrorOccurred,
			Error:        err,
			StdoutBuffer: cmdStdout,
			StderrBuffer: cmdStderr,
		}
		return
	}

	eventChannel <- &ActionEvent{
		Type:         ExecutionSuccessful,
		StdoutBuffer: cmdStdout,
		StderrBuffer: cmdStderr,
	}

	eventChannel <- &ActionEvent{
		Type: ActionCompletedSuccessfully,
	}
}

func (action *PipelineAction) runValuesTransform(pipelineVariables *PipelineVariables, eventChannel chan<- *ActionEvent) {
}

func (outcome *PipelineActionOutcome) WriteOutputToFile(filePath string, fileModeIfFileIsCreated os.FileMode) error {
	if outcome.OutputBuffer != nil {
		return writeReaderToFile(filePath, fileModeIfFileIsCreated, outcome.OutputBuffer)
	}

	return nil
}

func (outcome *PipelineActionOutcome) WriteErrorToFile(filePath string, fileModeIfFileIsCreated os.FileMode) error {
	if outcome.StderrBuffer != nil {
		writeReaderToFile(filePath, fileModeIfFileIsCreated, outcome.StderrBuffer)
	}

	return nil

}

func writeReaderToFile(filePath string, fileModeIfFileIsCreated os.FileMode, reader io.Reader) error {
	fileHandle, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileModeIfFileIsCreated)
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	if _, err := io.Copy(fileHandle, reader); err != nil {
		return err
	}

	return nil
}
