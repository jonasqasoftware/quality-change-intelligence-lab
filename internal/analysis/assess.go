package analysis

import (
	"fmt"
	"sort"
	"strings"
)

const exactOptimizationCandidateLimit = 20

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
	selectedIndexes, selectionMethod, selectionNote := selectPlan(candidates, risks, change.Constraints.MaxEstimatedMinutes)
	selected, selectedMinutes, covered := materializeRecommendations(candidates, selectedIndexes, risks)

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
		SelectionMethod:       selectionMethod,
		SelectionNote:         selectionNote,
		Risks:                 risks,
		Recommendations:       selected,
		DeferredRisks:         deferred,
		ResidualRiskStatement: "A selected test plan demonstrates planned verification only. Test execution results, environment conditions, data quality, and untested interactions remain residual risk.",
	}
}

// selectPlan chooses the maximum mapped risk score that fits inside the
// supplied budget. The exact search is deliberately bounded so a large,
// manually supplied catalog cannot make a local planning command unresponsive.
// A transparent deterministic fallback is used beyond the bound.
func selectPlan(candidates []Recommendation, risks []Risk, budget int) ([]int, string, string) {
	if len(candidates) <= exactOptimizationCandidateLimit {
		indexes := selectOptimalCandidateIndexes(candidates, risks, budget)
		return indexes,
			"exact-maximum-risk-coverage",
			fmt.Sprintf("Exact optimization evaluated all valid combinations of %d mapped candidate(s).", len(candidates))
	}

	indexes := selectGreedyCandidateIndexes(candidates, risks, budget)
	return indexes,
		"deterministic-greedy-fallback",
		fmt.Sprintf("The %d mapped candidates exceed the exact optimization limit of %d, so this report uses a deterministic greedy fallback.", len(candidates), exactOptimizationCandidateLimit)
}

type plan struct {
	indexes []int
	minutes int
	score   int
}

func selectOptimalCandidateIndexes(candidates []Recommendation, risks []Risk, budget int) []int {
	best := plan{indexes: []int{}}
	searchOptimalPlan(candidates, risks, budget, 0, plan{indexes: []int{}}, &best)
	return best.indexes
}

func searchOptimalPlan(candidates []Recommendation, risks []Risk, budget, index int, current plan, best *plan) {
	if current.minutes > budget {
		return
	}
	if index == len(candidates) {
		current.score = planScore(candidates, risks, current.indexes)
		if betterPlan(current, *best, candidates) {
			*best = current
		}
		return
	}

	searchOptimalPlan(candidates, risks, budget, index+1, current, best)
	withCandidate := plan{
		indexes: append(append([]int{}, current.indexes...), index),
		minutes: current.minutes + candidates[index].EstimatedMinutes,
	}
	searchOptimalPlan(candidates, risks, budget, index+1, withCandidate, best)
}

func betterPlan(candidate, current plan, candidates []Recommendation) bool {
	if candidate.score != current.score {
		return candidate.score > current.score
	}
	if candidate.minutes != current.minutes {
		return candidate.minutes < current.minutes
	}
	return planKey(candidate.indexes, candidates) < planKey(current.indexes, candidates)
}

func planScore(candidates []Recommendation, risks []Risk, indexes []int) int {
	covered := make(map[string]bool)
	for _, index := range indexes {
		for _, riskID := range candidates[index].CoveredRiskIDs {
			covered[riskID] = true
		}
	}
	total := 0
	for riskID := range covered {
		total += riskScoreByID(risks, riskID)
	}
	return total
}

func planKey(indexes []int, candidates []Recommendation) string {
	ids := make([]string, 0, len(indexes))
	for _, index := range indexes {
		ids = append(ids, candidates[index].TestID)
	}
	sort.Strings(ids)
	return strings.Join(ids, "\x00")
}

