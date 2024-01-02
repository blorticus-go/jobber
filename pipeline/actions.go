package pipeline

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/blorticus-go/jobber/api"
	"github.com/blorticus-go/jobber/wrapped"
	"gopkg.in/yaml.v2"
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

type ActionType int

const (
	InvalidActionType ActionType = iota
	TemplatedResource
	ValuesTransform
	Executable
)

type ActionOutcomeExecution struct {
	StdoutBuffer *bytes.Buffer
	StderrBuffer *bytes.Buffer
}

type ActionOutcomeTemplateExpansion struct {
	ExpandedTemplateBuffer *bytes.Buffer
}

type ActionOutcomeResourceCreation struct {
	Resource wrapped.Resource
}

type ActionOutcome struct {
	Execution         *ActionOutcomeExecution
	TemplateExpansion *ActionOutcomeTemplateExpansion
	ResourceCreation  *ActionOutcomeResourceCreation
	Error             error
}

type ActionMessage struct {
	Type    ActionMessageType
	Outcome *ActionOutcome
}

func descriptorStringToActionType(actionTypeAsAString string) (ActionType, error) {
	switch actionTypeAsAString {
	case "templates":
		return TemplatedResource, nil
	case "transforms":
		return ValuesTransform, nil
	case "executables":
		return Executable, nil
	}

	return InvalidActionType, fmt.Errorf("action (%s) does not translate to a known type", actionTypeAsAString)
}

var actionDescriptorElementsMatcher = regexp.MustCompile(`^/?([^/])+/(.+[^/])$`)

const (
	actionDescriptorMatchGroupWholeMatch = 0
	actionDescriptorMatchGroupActionType = 1
	actionDescriptorMatchGroupActionPath = 2
)

type Action interface {
	Type() ActionType
	Run(variables *Variables, messages chan<- *ActionMessage)
}

type ActionFactory struct {
	templateExpansionFuncMap template.FuncMap
	apiClient                *api.Client
}

func NewActionFactory(apiClient *api.Client, templateExpansionFuncMap template.FuncMap) *ActionFactory {
	if templateExpansionFuncMap == nil {
		templateExpansionFuncMap = make(template.FuncMap)
	}

	return &ActionFactory{
		templateExpansionFuncMap: templateExpansionFuncMap,
		apiClient:                apiClient,
	}
}

func actionDescriptorStringToNormalizedSet(descriptor string) (actionType ActionType, targetPathElements []string, err error) {
	elements := actionDescriptorElementsMatcher.FindStringSubmatch(descriptor)
	if len(elements) == 0 {
		return InvalidActionType, []string{}, fmt.Errorf("invalid format for action descriptor (%s)", descriptor)
	}

	at, err := descriptorStringToActionType(elements[actionDescriptorMatchGroupActionType])
	if err != nil {
		return InvalidActionType, []string{}, err
	}

	return at, strings.Split(elements[actionDescriptorMatchGroupActionPath], "/"), nil
}

func (factory *ActionFactory) NewActionFromStringDescriptor(descriptor string, pipelineActionBasePath string) (Action, error) {
	if descriptor == "" {
		return nil, fmt.Errorf("a pipeline action cannot be the empty string")
	}

	actionType, _, err := actionDescriptorStringToNormalizedSet(descriptor)
	if err != nil {
		return nil, err
	}

	actionFullyQualifiedPath, err := filepath.Abs(fmt.Sprintf("%s/%s", pipelineActionBasePath, strings.TrimLeft(descriptor, "/")))
	if err != nil {
		return nil, fmt.Errorf("cannot expand descriptor (%s) with base path (%s): %s", descriptor, pipelineActionBasePath, err)
	}

	switch actionType {
	case TemplatedResource:
		return &TemplatedResourceAction{
			pathToResourceTemplate: actionFullyQualifiedPath,
			expansionFuncMap:       factory.templateExpansionFuncMap,
		}, nil
	case ValuesTransform:
		return &ValuesTransformAction{
			pathToTransformExecutable: actionFullyQualifiedPath,
		}, nil

	case Executable:
		return &ExecutablesAction{
			pathToExecutable: actionFullyQualifiedPath,
		}, nil
	}

	return nil, nil
}

