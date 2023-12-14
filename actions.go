package jobber

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

type PipelineCreatedResource struct {
	Information *K8sResourceInformation
	ApiObject   *unstructured.Unstructured
}

type PipelineActionOutcome struct {
	Variables       *PipelineVariables
	OutputReader    io.Reader
	ErrorReader     io.Reader
	CreatedResource *PipelineCreatedResource
	Error           error
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

func (action *PipelineAction) Run(pipelineVariables *PipelineVariables, client *Client) []*PipelineActionOutcome {
	switch action.Type {
	case TemplatedResource:
		return action.runTemplatedResource(pipelineVariables, client)
	case Executable:
		return action.runExecutable(pipelineVariables)
	case ValuesTransform:
		return action.runValuesTransform(pipelineVariables)
	}

	return nil
}

var yamlDocumentSplitPattern = regexp.MustCompile(`(?m)^---$`)

func (action *PipelineAction) runTemplatedResource(pipelineVariables *PipelineVariables, client *Client) []*PipelineActionOutcome {
	tmpl, err := template.New(filepath.Base(action.ActionFullyQualifiedPath)).Funcs(sprig.FuncMap()).Funcs(JobberTemplateFunctions()).ParseFiles(action.ActionFullyQualifiedPath)
	if err != nil {
		return []*PipelineActionOutcome{
			{
				Variables: pipelineVariables,
				Error:     NewTemplateError(action.Descriptor, "failed to parse resource template at (%s): %s", action.ActionFullyQualifiedPath, err),
			},
		}
	}

	templateBuffer := new(bytes.Buffer)

	if err = tmpl.Execute(templateBuffer, pipelineVariables); err != nil {
		return []*PipelineActionOutcome{
			{
				Variables:    pipelineVariables,
				Error:        NewTemplateError(action.Descriptor, "failed to expand resource template: %s", err),
				OutputReader: templateBuffer,
			},
		}
	}

	yamlDocuments := yamlDocumentSplitPattern.Split(templateBuffer.String(), -1)

	// if the template starts with the yaml separator, remove the first element which will be the empty string
	if len(yamlDocuments) > 1 && yamlDocuments[0] == "" {
		yamlDocuments = yamlDocuments[1:]
	}

	outcomes := make([]*PipelineActionOutcome, 0, len(yamlDocuments))

	for _, yamlDocumentString := range yamlDocuments {
		outcome := &PipelineActionOutcome{
			Variables:    pipelineVariables,
			OutputReader: strings.NewReader(yamlDocumentString),
		}

		decoder := yaml.NewDecoder(strings.NewReader(yamlDocumentString))
		decodedYaml := make(map[string]any)

		if err = decoder.Decode(decodedYaml); err != nil {
			outcome.Error = NewTemplateError(action.Descriptor, "unable to decode expanded template as yaml: %s", err)
			return append(outcomes, outcome)
		}

		if len(decodedYaml) > 0 {
			unstructured, err := client.MapToUnstructured(decodedYaml)
			if err != nil {
				outcome.Error = NewTemplateError(action.Descriptor, "expanded template yaml is not valid resource definition: %s", err)
				return append(outcomes, outcome)
			}

			if unstructured.GetNamespace() == "" {
				unstructured.SetNamespace(pipelineVariables.Config.Namespaces["Default"].GeneratedName)
			}

			created, err := client.CreateResourceFromUnstructured(unstructured)
			if err != nil {
				outcome.Error = NewResourceCreationError(
					action.Descriptor,
					&K8sResourceInformation{
						Kind:          unstructured.GetKind(),
						Name:          unstructured.GetName(),
						NamespaceName: unstructured.GetNamespace(),
					},
					err.Error())
				return append(outcomes, outcome)
			}

			// created and unstructured are subtley different.  If unstructured used GenerateName in the metadata section, Name() will
			// usually return the empty string (assuming it was not also set), but in this case, created.Name() would return the fully
			// generated name.
			outcome.CreatedResource = &PipelineCreatedResource{
				Information: &K8sResourceInformation{
					Kind:          created.GetKind(),
					Name:          created.GetName(),
					NamespaceName: created.GetNamespace(),
				},
				ApiObject: created,
			}

			pipelineVariables.Runtime.Add(created)

			outcomes = append(outcomes, outcome)
		}
	}

	return outcomes
}

func (action *PipelineAction) runExecutable(pipelineVariables *PipelineVariables) []*PipelineActionOutcome {
	return nil
}

func (action *PipelineAction) runValuesTransform(pipelineVariables *PipelineVariables) []*PipelineActionOutcome {
	return nil
}

func (outcome *PipelineActionOutcome) WriteOutputToFile(filePath string, fileModeIfFileIsCreated os.FileMode) error {
	if outcome.OutputReader != nil {
		return writeReaderToFile(filePath, fileModeIfFileIsCreated, outcome.OutputReader)
	}

	return nil
}

func (outcome *PipelineActionOutcome) WriteErrorToFile(filePath string, fileModeIfFileIsCreated os.FileMode) error {
	if outcome.ErrorReader != nil {
		writeReaderToFile(filePath, fileModeIfFileIsCreated, outcome.ErrorReader)
	}

	return nil

}

func writeReaderToFile(filePath string, fileModeIfFileIsCreated os.FileMode, reader io.Reader) error {
	fileHandle, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC, fileModeIfFileIsCreated)
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	if _, err := io.Copy(fileHandle, reader); err != nil {
		return err
	}

	return nil
}