func selectGreedyCandidateIndexes(candidates []Recommendation, risks []Risk, budget int) []int {
	remaining := make(map[int]bool, len(candidates))
	for index := range candidates {
		remaining[index] = true
	}
	covered := make(map[string]bool)
	selected := make([]int, 0)
	selectedMinutes := 0

	for {
		bestIndex := -1
		bestBenefit := 0
		for index := range remaining {
			candidate := candidates[index]
			if selectedMinutes+candidate.EstimatedMinutes > budget {
				continue
			}
			benefit := uncoveredBenefit(candidate, risks, covered)
			if benefit == 0 || !betterCandidate(index, bestIndex, benefit, bestBenefit, candidates) {
				continue
			}
			bestIndex = index
			bestBenefit = benefit
		}
		if bestIndex == -1 {
			break
		}

		selected = append(selected, bestIndex)
		selectedMinutes += candidates[bestIndex].EstimatedMinutes
		for _, riskID := range candidates[bestIndex].CoveredRiskIDs {
			covered[riskID] = true
		}
		delete(remaining, bestIndex)
	}
	return selected
}

func materializeRecommendations(candidates []Recommendation, indexes []int, risks []Risk) ([]Recommendation, int, map[string]bool) {
	remaining := make(map[int]bool, len(indexes))
	for _, index := range indexes {
		remaining[index] = true
	}
	covered := make(map[string]bool)
	selected := make([]Recommendation, 0, len(indexes))
	selectedMinutes := 0

	for len(remaining) > 0 {
		bestIndex := -1
		bestBenefit := 0
		for index := range remaining {
			benefit := uncoveredBenefit(candidates[index], risks, covered)
			if betterCandidate(index, bestIndex, benefit, bestBenefit, candidates) {
				bestIndex = index
				bestBenefit = benefit
			}
		}

		candidate := candidates[bestIndex]
		newRiskIDs := uncoveredRiskIDs(candidate, covered)
		candidate.Benefit = bestBenefit
		candidate.CoveredRiskIDs = newRiskIDs
		candidate.CoveredRiskTags = riskTagsForIDs(risks, newRiskIDs)
		selected = append(selected, candidate)
		selectedMinutes += candidate.EstimatedMinutes
		for _, riskID := range newRiskIDs {
			covered[riskID] = true
		}
		delete(remaining, bestIndex)
	}
	return selected, selectedMinutes, covered
}

func betterCandidate(index, bestIndex, benefit, bestBenefit int, candidates []Recommendation) bool {
	if bestIndex == -1 {
		return true
	}
	leftValue := benefit * candidates[bestIndex].EstimatedMinutes
	rightValue := bestBenefit * candidates[index].EstimatedMinutes
	if leftValue != rightValue {
		return leftValue > rightValue
	}
	if benefit != bestBenefit {
		return benefit > bestBenefit
	}
	return candidates[index].TestID < candidates[bestIndex].TestID
}

func uncoveredBenefit(candidate Recommendation, risks []Risk, covered map[string]bool) int {
	return scoreForRiskIDs(risks, uncoveredRiskIDs(candidate, covered))
}

func uncoveredRiskIDs(candidate Recommendation, covered map[string]bool) []string {
	newRiskIDs := make([]string, 0, len(candidate.CoveredRiskIDs))
	for _, riskID := range candidate.CoveredRiskIDs {
		if !covered[riskID] {
			newRiskIDs = append(newRiskIDs, riskID)
		}
	}
	return newRiskIDs
}

func riskTagsForIDs(risks []Risk, riskIDs []string) []string {
	tags := make([]string, 0, len(riskIDs))
	for _, riskID := range riskIDs {
		for _, risk := range risks {
			if risk.ID == riskID {
				tags = appendUnique(tags, risk.Tag)
				break
			}
		}
	}
	return tags
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
	sort.Slice(candidates, func(left, right int) bool {
		return candidates[left].TestID < candidates[right].TestID
	})
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

func deferredReason(riskID string, candidates []Recommendation, remainingMinutes int) string {
	minimumDuration := 0
	for _, candidate := range candidates {
		for _, coveredRiskID := range candidate.CoveredRiskIDs {
			if coveredRiskID == riskID {
				if minimumDuration == 0 || candidate.EstimatedMinutes < minimumDuration {
					minimumDuration = candidate.EstimatedMinutes
				}
			}
		}
	}
	if minimumDuration > 0 {
		return fmt.Sprintf("the shortest matching planned test requires %d minutes, exceeding the %d minutes remaining after selected coverage", minimumDuration, remainingMinutes)
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
