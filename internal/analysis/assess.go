package analysis

import (
	"fmt"
	"sort"
	"strings"
)

var criticalityWeights = map[string]int{
	"low":    1,
	"medium": 3,
	"high":   5,
}

var changeTypeWeights = map[string]int{
	"config":         2,
	"ui":             2,
	"logic":          4,
	"dependency":     4,
	"api-contract":   5,
	"data-migration": 5,
	"security":       5,
}

// Assess creates a deterministic test recommendation. It does not execute
// tests and it does not treat planned coverage as release approval.
func Assess(change ChangeSet, catalog []TestCase) Assessment {
	risks := buildRisks(change)
	candidates := buildCandidates(change, risks, catalog)
	sortCandidates(candidates)

	selectedMinutes := 0
	selected := make([]Recommendation, 0)
	covered := make(map[string]bool)
	for _, candidate := range candidates {
		if selectedMinutes+candidate.EstimatedMinutes > change.Constraints.MaxEstimatedMinutes {
			continue
		}
		benefit := 0
		newRiskIDs := make([]string, 0, len(candidate.CoveredRiskIDs))
		for _, riskID := range candidate.CoveredRiskIDs {
			if !covered[riskID] {
				benefit += riskScoreByID(risks, riskID)
				newRiskIDs = append(newRiskIDs, riskID)
			}
		}
		if benefit == 0 {
			continue
		}
		candidate.Benefit = benefit
		candidate.CoveredRiskIDs = newRiskIDs
		selected = append(selected, candidate)
		selectedMinutes += candidate.EstimatedMinutes
		for _, riskID := range newRiskIDs {
			covered[riskID] = true
		}
	}

	deferred := make([]DeferredRisk, 0)
	for _, risk := range risks {
		if covered[risk.ID] {
			continue
		}
		deferred = append(deferred, DeferredRisk{
			RiskID: risk.ID,
			Reason: deferredReason(risk.ID, candidates, change.Constraints.MaxEstimatedMinutes-selectedMinutes),
		})
	}

	return Assessment{
		ChangeID:              change.ID,
		ChangeSummary:         change.Summary,
		Source:                change.Source,
		BudgetMinutes:         change.Constraints.MaxEstimatedMinutes,
		SelectedMinutes:       selectedMinutes,
		Risks:                 risks,
		Recommendations:       selected,
		DeferredRisks:         deferred,
		ResidualRiskStatement: "A selected test plan demonstrates planned verification only. Test execution results, environment conditions, data quality, and untested interactions remain residual risk.",
	}
}

func buildRisks(change ChangeSet) []Risk {
	risks := make([]Risk, 0)
	for componentIndex, component := range change.Components {
		score := criticalityWeights[component.Criticality] + changeTypeWeights[component.ChangeType]
		for tagIndex, tag := range component.RiskTags {
			risks = append(risks, Risk{
				ID:         fmt.Sprintf("R-%02d-%02d", componentIndex+1, tagIndex+1),
				Component:  component.Name,
				Tag:        tag,
				Score:      score,
				ChangePath: component.Path,
				Reason:     fmt.Sprintf("%s criticality (%d) + %s change (%d)", component.Criticality, criticalityWeights[component.Criticality], component.ChangeType, changeTypeWeights[component.ChangeType]),
			})
		}
	}
	sort.Slice(risks, func(left, right int) bool {
		if risks[left].Score == risks[right].Score {
			return risks[left].ID < risks[right].ID
		}
		return risks[left].Score > risks[right].Score
	})
	return risks
}

func buildCandidates(change ChangeSet, risks []Risk, catalog []TestCase) []Recommendation {
	allowedLevels := asSet(change.Constraints.AllowedLevels)
	candidates := make([]Recommendation, 0)
	for _, testCase := range catalog {
		if !allowedLevels[strings.ToLower(testCase.Level)] {
			continue
		}

		coveredRiskIDs := make([]string, 0)
		coveredTags := make([]string, 0)
		for _, risk := range risks {
			if testMatchesRisk(testCase, risk) {
				coveredRiskIDs = append(coveredRiskIDs, risk.ID)
				coveredTags = appendUnique(coveredTags, risk.Tag)
			}
		}
		if len(coveredRiskIDs) == 0 {
			continue
		}
		candidates = append(candidates, Recommendation{
			TestID:           testCase.ID,
			Level:            testCase.Level,
			EstimatedMinutes: testCase.EstimatedMinutes,
			Benefit:          scoreForRiskIDs(risks, coveredRiskIDs),
			CoveredRiskIDs:   coveredRiskIDs,
			CoveredRiskTags:  coveredTags,
			Rationale:        testCase.Rationale,
		})
	}
	return candidates
}

func testMatchesRisk(testCase TestCase, risk Risk) bool {
	componentText := strings.ToLower(risk.Component + " " + risk.ChangePath)
	keywordMatched := false
	for _, keyword := range testCase.ComponentKeywords {
		if strings.Contains(componentText, strings.ToLower(keyword)) {
			keywordMatched = true
			break
		}
	}
	if !keywordMatched {
		return false
	}
	for _, tag := range testCase.CoversRiskTags {
		if strings.EqualFold(tag, risk.Tag) {
			return true
		}
	}
	return false
}

func sortCandidates(candidates []Recommendation) {
	sort.Slice(candidates, func(left, right int) bool {
		leftValue := candidates[left].Benefit * candidates[right].EstimatedMinutes
		rightValue := candidates[right].Benefit * candidates[left].EstimatedMinutes
		if leftValue == rightValue {
			if candidates[left].Benefit == candidates[right].Benefit {
				return candidates[left].TestID < candidates[right].TestID
			}
			return candidates[left].Benefit > candidates[right].Benefit
		}
		return leftValue > rightValue
	})
}

func deferredReason(riskID string, candidates []Recommendation, remainingMinutes int) string {
	for _, candidate := range candidates {
		for _, coveredRiskID := range candidate.CoveredRiskIDs {
			if coveredRiskID == riskID {
				return fmt.Sprintf("matching planned test(s) exceed the remaining %d minute budget", remainingMinutes)
			}
		}
	}
	return "no matching test in the allowed catalog levels"
}

func scoreForRiskIDs(risks []Risk, riskIDs []string) int {
	total := 0
	for _, riskID := range riskIDs {
		total += riskScoreByID(risks, riskID)
	}
	return total
}

func riskScoreByID(risks []Risk, riskID string) int {
	for _, risk := range risks {
		if risk.ID == riskID {
			return risk.Score
		}
	}
	return 0
}

func asSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[strings.ToLower(value)] = true
	}
	return result
}

func appendUnique(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}
