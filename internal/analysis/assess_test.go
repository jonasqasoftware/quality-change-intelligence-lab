package analysis

import (
	"fmt"
	"strings"
	"testing"
)

func TestAssessSelectsHighestValuePlanWithinBudget(t *testing.T) {
	change := ChangeSet{
		ID:      "CHG-1",
		Summary: "payment and checkout change",
		Source:  "synthetic",
		Components: []Component{
			{
				Name:        "payment authorization",
				Path:        "services/payments/authorization.go",
				Criticality: "high",
				ChangeType:  "logic",
				RiskTags:    []string{"authorization", "financial-integrity"},
			},
			{
				Name:        "checkout form",
				Path:        "web/checkout/form.ts",
				Criticality: "medium",
				ChangeType:  "ui",
				RiskTags:    []string{"accessibility"},
			},
		},
		Constraints: Constraints{MaxEstimatedMinutes: 7, AllowedLevels: []string{"unit", "accessibility", "api"}},
	}
	catalog := []TestCase{
		{
			ID:                "api-payment",
			Level:             "api",
			EstimatedMinutes:  7,
			ComponentKeywords: []string{"payment"},
			CoversRiskTags:    []string{"authorization", "financial-integrity"},
		},
		{
			ID:                "unit-payment",
			Level:             "unit",
			EstimatedMinutes:  3,
			ComponentKeywords: []string{"authorization"},
			CoversRiskTags:    []string{"authorization", "financial-integrity"},
		},
		{
			ID:                "a11y-checkout",
			Level:             "accessibility",
			EstimatedMinutes:  4,
			ComponentKeywords: []string{"checkout"},
			CoversRiskTags:    []string{"accessibility"},
		},
	}

	assessment := Assess(change, catalog)
	if assessment.SelectedMinutes != 7 {
		t.Fatalf("selected minutes = %d, want 7", assessment.SelectedMinutes)
	}
	if len(assessment.Recommendations) != 2 {
		t.Fatalf("recommendation count = %d, want 2", len(assessment.Recommendations))
	}
	if assessment.Recommendations[0].TestID != "unit-payment" || assessment.Recommendations[1].TestID != "a11y-checkout" {
		t.Fatalf("unexpected plan: %#v", assessment.Recommendations)
	}
	if len(assessment.DeferredRisks) != 0 {
		t.Fatalf("deferred risks = %#v, want none", assessment.DeferredRisks)
	}
	if assessment.SelectionMethod != "exact-maximum-risk-coverage" {
		t.Fatalf("selection method = %q", assessment.SelectionMethod)
	}
	assertCatalogDecision(t, assessment.CatalogDecisions, "unit-payment", "selected")
	assertCatalogDecision(t, assessment.CatalogDecisions, "a11y-checkout", "selected")
	assertCatalogDecision(t, assessment.CatalogDecisions, "api-payment", "not-selected")
}

func TestAssessFindsBestCombinationInsteadOfGreedyFirstChoice(t *testing.T) {
	change := ChangeSet{
		ID:      "CHG-COMBO",
		Summary: "four equally weighted rules change together",
		Source:  "synthetic",
		Components: []Component{
			{Name: "alpha rule", Path: "rules/alpha.go", Criticality: "medium", ChangeType: "ui", RiskTags: []string{"alpha"}},
			{Name: "beta rule", Path: "rules/beta.go", Criticality: "medium", ChangeType: "ui", RiskTags: []string{"beta"}},
			{Name: "gamma rule", Path: "rules/gamma.go", Criticality: "medium", ChangeType: "ui", RiskTags: []string{"gamma"}},
			{Name: "delta rule", Path: "rules/delta.go", Criticality: "medium", ChangeType: "ui", RiskTags: []string{"delta"}},
		},
		Constraints: Constraints{MaxEstimatedMinutes: 10, AllowedLevels: []string{"unit"}},
	}
	catalog := []TestCase{
		{ID: "wide-first-choice", Level: "unit", EstimatedMinutes: 6, ComponentKeywords: []string{"alpha", "beta"}, CoversRiskTags: []string{"alpha", "beta"}},
		{ID: "alpha-gamma", Level: "unit", EstimatedMinutes: 5, ComponentKeywords: []string{"alpha", "gamma"}, CoversRiskTags: []string{"alpha", "gamma"}},
		{ID: "beta-delta", Level: "unit", EstimatedMinutes: 5, ComponentKeywords: []string{"beta", "delta"}, CoversRiskTags: []string{"beta", "delta"}},
	}

	assessment := Assess(change, catalog)
	if assessment.SelectedMinutes != 10 {
		t.Fatalf("selected minutes = %d, want 10", assessment.SelectedMinutes)
	}
	if len(assessment.Recommendations) != 2 {
		t.Fatalf("recommendation count = %d, want 2", len(assessment.Recommendations))
	}
	selectedIDs := []string{assessment.Recommendations[0].TestID, assessment.Recommendations[1].TestID}
	if !containsAll(selectedIDs, "alpha-gamma", "beta-delta") {
		t.Fatalf("selected tests = %v, want alpha-gamma and beta-delta", selectedIDs)
	}
	if len(assessment.DeferredRisks) != 0 {
		t.Fatalf("deferred risks = %#v, want none", assessment.DeferredRisks)
	}
}

