package jobber

import (
	"fmt"
	"os"
)

type TestCaseAssetsDirectoryCreationOutcome struct {
	SuccessfullyCreatedDirectoryPaths []string
	DirectoryPathOfFailedCreation     string
	DirectoryCreationFailureError     error
}

type TestCaseDirectoryPaths struct {
	Root              string
	ExpandedTemplates string
	RetrievedAssets   string
}

type ContextualAssetsDirectoryManager struct {
	testRootAssetDirectoryPath                   string
	testUnitAssetDirectoryPathByUnitName         map[string]string
	testCaseAssetsDirectoryPathByUnitAndCaseName map[string]map[string]*TestCaseDirectoryPaths
}

func NewContextualAssetsDirectoryManager() *ContextualAssetsDirectoryManager {
	return &ContextualAssetsDirectoryManager{
		testUnitAssetDirectoryPathByUnitName:         make(map[string]string),
		testCaseAssetsDirectoryPathByUnitAndCaseName: make(map[string]map[string]*TestCaseDirectoryPaths),
	}
}

func (m *ContextualAssetsDirectoryManager) CreateTestAssetsRootDirectory() error {
	createdDirectoryPath, err := os.MkdirTemp("", "jobber.")
	if err != nil {
		return err
	}

	m.testRootAssetDirectoryPath = createdDirectoryPath
	return nil
}

func (m *ContextualAssetsDirectoryManager) CreateTestUnitDirectory(testUnit *TestUnit) error {
	if m.testRootAssetDirectoryPath == "" {
		panic("attempt to CreateTestUnitDirectory() before CreateTestAssetsRootDirectory()")
	}

	proposedPath := fmt.Sprintf("%s/%s", m.testRootAssetDirectoryPath, testUnit.Name)
	m.testUnitAssetDirectoryPathByUnitName[testUnit.Name] = proposedPath

	m.testCaseAssetsDirectoryPathByUnitAndCaseName[testUnit.Name] = make(map[string]*TestCaseDirectoryPaths)

	return os.Mkdir(proposedPath, 0700)
}

func (m *ContextualAssetsDirectoryManager) CreateTestCaseDirectories(testUnit *TestUnit, testCase *TestCase) *TestCaseAssetsDirectoryCreationOutcome {
	testUnitAssetDirectoryPath := m.TestUnitAssetDirectoryPathFor(testUnit)
	if testUnitAssetDirectoryPath == "" {
		panic("attempt to CreateTestCaseDirectories() before corresponding CreateTestUnitDirectory()")
	}

	outcome := &TestCaseAssetsDirectoryCreationOutcome{
		SuccessfullyCreatedDirectoryPaths: make([]string, 0),
	}

	proposedTestCaseRootPath := fmt.Sprintf("%s/%s", testUnitAssetDirectoryPath, testCase.Name)

	if err := os.Mkdir(proposedTestCaseRootPath, 0700); err != nil {
		outcome.DirectoryPathOfFailedCreation = proposedTestCaseRootPath
		outcome.DirectoryCreationFailureError = err
		return outcome
	}

	outcome.SuccessfullyCreatedDirectoryPaths = append(outcome.SuccessfullyCreatedDirectoryPaths, proposedTestCaseRootPath)

	proposedExpandedTemplatesPath := fmt.Sprintf("%s/%s", proposedTestCaseRootPath, "expanded-templates")
	proposedRetrievedAssetsPath := fmt.Sprintf("%s/%s", proposedTestCaseRootPath, "retrieved-assets")

	for _, proposedPath := range []string{proposedExpandedTemplatesPath, proposedRetrievedAssetsPath} {
		if err := os.Mkdir(proposedPath, 0700); err != nil {
			outcome.DirectoryPathOfFailedCreation = proposedPath
			outcome.DirectoryCreationFailureError = err
			return outcome
		}

		outcome.SuccessfullyCreatedDirectoryPaths = append(outcome.SuccessfullyCreatedDirectoryPaths, proposedPath)
	}

	m.testCaseAssetsDirectoryPathByUnitAndCaseName[testUnit.Name][testCase.Name] = &TestCaseDirectoryPaths{
		Root:              proposedTestCaseRootPath,
		ExpandedTemplates: proposedExpandedTemplatesPath,
		RetrievedAssets:   proposedRetrievedAssetsPath,
	}

	return outcome
}

func (m *ContextualAssetsDirectoryManager) TestRootAssetDirectoryPath() string {
	return m.testRootAssetDirectoryPath
}

func (m *ContextualAssetsDirectoryManager) TestUnitAssetDirectoryPathFor(testUnit *TestUnit) string {
	return m.testUnitAssetDirectoryPathByUnitName[testUnit.Name]
}

func (m *ContextualAssetsDirectoryManager) TestCaseAssetsDirectoryPathsFor(testUnit *TestUnit, testCase *TestCase) *TestCaseDirectoryPaths {
	if s := m.testCaseAssetsDirectoryPathByUnitAndCaseName[testUnit.Name]; s != nil {
		return s[testCase.Name]
	}

	return nil
}
