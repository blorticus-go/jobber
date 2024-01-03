package pipeline

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/blorticus-go/jobber/wrapped"
	"gopkg.in/yaml.v2"
)

type ActionType int

const (
	InvalidActionType ActionType = iota
	TemplatedResource
	ValuesTransform
	Executable
)

func ActionTypeToString(at ActionType) string {
	switch at {
	case InvalidActionType:
		return "InvalidActionType"
	case TemplatedResource:
		return "TempatedResource"
	case ValuesTransform:
		return "ValuesTransform"
	case Executable:
		return "Executable"
	}

	return ""
}

func descriptorStringToActionType(actionTypeAsAString string) (ActionType, error) {
	switch actionTypeAsAString {
	case "resources":
		return TemplatedResource, nil
	case "transforms":
		return ValuesTransform, nil
	case "executables":
		return Executable, nil
	}

	return InvalidActionType, fmt.Errorf("action (%s) does not translate to a known type", actionTypeAsAString)
}

var actionDescriptorElementsMatcher = regexp.MustCompile(`^/?([^/]+)/(.+[^/])$`)

const (
	actionDescriptorMatchGroupWholeMatch = 0
	actionDescriptorMatchGroupActionType = 1
	actionDescriptorMatchGroupActionPath = 2
)

type Action interface {
	Type() ActionType
	Run(variables *Variables, messages chan<- *ActionMessage)
}

type ActionMechanic interface {
	ExpandFileAsTemplate(filePath string, templateFuncMap template.FuncMap, pipelineVariables *Variables) (expandedBuffer *bytes.Buffer, err error)
	ProcessBytesBufferAsYamlDocuments(buff *bytes.Buffer) (nonEmptyYamlDocuments []map[string]any, err error)
	ConvertDecodedYamlToResource(decodedYaml map[string]any, defaultNamespaceName string) (wrapped.Resource, error)
	CreateResource(resource wrapped.Resource) error
	TreatResourceAsPodAndWaitForRunningState(r wrapped.Resource) error
	TreatResourceAsAJobAndWaitForCompletion(r wrapped.Resource) error
}

type ActionFactory struct {
	templateExpansionFuncMap template.FuncMap
	resourceFactory          wrapped.ResourceFactory
	actionMechanic           ActionMechanic
}

func NewActionFactory(resourceFactory wrapped.ResourceFactory, templateExpansionFuncMap template.FuncMap) *ActionFactory {
	if templateExpansionFuncMap == nil {
		templateExpansionFuncMap = make(template.FuncMap)
	}

	return &ActionFactory{
		templateExpansionFuncMap: templateExpansionFuncMap,
		resourceFactory:          resourceFactory,
		actionMechanic:           NewDefaultActionMechanic(resourceFactory),
	}
}

func (factory *ActionFactory) ReplaceActionMechanicWith(mechanic ActionMechanic) {
	factory.actionMechanic = mechanic
}

type DefaultActionMechanic struct {
	resourceFactory wrapped.ResourceFactory
}

func NewDefaultActionMechanic(resourceFactory wrapped.ResourceFactory) *DefaultActionMechanic {
	return &DefaultActionMechanic{
		resourceFactory: resourceFactory,
	}
}

func (m *DefaultActionMechanic) ExpandFileAsTemplate(filePath string, templateFuncMap template.FuncMap, pipelineVariables *Variables) (expandedBuffer *bytes.Buffer, err error) {
	tmpl, err := template.New(filepath.Base(filePath)).Funcs(templateFuncMap).ParseFiles(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource template at (%s): %s", filePath, err)
	}

	templateBuffer := new(bytes.Buffer)

	if err = tmpl.Execute(templateBuffer, pipelineVariables); err != nil {
		return nil, fmt.Errorf("failed to expand resource template at (%s): %s", filePath, err)
	}

	return templateBuffer, nil
}

func (m *DefaultActionMechanic) ProcessBytesBufferAsYamlDocuments(buff *bytes.Buffer) (nonEmptyYamlDocuments []map[string]any, err error) {
	yamlDocuments := yamlDocumentSplitPattern.Split(buff.String(), -1)

	yamlDocumentsThatAreNotEmpty := make([]string, 0, len(yamlDocuments))
	for _, yamlDocument := range yamlDocuments {
		if !emptyYamlDocumentMatch.MatchString(yamlDocument) {
			yamlDocumentsThatAreNotEmpty = append(yamlDocumentsThatAreNotEmpty, yamlDocument)
		}
	}

	decodedYamlDocuments := make([]map[string]any, 0, len(yamlDocumentsThatAreNotEmpty))
	for _, yamlDocumentString := range yamlDocumentsThatAreNotEmpty {
		decoder := yaml.NewDecoder(strings.NewReader(yamlDocumentString))
		decodedYaml := make(map[string]any)

		if err = decoder.Decode(decodedYaml); err != nil {
			return nil, fmt.Errorf("failed to decode resource yaml: %s", err)
		}

		if len(decodedYaml) > 0 {
			decodedYamlDocuments = append(decodedYamlDocuments, decodedYaml)
		}
	}

	return decodedYamlDocuments, nil
}

