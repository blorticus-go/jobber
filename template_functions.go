package jobber

import (
	"fmt"
)

type CustomerTemplateFunctions struct {
	fmap map[string]any
}

var templateFunctions *CustomerTemplateFunctions

func init() {
	templateFunctions = &CustomerTemplateFunctions{
		fmap: make(map[string]any),
	}

	templateFunctions.Add("pod_ip_string", TemplateFunc_PodIPString)
}

func (t *CustomerTemplateFunctions) Add(functionName string, function any) *CustomerTemplateFunctions {
	t.fmap[functionName] = function
	return t
}

func JobberTemplateFunctions() map[string]any {
	return templateFunctions.fmap
}

func TemplateFunc_PodIPString(resource *TransitivePod) (string, error) {
	if resource == nil {
		return "", fmt.Errorf("pod not found")
	}

	return resource.IpString()
}
