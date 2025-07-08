package jobber

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

type ConfigurationDefaultNamespace struct {
	Basename string `yaml:"Basename"`
}

type TestCase struct {
	Name   string         `yaml:"Name"`
	Values map[string]any `yaml:"Values"`
}

type TestUnit struct {
	Name   string         `yaml:"Name"`
	Values map[string]any `yaml:"Values"`
}

type ConfigurationAssetArchive struct {
	FilePath string `yaml:"FilePath"`
}

type ConfigurationPipeline struct {
	ActionDefinitionsRootDirectory string            `yaml:"ActionDefinitionsRootDirectory"`
	ActionsInOrder                 []string          `yaml:"ActionsInOrder"`
	ExecutionEnvironment           map[string]string `yaml:"ExecutionEnvironment"`
}

type ConfigurationTest struct {
	AssetArchive     *ConfigurationAssetArchive     `yaml:"AssetArchive"`
	DefaultNamespace *ConfigurationDefaultNamespace `yaml:"DefaultNamespace"`
	GlobalValues     map[string]any                 `yaml:"GlobalValues"`
	Pipeline         *ConfigurationPipeline         `yaml:"Pipeline"`
	Cases            []*TestCase                    `yaml:"Cases"`
	Units            []*TestUnit                    `yaml:"Units"`
}

type Configuration struct {
	Test *ConfigurationTest `yaml:"Test"`
}

func (c *Configuration) validate() error {
	if c.Test == nil {
		return fmt.Errorf(".Test must exist")
	}

	if len(c.Test.Cases) == 0 {
		return fmt.Errorf(".Test.Cases must exist and cannot be an empty list")
	}

	if len(c.Test.Units) == 0 {
		return fmt.Errorf(".Test.Units must exist and cannot be an empty list")
	}

	if c.Test.AssetArchive == nil {
		return fmt.Errorf(".Test.AssetArchive must be defined")
	}

	if c.Test.AssetArchive.FilePath == "" {
		return fmt.Errorf(".Test.AssetArchive.FilePath must exist and cannot be the empty string")
	}

	if c.Test.DefaultNamespace == nil {
		return fmt.Errorf(".Test.DefaultNamespace must be defined")
	}

	if c.Test.DefaultNamespace.Basename == "" {
		return fmt.Errorf(".Test.DefaultNamespace.Basename must be defined and cannot be the empty string")
	}

	if c.Test.Pipeline == nil {
		return fmt.Errorf(".Test.Pipeline must be defined and cannot be empty")
	}

	if c.Test.Pipeline.ActionDefinitionsRootDirectory == "" {
		return fmt.Errorf(".Test.Pipeline.ActionDefinitionRootDirectory must be defined and cannot be empty")
	}

	if len(c.Test.Pipeline.ActionsInOrder) == 0 {
		return fmt.Errorf(".Test.Pipeline.ActionsInOrder must have at least one entry")
	}

	for pipelineEntryIndex, value := range c.Test.Pipeline.ActionsInOrder {
		s := strings.Split(value, "/")
		if len(s) != 2 {
			return fmt.Errorf(".Test.Pipeline.ActionsInOrder[%d] must be of format <type>/<target>", pipelineEntryIndex)
		}
		switch s[0] {
		case "resources":
		case "values-transforms":
		case "executables":
		default:
			return fmt.Errorf(".Test.Pipeline.ActionsInOrder[%d] type indicator [%s] is not understood", pipelineEntryIndex, s[0])
		}
	}

	return nil
}

func (c *Configuration) expandDefaults() {
	if c.Test.GlobalValues == nil {
		c.Test.GlobalValues = make(map[string]any)
	}

	for _, testCase := range c.Test.Cases {
		if testCase.Values == nil {
			testCase.Values = make(map[string]any)
		}
	}

	for _, testUnit := range c.Test.Units {
		if testUnit.Values == nil {
			testUnit.Values = make(map[string]any)
		}
	}
}