type TemplatedResourceAction struct {
	pathToResourceTemplate string
	expansionFuncMap       template.FuncMap
	apiClient              *api.Client
}

func (action *TemplatedResourceAction) Type() ActionType {
	return TemplatedResource
}

var yamlDocumentSplitPattern = regexp.MustCompile(`(?m)^---$`)
var emptyYamlDocumentMatch = regexp.MustCompile(`(?s)^\s*$`)

func (action *TemplatedResourceAction) Run(pipelineVariables *Variables, messages chan<- *ActionMessage) {
	defer func() {
		messages <- &ActionMessage{
			Type: ActionCompleted,
		}
	}()

	tmpl, err := template.New(filepath.Base(action.pathToResourceTemplate)).Funcs(action.expansionFuncMap).ParseFiles(action.pathToResourceTemplate)
	if err != nil {
		messages <- &ActionMessage{
			Type: TemplateExpansionFailed,
			Outcome: &ActionOutcome{
				Error: fmt.Errorf("failed to parse resource template at (%s): %s", action.pathToResourceTemplate, err),
			},
		}
		return
	}

	templateBuffer := new(bytes.Buffer)

	if err = tmpl.Execute(templateBuffer, pipelineVariables); err != nil {
		messages <- &ActionMessage{
			Type: TemplateExpansionFailed,
			Outcome: &ActionOutcome{
				Error: fmt.Errorf("failed to expand resource template: %s", err),
			},
		}
		return
	}

	messages <- &ActionMessage{
		Type: TemplateExpandedSuccessfully,
		Outcome: &ActionOutcome{
			TemplateExpansion: &ActionOutcomeTemplateExpansion{
				ExpandedTemplateBuffer: templateBuffer,
			},
		},
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
			messages <- &ActionMessage{
				Type: ResourceYamlParseFailed,
				Outcome: &ActionOutcome{
					Error: fmt.Errorf("failed to decode resource yaml: %s", err),
				},
			}
			return
		}

		if len(decodedYaml) == 0 {
			continue
		}

		resource := wrapped.NewResourceFromMap(decodedYaml, action.apiClient)

		if resource.NamespaceName() == "" {
			resource.SetNamespaceWithoutCommit(pipelineVariables.Runtime.DefaultNamespace.Name)
		}

		if err := resource.Create(); err != nil {
			messages <- &ActionMessage{
				Type: ResourceCreationFailed,
				Outcome: &ActionOutcome{
					ResourceCreation: &ActionOutcomeResourceCreation{
						Resource: resource,
					},
					Error: fmt.Errorf("failed to create resource: %s", err),
				},
			}
			return
		}

		switch resource.GroupVersionKindAsAString() {
		case "/v1/Pod":
			pod, err := wrapped.NewPodFromGeneric(resource)
			if err != nil {
				messages <- &ActionMessage{
					Type: ResourceCreationFailed,
					Outcome: &ActionOutcome{
						ResourceCreation: &ActionOutcomeResourceCreation{
							Resource: resource,
						},
						Error: fmt.Errorf("failed to coerce Generic resource to Pod resource: %s", err),
					},
				}
				return
			}

			messages <- &ActionMessage{
				Type: WaitingForPodRunningState,
				Outcome: &ActionOutcome{
					ResourceCreation: &ActionOutcomeResourceCreation{
						Resource: resource,
					},
				},
			}

			if err := pod.WaitForRunningState(60 * time.Second); err != nil {
				if err == wrapped.ErrorTimeExceeded {
					err = fmt.Errorf("timed out waiting for Running state")
				}
				messages <- &ActionMessage{
					Type: ResourceCreationFailed,
					Outcome: &ActionOutcome{
						ResourceCreation: &ActionOutcomeResourceCreation{
							Resource: resource,
						},
						Error: err,
					},
				}
				return
			}

		case "batch/v1/Job":
			job, err := wrapped.NewJobFromGeneric(resource)
			if err != nil {
				messages <- &ActionMessage{
					Type: ResourceCreationFailed,
					Outcome: &ActionOutcome{
						ResourceCreation: &ActionOutcomeResourceCreation{
							Resource: resource,
						},
						Error: fmt.Errorf("failed to coerce Generic resource to Job resource: %s", err),
					},
				}
				return
			}

			messages <- &ActionMessage{
				Type: WaitingForJobCompletion,
				Outcome: &ActionOutcome{
					ResourceCreation: &ActionOutcomeResourceCreation{
						Resource: job,
					},
				},
			}

			if err := job.WaitForJobCompletion(); err != nil {
				messages <- &ActionMessage{
					Type: ResourceCreationFailed,
					Outcome: &ActionOutcome{
						ResourceCreation: &ActionOutcomeResourceCreation{
							Resource: job,
						},
						Error: err,
					},
				}
				return
			}
		}

		messages <- &ActionMessage{
			Type: ResourceCreatedSuccessfully,
			Outcome: &ActionOutcome{
				ResourceCreation: &ActionOutcomeResourceCreation{
					Resource: resource,
				},
			},
		}
	}
}

