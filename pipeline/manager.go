package pipeline

import (
	"fmt"
	"html/template"
)

type Manager struct {
	functionMap                 template.FuncMap
	actionFactory               *ActionFactory
	pipelineActionDirectoryPath string
	pipelineActionSet           []Action
}

type ActionIterator struct {
	pipelineActionSetQueue []Action
	pendingAction          Action
}

func NewManager(pipelineActionDirectoryPath string, actionFactory *ActionFactory) *Manager {
	return &Manager{
		functionMap:                 make(template.FuncMap),
		actionFactory:               actionFactory,
		pipelineActionDirectoryPath: pipelineActionDirectoryPath,
		pipelineActionSet:           []Action{},
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
	manager.pipelineActionSet = make([]Action, 0, len(actions))

	for _, actionDescriptor := range actions {
		action, err := manager.actionFactory.NewActionFromStringDescriptor(actionDescriptor, manager.pipelineActionDirectoryPath)
		if err != nil {
			return err
		}

		manager.pipelineActionSet = append(manager.pipelineActionSet, action)
	}

	return nil
}

func (manager *Manager) ActionIterator() *ActionIterator {
	return &ActionIterator{
		pipelineActionSetQueue: manager.pipelineActionSet,
	}
}

func (iterator *ActionIterator) Next() bool {
	if len(iterator.pipelineActionSetQueue) > 0 {
		iterator.pendingAction = iterator.pipelineActionSetQueue[0]
		iterator.pipelineActionSetQueue = iterator.pipelineActionSetQueue[1:]
		return true
	}

	return false
}

func (iterator *ActionIterator) Value() Action {
	return iterator.pendingAction
}