func (c *Configuration) MergeOverrideValues(overrideValues map[string]any) error {
	for overrideKey, overrideValue := range overrideValues {
		overrideKey = trimFirstRuneInStringIfItMatches(overrideKey, '.')
		if overrideKey == "" {
			return fmt.Errorf("override key (%s) cannot be empty", overrideKey)
		}

		if !strings.HasPrefix(overrideKey, "Test.") {
			return fmt.Errorf("no such configuration key (%s); must start with .Test or Test", overrideKey)
		}

		keyStack := strings.Split(overrideKey, ".")
		if len(keyStack) < 3 {
			return fmt.Errorf("no such configuration key (%s)", overrideKey)
		}

		keyStack = keyStack[1:] // Remove the leading "Test" part

		switch keyStack[0] {
		case "AssetArchive":
			if err := c.assetArchiveOverride(keyStack[1:], overrideValue, overrideKey); err != nil {
				return fmt.Errorf("failed to apply override for (%s): %s", overrideKey, err)
			}

		case "DefaultNamespace":
			if err := c.defaultNamespaceOverride(keyStack[1:], overrideValue, overrideKey); err != nil {
				return fmt.Errorf("failed to apply override for (%s): %s", overrideKey, err)
			}

		case "GlobalValues":
			if err := c.globalValuesOverride(keyStack[1:], overrideValue, overrideKey); err != nil {
				return fmt.Errorf("failed to apply override for (%s): %s", overrideKey, err)
			}

		case "Pipeline":
			if err := c.pipelineOverride(keyStack[1:], overrideValue, overrideKey); err != nil {
				return fmt.Errorf("failed to apply override for (%s): %s", overrideKey, err)
			}

		case "Cases":
			if err := c.casesOverride(keyStack[1:], overrideValue, overrideKey); err != nil {
				return fmt.Errorf("failed to apply override for (%s): %s", overrideKey, err)
			}

		case "Units":
			if err := c.unitsOverride(keyStack[1:], overrideValue, overrideKey); err != nil {
				return fmt.Errorf("failed to apply override for (%s): %s", overrideKey, err)
			}

		default:
			return fmt.Errorf("no such configuration key (%s)", overrideKey)
		}
	}

	return nil
}

func (c *Configuration) assetArchiveOverride(subKeyStack []string, overrideValue any, originalOverrideKey string) error {
	if len(subKeyStack) != 1 {
		return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
	}

	if subKeyStack[0] != "FilePath" {
		return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
	}

	if overrideValueAsString := overrideValueToString(overrideValue); overrideValueAsString == "" {
		return fmt.Errorf("override value for .Test.AssetArchive.FilePath cannot be the empty string")
	} else {
		c.Test.AssetArchive.FilePath = overrideValueAsString
	}

	return nil
}

func (c *Configuration) defaultNamespaceOverride(subKeyStack []string, overrideValue any, originalOverrideKey string) error {
	if len(subKeyStack) != 1 {
		return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
	}

	if subKeyStack[0] != "Basename" {
		return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
	}

	if overrideValueAsString := overrideValueToString(overrideValue); overrideValueAsString == "" {
		return fmt.Errorf("override value for .Test.DefaultNamespace.Basename cannot be the empty string")
	} else {
		c.Test.DefaultNamespace.Basename = overrideValueAsString
	}

	return nil
}

func (c *Configuration) globalValuesOverride(subKeyStack []string, overrideValue any, originalOverrideKey string) error {
	return setConfigurationKeyValueInMultiLevelMap(c.Test.GlobalValues, subKeyStack, overrideValue, originalOverrideKey)
}

var listSelectorRegexp = regexp.MustCompile(`^\[(\d+)\]$`)

