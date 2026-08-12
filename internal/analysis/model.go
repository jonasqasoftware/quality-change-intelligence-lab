package analysis

// ChangeSet is an explicitly provided description of a proposed change. It is
// not inferred from a production system or a source-control provider.
type ChangeSet struct {
	ID          string      `json:"id"`
	Summary     string      `json:"summary"`
	Source      string      `json:"source"`
	Components  []Component `json:"components"`
	Constraints Constraints `json:"constraints"`
}

type Component struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Criticality string   `json:"criticality"`
	ChangeType  string   `json:"changeType"`
	RiskTags    []string `json:"riskTags"`
}

type Constraints struct {
	MaxEstimatedMinutes int      `json:"maxEstimatedMinutes"`
	AllowedLevels       []string `json:"allowedLevels"`
}

// TestCase is a catalog entry. The catalog is a planning input, not evidence
// that a test exists or has passed in another environment.
type TestCase struct {
	ID                string   `json:"id"`
	Level             string   `json:"level"`
	EstimatedMinutes  int      `json:"estimatedMinutes"`
	ComponentKeywords []string `json:"componentKeywords"`
	CoversRiskTags    []string `json:"coversRiskTags"`
	Rationale         string   `json:"rationale"`
}

type Risk struct {
	ID         string `json:"id"`
	Component  string `json:"component"`
	Tag        string `json:"tag"`
	Score      int    `json:"score"`
	Reason     string `json:"reason"`
	ChangePath string `json:"changePath"`
}

type Recommendation struct {
	TestID           string   `json:"testId"`
	Level            string   `json:"level"`
	EstimatedMinutes int      `json:"estimatedMinutes"`
	Benefit          int      `json:"benefit"`
	CoveredRiskIDs   []string `json:"coveredRiskIds"`
	CoveredRiskTags  []string `json:"coveredRiskTags"`
	Rationale        string   `json:"rationale"`
}

type DeferredRisk struct {
	RiskID string `json:"riskId"`
	Reason string `json:"reason"`
}

type Assessment struct {
	ChangeID              string           `json:"changeId"`
	ChangeSummary         string           `json:"changeSummary"`
	Source                string           `json:"source"`
	BudgetMinutes         int              `json:"budgetMinutes"`
	SelectedMinutes       int              `json:"selectedMinutes"`
	Risks                 []Risk           `json:"risks"`
	Recommendations       []Recommendation `json:"recommendations"`
	DeferredRisks         []DeferredRisk   `json:"deferredRisks"`
	ResidualRiskStatement string           `json:"residualRiskStatement"`
}
