package jobber

import "fmt"

type TemplateError struct {
	errorText    string
	TemplateName string
}

type ResourceCreationError struct {
	errorText           string
	ResourceInformation *K8sResourceInformation
	TemplateName        string
}

type FileOrDirectoryCreationError struct {
	Path      string
	errorText string
}

func NewTemplateError(templateName string, errorStringFormat string, a ...any) *TemplateError {
	return &TemplateError{
		TemplateName: templateName,
		errorText:    fmt.Sprintf(errorStringFormat, a...),
	}
}

func NewResourceCreationError(fromTemplateNamed string, resourceInformation *K8sResourceInformation, errorStringFormat string, a ...any) *ResourceCreationError {
	return &ResourceCreationError{
		ResourceInformation: resourceInformation,
		errorText:           fmt.Sprintf(errorStringFormat, a...),
	}
}

func NewFileOrDirectoryCreationError(path string, errorStringFormat string, a ...any) *FileOrDirectoryCreationError {
	return &FileOrDirectoryCreationError{
		Path:      path,
		errorText: fmt.Sprintf(errorStringFormat, a...),
	}
}

func (e *TemplateError) Error() string {
	return e.errorText
}

func (e *ResourceCreationError) Error() string {
	return e.errorText
}

func (e *FileOrDirectoryCreationError) Error() string {
	return e.errorText
}