func (c *Configuration) pipelineOverride(subKeyStack []string, overrideValue any, originalOverrideKey string) error {
	switch subKeyStack[0] {
	case "ActionDefinitionsRootDirectory":
		if len(subKeyStack) != 1 {
			return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
		}

		if overrideValueAsString := overrideValueToString(overrideValue); overrideValueAsString == "" {
			return fmt.Errorf("override value for (%s) cannot be the empty string", originalOverrideKey)
		} else {
			c.Test.Pipeline.ActionDefinitionsRootDirectory = overrideValueAsString
		}

	case "ExecutionEnvironment":
		if len(subKeyStack) != 2 {
			return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
		}

		if c.Test.Pipeline.ExecutionEnvironment == nil {
			c.Test.Pipeline.ExecutionEnvironment = make(map[string]string)
		}

		c.Test.Pipeline.ExecutionEnvironment[subKeyStack[1]] = overrideValueToString(overrideValue)

	case "ActionsInOrder":
		if len(subKeyStack) != 2 {
			return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
		}

		listSelectorMatch := listSelectorRegexp.FindStringSubmatch(subKeyStack[1])
		if len(listSelectorMatch) != 2 {
			return fmt.Errorf("the configuration key (%s) must use a list index selector in square brackets", originalOverrideKey)
		}

		listSelectorIndex, _ := strconv.Atoi(listSelectorMatch[1])

		switch {
		case listSelectorIndex < len(c.Test.Pipeline.ActionsInOrder):
			c.Test.Pipeline.ActionsInOrder[listSelectorIndex] = overrideValueToString(overrideValue)
		case listSelectorIndex == len(c.Test.Pipeline.ActionsInOrder):
			c.Test.Pipeline.ActionsInOrder = append(c.Test.Pipeline.ActionsInOrder, overrideValueToString(overrideValue))
		default:
			return fmt.Errorf("list selector index (%d) out of range in (%s)", listSelectorIndex, originalOverrideKey)
		}
	}

	return nil
}

var caseOrUnitNameSelectorRegexp = regexp.MustCompile(`^\[(\S+)\]$`)

func (c *Configuration) casesOverride(subKeyStack []string, overrideValue any, originalOverrideKey string) error {
	if len(subKeyStack) < 1 { // Must be at least [$CaseName].Values.$key but to make more descriptive error message, process starting with first key
		return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
	}

	caseNameSelectorMatch := caseOrUnitNameSelectorRegexp.FindStringSubmatch(subKeyStack[0])
	if len(caseNameSelectorMatch) != 2 {
		return fmt.Errorf("the configuration key (%s) must use a casename selector in square brackets", originalOverrideKey)
	}

	caseNameSelectorAsAString := caseNameSelectorMatch[1]

	if len(subKeyStack) < 2 {
		return fmt.Errorf("must select override for (%s)", originalOverrideKey)
	}

	var matchingCase *TestCase = nil

	for _, testCase := range c.Test.Cases {
		if testCase.Name == caseNameSelectorAsAString {
			matchingCase = testCase
			break
		}
	}

	if matchingCase == nil {
		return fmt.Errorf("for the configuration key (%s), there is no such case named (%s)", originalOverrideKey, caseNameSelectorAsAString)
	}

	if subKeyStack[1] != "Values" {
		return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
	}

	if len(subKeyStack) < 3 {
		return fmt.Errorf("cannot override non-leaf configuration key (%s)", originalOverrideKey)
	}

	return setConfigurationKeyValueInMultiLevelMap(matchingCase.Values, subKeyStack[2:], overrideValue, originalOverrideKey)
}

func (c *Configuration) unitsOverride(subKeyStack []string, overrideValue any, originalOverrideKey string) error {
	if len(subKeyStack) < 1 { // Must be at least [$CaseName].Values.$key but to make more descriptive error message, process starting with first key
		return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
	}

	unitNameSelectorMatch := caseOrUnitNameSelectorRegexp.FindStringSubmatch(subKeyStack[0])
	if len(unitNameSelectorMatch) != 2 {
		return fmt.Errorf("the configuration key (%s) must use a unitname selector in square brackets", originalOverrideKey)
	}

	unitNameSelectorAsAString := unitNameSelectorMatch[1]

	if len(subKeyStack) < 2 {
		return fmt.Errorf("must select override for (%s)", originalOverrideKey)
	}

	var matchingCase *TestUnit = nil

	for _, testCase := range c.Test.Units {
		if testCase.Name == unitNameSelectorAsAString {
			matchingCase = testCase
			break
		}
	}

	if matchingCase == nil {
		return fmt.Errorf("for the configuration key (%s), there is no such unit named (%s)", originalOverrideKey, unitNameSelectorAsAString)
	}

	if subKeyStack[1] != "Values" {
		return fmt.Errorf("no such configuration key (%s)", originalOverrideKey)
	}

	if len(subKeyStack) < 3 {
		return fmt.Errorf("cannot override non-leaf configuration key (%s)", originalOverrideKey)
	}

	return setConfigurationKeyValueInMultiLevelMap(matchingCase.Values, subKeyStack[2:], overrideValue, originalOverrideKey)
}