type ValuesTransformAction struct {
	pathToTransformExecutable string
}

func NewValuesTransformAction(fullyQualifiedTransformExecutablePath string) *ValuesTransformAction {
	return &ValuesTransformAction{
		pathToTransformExecutable: fullyQualifiedTransformExecutablePath,
	}
}

func (action *ValuesTransformAction) Type() ActionType {
	return ValuesTransform
}

func (action *ValuesTransformAction) Run(variables *Variables, messages chan<- *ActionMessage) {
}

type ExecutablesAction struct {
	pathToExecutable string
}

func NewExecutablesActions(fullyQualifiedExecutablePath string) *ExecutablesAction {
	return &ExecutablesAction{
		pathToExecutable: fullyQualifiedExecutablePath,
	}
}

func (action *ExecutablesAction) Type() ActionType {
	return Executable
}

func (action *ExecutablesAction) Run(variables *Variables, messages chan<- *ActionMessage) {
}

// type PipelineActionType int

// const (
// 	TemplatedResource PipelineActionType = iota
// 	ValuesTransform
// 	Executable
// )

// type PipelineAction struct {
// 	Type                     PipelineActionType
// 	Descriptor               string
// 	ActionFullyQualifiedPath string
// }

// type PipelineActionOutcome struct {
// 	Variables       *PipelineVariables
// 	OutputBuffer    *bytes.Buffer
// 	StderrBuffer    *bytes.Buffer
// 	CreatedResource *GenericK8sResource
// 	Error           error
// }

// func PipelineActionFromStringDescriptor(descriptor string, pipelineActionBasePath string) (*PipelineAction, error) {
// 	if descriptor == "" {
// 		return nil, fmt.Errorf("a pipeline action cannot be the empty string")
// 	}

// 	pr := []rune(pipelineActionBasePath)
// 	if pr[len(pr)-1] == '/' {
// 		pipelineActionBasePath = string(pr[:len(pr)-1])
// 	}

// 	pathElements := strings.Split(descriptor, "/")
// 	if len(pathElements) < 2 {
// 		return nil, fmt.Errorf("pipeline action descriptor (%s) contains only the type", descriptor)
// 	}

// 	actionFullyQualifiedPath, err := filepath.Abs(fmt.Sprintf("%s/%s", pipelineActionBasePath, descriptor))
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot expand descriptor (%s) with base path (%s)", descriptor, pipelineActionBasePath)
// 	}

