package analysis

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func LoadChangeSet(path string) (ChangeSet, error) {
	var change ChangeSet
	if err := decodeFile(path, &change); err != nil {
		return ChangeSet{}, err
	}
	if err := validateChangeSet(change); err != nil {
		return ChangeSet{}, err
	}
	return change, nil
}

func LoadCatalog(path string) ([]TestCase, error) {
	var catalog []TestCase
	if err := decodeFile(path, &catalog); err != nil {
		return nil, err
	}
	if err := validateCatalog(catalog); err != nil {
		return nil, err
	}
	return catalog, nil
}

func decodeFile(path string, target interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode %q: %w", path, err)
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode %q: more than one JSON value", path)
		}
		return fmt.Errorf("decode %q: %w", path, err)
	}
	return nil
}

func validateChangeSet(change ChangeSet) error {
	if strings.TrimSpace(change.ID) == "" {
		return fmt.Errorf("change id is required")
	}
	if strings.TrimSpace(change.Summary) == "" {
		return fmt.Errorf("change summary is required")
	}
	if strings.TrimSpace(change.Source) == "" {
		return fmt.Errorf("change source is required; identify synthetic or trusted input")
	}
	if len(change.Components) == 0 {
		return fmt.Errorf("at least one changed component is required")
	}
	if change.Constraints.MaxEstimatedMinutes <= 0 {
		return fmt.Errorf("maxEstimatedMinutes must be positive")
	}
	if len(change.Constraints.AllowedLevels) == 0 {
		return fmt.Errorf("at least one allowed test level is required")
	}

	for index, component := range change.Components {
		if strings.TrimSpace(component.Name) == "" || strings.TrimSpace(component.Path) == "" {
			return fmt.Errorf("component %d requires name and path", index+1)
		}
		if _, ok := criticalityWeights[component.Criticality]; !ok {
			return fmt.Errorf("component %q has unsupported criticality %q", component.Name, component.Criticality)
		}
		if _, ok := changeTypeWeights[component.ChangeType]; !ok {
			return fmt.Errorf("component %q has unsupported changeType %q", component.Name, component.ChangeType)
		}
		if len(component.RiskTags) == 0 {
			return fmt.Errorf("component %q requires at least one risk tag", component.Name)
		}
	}
	return nil
}

func validateCatalog(catalog []TestCase) error {
	if len(catalog) == 0 {
		return fmt.Errorf("test catalog cannot be empty")
	}
	seen := make(map[string]bool)
	for index, testCase := range catalog {
		if strings.TrimSpace(testCase.ID) == "" || strings.TrimSpace(testCase.Level) == "" {
			return fmt.Errorf("catalog entry %d requires id and level", index+1)
		}
		if seen[testCase.ID] {
			return fmt.Errorf("duplicate test id %q", testCase.ID)
		}
		seen[testCase.ID] = true
		if testCase.EstimatedMinutes <= 0 {
			return fmt.Errorf("test %q must have a positive estimatedMinutes", testCase.ID)
		}
		if len(testCase.ComponentKeywords) == 0 || len(testCase.CoversRiskTags) == 0 {
			return fmt.Errorf("test %q requires componentKeywords and coversRiskTags", testCase.ID)
		}
	}
	return nil
}