func ReadConfigurationYamlFromReader(r io.Reader) (*Configuration, error) {
	c := &Configuration{}

	encoder := yaml.NewDecoder(r)
	err := encoder.Decode(c)

	if err != nil {
		return nil, err
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	c.expandDefaults()

	return c, err
}

func setConfigurationKeyValueInMultiLevelMap(m map[string]any, keyStack []string, value any, originalKey string) error {
	if m == nil || len(keyStack) == 0 || value == nil {
		return nil
	}

	if len(keyStack) > 1 {
		if _, exists := m[keyStack[0]]; !exists {
			return fmt.Errorf("no such configuration key (%s)", originalKey)
		}

		if reflect.ValueOf(m[keyStack[0]]).Kind() != reflect.Map {
			return fmt.Errorf("cannot set a non-leaf configuration key (%s)", originalKey)
		}

		return setConfigurationKeyValueInMultiLevelMap(m[keyStack[0]].(map[string]any), keyStack[1:], value, originalKey)
	}

	if _, exists := m[keyStack[0]]; !exists {
		return fmt.Errorf("no such configuration key (%s)", originalKey)
	}

	if reflect.ValueOf(m[keyStack[0]]).Kind() != reflect.ValueOf(value).Kind() {
		coercedValue, err := attemptCoersionToYamlMarshallType(m[keyStack[0]], value)
		if err != nil {
			return fmt.Errorf("failed to coerce value for (%s): %w", originalKey, err)
		}
		m[keyStack[0]] = coercedValue
	} else {
		m[keyStack[0]] = value
	}

	return nil
}

func attemptCoersionToYamlMarshallType(existingValue, newValue any) (any, error) {
	switch existingValue.(type) {
	case string:
		return fmt.Sprintf("%v", newValue), nil

	case int:
		switch n := newValue.(type) {
		case string:
			convertedValue, err := strconv.Atoi(n)
			if err != nil {
				return newValue, fmt.Errorf("failed to convert string (%s) to int: %w", newValue, err)
			}

			return convertedValue, nil

		case uint:
			return int(n), nil
		}

	case bool:
		switch n := newValue.(type) {
		case string:
			convertedValue, err := strconv.ParseBool(n)
			if err != nil {
				return newValue, fmt.Errorf("failed to convert string (%s) to bool: %w", newValue, err)
			}

			return convertedValue, nil
		}
	}

	return newValue, fmt.Errorf("cannot coerce value (%v) of type %T to type %T", newValue, newValue, existingValue)
}

func ReadConfigurationYamlFromFile(filePath string) (*Configuration, error) {
	fh, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file (%s): %s", filePath, err)
	}
	defer fh.Close()

	return ReadConfigurationYamlFromReader(fh)
}

func (c *Configuration) CharactersInLongestCaseName() uint {
	longest := 0
	for _, c := range c.Test.Cases {
		if len(c.Name) > longest {
			longest = len(c.Name)
		}
	}

	return uint(longest)
}

func (c *Configuration) CharactersInLongestUnitName() uint {
	longest := 0
	for _, c := range c.Test.Units {
		if len(c.Name) > longest {
			longest = len(c.Name)
		}
	}

	return uint(longest)
}

func trimFirstRuneInStringIfItMatches(s string, r rune) string {
	if len(s) == 0 {
		return s
	}

	decodedRune, i := utf8.DecodeRuneInString(s)
	if decodedRune == r {
		return s[i:]
	}

	return s
}

func overrideValueToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