func (m *DefaultActionMechanic) ConvertDecodedYamlToResource(decodedYaml map[string]any, defaultNamespaceName string) (wrapped.Resource, error) {
	resource := m.resourceFactory.NewResourceFromMap(decodedYaml)
	if resource.NamespaceName() == "" {
		return m.resourceFactory.NewResourceForNamespaceFromMap(decodedYaml, defaultNamespaceName), nil
	}

	return resource, nil
}

func (m *DefaultActionMechanic) CreateResource(resource wrapped.Resource) error {
	if err := resource.Create(); err != nil {
		return fmt.Errorf("failed to create resource: %s", err)
	}

	return nil
}

func (m *DefaultActionMechanic) TreatResourceAsPodAndWaitForRunningState(resource wrapped.Resource) error {
	pod, err := m.resourceFactory.CoerceResourceToPod(resource)
	if err != nil {
		return fmt.Errorf("failed to coerce Generic resource to Pod resource: %s", err)
	}

	if err := pod.WaitForRunningState(60 * time.Second); err != nil {
		if err == wrapped.ErrorTimeExceeded {
			err = fmt.Errorf("timed out waiting for Running state")
		}
		return err
	}

	return nil
}

func (m *DefaultActionMechanic) TreatResourceAsAJobAndWaitForCompletion(resource wrapped.Resource) error {
	job, err := m.resourceFactory.CoerceResourceToJob(resource)
	if err != nil {
		return fmt.Errorf("failed to coerce Generic resource to Job resource: %s", err)
	}

	if err := job.WaitForCompletion(); err != nil {
		return err
	}

	return err
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
		return NewTemplatedResourceAction(actionFullyQualifiedPath, factory.templateExpansionFuncMap, factory.actionMechanic), nil
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
	mechanic               ActionMechanic
}

func NewTemplatedResourceAction(pathToResourceTemplate string, expansionFuncMap template.FuncMap, mechanic ActionMechanic) *TemplatedResourceAction {
	return &TemplatedResourceAction{
		pathToResourceTemplate: pathToResourceTemplate,
		expansionFuncMap:       expansionFuncMap,
		mechanic:               mechanic,
	}
}

func (action *TemplatedResourceAction) Type() ActionType {
	return TemplatedResource
}

var yamlDocumentSplitPattern = regexp.MustCompile(`(?m)^---$`)
var emptyYamlDocumentMatch = regexp.MustCompile(`(?s)^\s*$`)

func (action *TemplatedResourceAction) Run(pipelineVariables *Variables, messages chan<- *ActionMessage) {
	defer func() {
		messages <- NewActionCompletedMessage()
		close(messages)
	}()

	templateBuffer, err := action.mechanic.ExpandFileAsTemplate(action.pathToResourceTemplate, action.expansionFuncMap, pipelineVariables)
	if err != nil {
		messages <- NewTemplateExpansionFailed(err)
		return
	}

	messages <- NewTemplateExpandedSuccessfully(templateBuffer)

	decodedYamlDocuments, err := action.mechanic.ProcessBytesBufferAsYamlDocuments(templateBuffer)
	if err != nil {
		messages <- NewResourceYamlParseFailed(fmt.Errorf("here: %s", err))
		return
	}

	if len(decodedYamlDocuments) == 0 {
		messages <- NewResourceYamlParseFailed(fmt.Errorf("no non-empty yaml documents found"))
		return
	}

	for _, decodedYaml := range decodedYamlDocuments {
		resource, err := action.mechanic.ConvertDecodedYamlToResource(decodedYaml, pipelineVariables.Runtime.DefaultNamespace.Name)
		if err != nil {
			messages <- NewResourceCreationFailed(resource, err)
			return
		}

		if err := action.mechanic.CreateResource(resource); err != nil {
			messages <- NewResourceCreationFailed(resource, err)
			return
		}

		switch wrapped.GroupVersionKindAsAString(resource.GroupVersionKind()) {
		case "v1/Pod":
			messages <- NewWaitingForPodRunningState(resource)

			if err := action.mechanic.TreatResourceAsPodAndWaitForRunningState(resource); err != nil {
				messages <- NewResourceCreationFailed(resource, err)
				return
			}

		case "batch/v1/Job":
			messages <- NewWaitingForJobCompletion(resource)

			if err := action.mechanic.TreatResourceAsAJobAndWaitForCompletion(resource); err != nil {
				messages <- NewResourceCreationFailed(resource, err)
				return
			}
		}

		messages <- NewResourceCreatedSuccessfully(resource)
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
