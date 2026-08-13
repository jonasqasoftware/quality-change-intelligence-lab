package analysis

import (
	"fmt"
	"strings"
)

func MarkdownReport(assessment Assessment) string {
	var report strings.Builder
	fmt.Fprintln(&report, "# Change Intelligence Report")
	fmt.Fprintf(&report, "\n> Input source: %s\n", assessment.Source)
	fmt.Fprintln(&report, "> This is a deterministic planning report. It does not claim that tests ran, passed, or establish release readiness.")
	fmt.Fprintf(&report, "\n## Change\n\n- **ID:** `%s`\n- **Summary:** %s\n- **Time budget:** %d minutes\n- **Selected plan:** %d minutes\n- **Selection method:** `%s`\n- **Selection note:** %s\n", assessment.ChangeID, assessment.ChangeSummary, assessment.BudgetMinutes, assessment.SelectedMinutes, assessment.SelectionMethod, assessment.SelectionNote)

	fmt.Fprint(&report, "\n## Prioritized risks\n\n")
	fmt.Fprintln(&report, "| ID | Component | Risk | Score | Decision basis |")
	fmt.Fprintln(&report, "| --- | --- | --- | ---: | --- |")
	for _, risk := range assessment.Risks {
		fmt.Fprintf(&report, "| `%s` | %s | `%s` | %d | %s |\n", risk.ID, risk.Component, risk.Tag, risk.Score, risk.Reason)
	}

	fmt.Fprint(&report, "\n## Recommended test plan\n\n")
	if len(assessment.Recommendations) == 0 {
		fmt.Fprintln(&report, "No tests were selected within the supplied catalog and constraints.")
	} else {
		fmt.Fprintln(&report, "| Test | Level | Estimate | Newly addressed risks | Why selected |")
		fmt.Fprintln(&report, "| --- | --- | ---: | --- | --- |")
		for _, recommendation := range assessment.Recommendations {
			fmt.Fprintf(&report, "| `%s` | %s | %d min | %s | %s |\n", recommendation.TestID, recommendation.Level, recommendation.EstimatedMinutes, strings.Join(recommendation.CoveredRiskIDs, ", "), recommendation.Rationale)
		}
	}

	fmt.Fprint(&report, "\n## Deferred risks and residual risk\n\n")
	if len(assessment.DeferredRisks) == 0 {
		fmt.Fprintln(&report, "All catalog-mapped risks fit inside the selected plan. This is not proof that all change risk is covered.")
	} else {
		for _, deferred := range assessment.DeferredRisks {
			fmt.Fprintf(&report, "- `%s`: %s.\n", deferred.RiskID, deferred.Reason)
		}
	}
	fmt.Fprintf(&report, "\n%s\n", assessment.ResidualRiskStatement)
	fmt.Fprintln(&report, "\n## Limits\n\n- Matching uses explicit component keywords and risk tags; it does not inspect a real repository diff.\n- Estimates are planning inputs, not observed execution duration.\n- A catalog entry is a proposed verification activity, not pass/fail evidence.\n- Human review is required before using this plan for release decisions.")
	return report.String()
}