func TestSelectPlanLabelsLargeCatalogFallback(t *testing.T) {
	candidates := make([]Recommendation, exactOptimizationCandidateLimit+1)
	for index := range candidates {
		candidates[index] = Recommendation{
			TestID:           fmt.Sprintf("candidate-%02d", index),
			EstimatedMinutes: 1,
			CoveredRiskIDs:   []string{"R-01"},
		}
	}

	indexes, method, note := selectPlan(candidates, []Risk{{ID: "R-01", Score: 5}}, 1)
	if method != "deterministic-greedy-fallback" {
		t.Fatalf("method = %q", method)
	}
	if !strings.Contains(note, "exceed the exact optimization limit") {
		t.Fatalf("note = %q", note)
	}
	if len(indexes) != 1 {
		t.Fatalf("selected candidates = %v, want one", indexes)
	}
}

func TestAssessDefersRiskWithoutMatchingAllowedTest(t *testing.T) {
	change := ChangeSet{
		ID:      "CHG-2",
		Summary: "migration",
		Source:  "synthetic",
		Components: []Component{{
			Name:        "account migration",
			Path:        "db/migration.sql",
			Criticality: "high",
			ChangeType:  "data-migration",
			RiskTags:    []string{"data-loss"},
		}},
		Constraints: Constraints{MaxEstimatedMinutes: 5, AllowedLevels: []string{"unit"}},
	}
	catalog := []TestCase{{
		ID:                "api-migration",
		Level:             "api",
		EstimatedMinutes:  3,
		ComponentKeywords: []string{"migration"},
		CoversRiskTags:    []string{"data-loss"},
	}}

	assessment := Assess(change, catalog)
	if len(assessment.Recommendations) != 0 {
		t.Fatalf("recommendations = %#v, want none", assessment.Recommendations)
	}
	if len(assessment.DeferredRisks) != 1 {
		t.Fatalf("deferred risks = %#v, want one", assessment.DeferredRisks)
	}
	if got := assessment.DeferredRisks[0].Reason; got != "no matching test in the allowed catalog levels" {
		t.Fatalf("deferred reason = %q", got)
	}
	assertCatalogDecision(t, assessment.CatalogDecisions, "api-migration", "excluded")
}

func TestMarkdownReportStatesItsLimits(t *testing.T) {
	assessment := Assessment{
		ChangeID:              "CHG-3",
		ChangeSummary:         "sample",
		Source:                "synthetic",
		BudgetMinutes:         10,
		SelectionMethod:       "exact-maximum-risk-coverage",
		SelectionNote:         "Exact optimization evaluated all valid combinations.",
		ResidualRiskStatement: "Planning is not proof.",
	}

	assessment.CatalogDecisions = []CatalogDecision{{
		TestID:           "unit-payment",
		Level:            "unit",
		EstimatedMinutes: 3,
		Status:           "selected",
		MappedRiskIDs:    []string{"R-01"},
		Reason:           "Selected by exact-maximum-risk-coverage within the supplied time budget.",
	}}

	report := MarkdownReport(assessment)
	for _, expected := range []string{"does not claim that tests ran", "## Limits", "Planning is not proof.", "exact-maximum-risk-coverage", "## Decision trace for catalog entries", "unit-payment"} {
		if !strings.Contains(report, expected) {
			t.Fatalf("report does not contain %q:\n%s", expected, report)
		}
	}
}

func assertCatalogDecision(t *testing.T, decisions []CatalogDecision, testID, status string) {
	t.Helper()
	for _, decision := range decisions {
		if decision.TestID == testID {
			if decision.Status != status {
				t.Fatalf("catalog decision for %q = %q, want %q", testID, decision.Status, status)
			}
			return
		}
	}
	t.Fatalf("no catalog decision found for %q", testID)
}

func containsAll(values []string, expected ...string) bool {
	found := make(map[string]bool, len(values))
	for _, value := range values {
		found[value] = true
	}
	for _, value := range expected {
		if !found[value] {
			return false
		}
	}
	return true
}
