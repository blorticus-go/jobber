package pipeline

import (
	"fmt"
	"html/template"
)

type Manager struct {
	functionMap template.FuncMap
}

type ActionIterator struct {
}

func NewManager() *Manager {
	return &Manager{
		functionMap: make(template.FuncMap),
	}
}

func (manager *Manager) AddTemplateExpansionFunctionSet(functionMap template.FuncMap) *Manager {
	for k, v := range functionMap {
		if manager.functionMap[k] != nil {
			panic(fmt.Sprintf("An attempt was made to AddTemplateExpansionFunctionSet with key [%s] but that function key is already defined.  Manually merge maps before calling AddTemplateExpansionFunctionSet if you really want to do this.", k))
		}

		manager.functionMap[k] = v
	}

	return manager
}

func (manager *Manager) PrepareActionsFromStringList(actions []string) error {
	return nil
}

func (manager *Manager) ActionIterator() *ActionIterator {
	return nil
}

func (iterator *ActionIterator) Next() bool {
	return false
}

func (iterator *ActionIterator) Value() Action {
	return nil
}