// 	switch pathElements[0] {
// 	case "resources":
// 		return &PipelineAction{
// 			Type:                     TemplatedResource,
// 			Descriptor:               descriptor,
// 			ActionFullyQualifiedPath: actionFullyQualifiedPath,
// 		}, nil
// 	case "values-transforms":
// 		return &PipelineAction{
// 			Type:                     ValuesTransform,
// 			Descriptor:               descriptor,
// 			ActionFullyQualifiedPath: actionFullyQualifiedPath,
// 		}, nil
// 	case "executables":
// 		return &PipelineAction{
// 			Type:                     Executable,
// 			Descriptor:               descriptor,
// 			ActionFullyQualifiedPath: actionFullyQualifiedPath,
// 		}, nil
// 	default:
// 		return nil, fmt.Errorf("pipeline type (%s) is not valid", pathElements[0])
// 	}
// }

// func (action *PipelineAction) Run(pipelineVariables *PipelineVariables, client *Client) []*PipelineActionOutcome {
// 	switch action.Type {
// 	case TemplatedResource:
// 		return action.runTemplatedResource(pipelineVariables, client)
// 	case Executable:
// 		return action.runExecutable(pipelineVariables)
// 	case ValuesTransform:
// 		return action.runValuesTransform(pipelineVariables)
// 	}

// 	return nil
// }

// var yamlDocumentSplitPattern = regexp.MustCompile(`(?m)^---$`)
// var emptyYamlDocumentMatch = regexp.MustCompile(`(?s)^\s*$`)

// func (action *PipelineAction) runTemplatedResource(pipelineVariables *PipelineVariables, client *Client) []*PipelineActionOutcome {
// 	tmpl, err := template.New(filepath.Base(action.ActionFullyQualifiedPath)).Funcs(sprig.FuncMap()).Funcs(JobberTemplateFunctions()).ParseFiles(action.ActionFullyQualifiedPath)
// 	if err != nil {
// 		return []*PipelineActionOutcome{
// 			{
// 				Variables: pipelineVariables,
// 				Error:     NewTemplateError(action.Descriptor, "failed to parse resource template at (%s): %s", action.ActionFullyQualifiedPath, err),
// 			},
// 		}
// 	}

// 	templateBuffer := new(bytes.Buffer)

// 	if err = tmpl.Execute(templateBuffer, pipelineVariables); err != nil {
// 		return []*PipelineActionOutcome{
// 			{
// 				Variables:    pipelineVariables,
// 				Error:        NewTemplateError(action.Descriptor, "failed to expand resource template: %s", err),
// 				OutputBuffer: templateBuffer,
// 			},
// 		}
// 	}

// 	yamlDocuments := yamlDocumentSplitPattern.Split(templateBuffer.String(), -1)

// 	yamlDocumentsThatAreNotEmpty := make([]string, 0, len(yamlDocuments))
// 	for _, yamlDocument := range yamlDocuments {
// 		if !emptyYamlDocumentMatch.MatchString(yamlDocument) {
// 			yamlDocumentsThatAreNotEmpty = append(yamlDocumentsThatAreNotEmpty, yamlDocument)
// 		}
// 	}

// 	outcomes := make([]*PipelineActionOutcome, 0, len(yamlDocuments))

// 	for _, yamlDocumentString := range yamlDocumentsThatAreNotEmpty {
// 		outcome := &PipelineActionOutcome{
// 			Variables:    pipelineVariables,
// 			OutputBuffer: bytes.NewBufferString(yamlDocumentString),
// 		}

// 		decoder := yaml.NewDecoder(strings.NewReader(yamlDocumentString))
// 		decodedYaml := make(map[string]any)

// 		if err = decoder.Decode(decodedYaml); err != nil {
// 			outcome.Error = NewTemplateError(action.Descriptor, "unable to decode expanded template as yaml: %s", err)
// 			return append(outcomes, outcome)
// 		}

// 		if len(decodedYaml) > 0 {
// 			resource, err := NewGenericK8sResourceFromUnstructuredMap(decodedYaml, client)
// 			if err != nil {
// 				outcome.Error = NewTemplateError(action.Descriptor, "expanded template yaml is not valid resource definition: %s", err)
// 				return append(outcomes, outcome)
// 			}

