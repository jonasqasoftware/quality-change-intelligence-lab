package analysis

import (
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
}

func TestMarkdownReportStatesItsLimits(t *testing.T) {
	assessment := Assessment{
		ChangeID:              "CHG-3",
		ChangeSummary:         "sample",
		Source:                "synthetic",
		BudgetMinutes:         10,
		ResidualRiskStatement: "Planning is not proof.",
	}

	report := MarkdownReport(assessment)
	for _, expected := range []string{"does not claim that tests ran", "## Limits", "Planning is not proof."} {
		if !strings.Contains(report, expected) {
			t.Fatalf("report does not contain %q:\n%s", expected, report)
		}
	}
}