// 			if resource.NamespaceName() == "" {
// 				resource.SetNamespace(pipelineVariables.Config.DefaultNamespace.GeneratedName)
// 			}

// 			if err := resource.Create(); err != nil {
// 				outcome.Error = NewResourceCreationError(action.Descriptor, resource.Information(), err.Error())
// 				return append(outcomes, outcome)
// 			}

// 			switch resource.SimplifiedTypeString() {
// 			case "Pod":
// 				if err = resource.AsAPod().WaitForRunningState(60 * time.Second); err != nil {
// 					if err == ErrorTimeExceeded {
// 						err = fmt.Errorf("timed out waiting for Running state")
// 					}
// 					outcome.Error = NewResourceCreationError(action.Descriptor, resource.Information(), err.Error())
// 					return append(outcomes, outcome)
// 				}
// 			case "Job":
// 				if err = resource.AsAJob().WaitForCompletion(); err != nil {
// 					outcome.Error = NewJobCompletionFailureError(resource.Information(), err.Error())
// 					return append(outcomes, outcome)
// 				}
// 			}

// 			outcome.CreatedResource = resource

// 			pipelineVariables.Runtime.Add(resource)

// 			outcomes = append(outcomes, outcome)
// 		}
// 	}

// 	return outcomes
// }

// func (action *PipelineAction) runExecutable(pipelineVariables *PipelineVariables) []*PipelineActionOutcome {
// 	cmdStdout := new(bytes.Buffer)
// 	cmdStderr := new(bytes.Buffer)

// 	cmd := exec.Command(action.ActionFullyQualifiedPath)
// 	cmd.Stdout = cmdStdout
// 	cmd.Stderr = cmdStderr

// 	outcome := &PipelineActionOutcome{
// 		Variables:    pipelineVariables,
// 		OutputBuffer: cmdStdout,
// 		StderrBuffer: cmdStderr,
// 	}

// 	jsonBytes, err := json.Marshal(pipelineVariables)
// 	if err != nil {
// 		outcome.Error = fmt.Errorf("failed to marshall variables to json: %s", err)
// 		return []*PipelineActionOutcome{outcome}
// 	}

// 	stdinWritePipe, err := cmd.StdinPipe()
// 	if err != nil {
// 		outcome.Error = fmt.Errorf("could not connect stdin pipe: %s", err)
// 		return []*PipelineActionOutcome{outcome}
// 	}

// 	go func() {
// 		defer stdinWritePipe.Close()
// 		stdinWritePipe.Write(jsonBytes)
// 	}()

// 	if err := cmd.Run(); err != nil {
// 		outcome.Error = err
// 	}

// 	return []*PipelineActionOutcome{outcome}
// }

// func (action *PipelineAction) runValuesTransform(pipelineVariables *PipelineVariables) []*PipelineActionOutcome {
// 	return nil
// }

// func (outcome *PipelineActionOutcome) WriteOutputToFile(filePath string, fileModeIfFileIsCreated os.FileMode) error {
// 	if outcome.OutputBuffer != nil {
// 		return writeReaderToFile(filePath, fileModeIfFileIsCreated, outcome.OutputBuffer)
// 	}

// 	return nil
// }

// func (outcome *PipelineActionOutcome) WriteErrorToFile(filePath string, fileModeIfFileIsCreated os.FileMode) error {
// 	if outcome.StderrBuffer != nil {
// 		writeReaderToFile(filePath, fileModeIfFileIsCreated, outcome.StderrBuffer)
// 	}

// 	return nil

// }

// func writeReaderToFile(filePath string, fileModeIfFileIsCreated os.FileMode, reader io.Reader) error {
// 	fileHandle, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileModeIfFileIsCreated)
// 	if err != nil {
// 		return err
// 	}
// 	defer fileHandle.Close()

// 	if _, err := io.Copy(fileHandle, reader); err != nil {
// 		return err
// 	}

// 	return nil
// }
