package analysis

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadChangeSetRejectsUnknownFields(t *testing.T) {
	path := writeFixture(t, `{
  "id": "CHG-1",
  "summary": "sample",
  "source": "synthetic",
  "components": [{
    "name": "checkout",
    "path": "web/checkout.ts",
    "criticality": "high",
    "changeType": "logic",
    "riskTags": ["validation"],
    "unexpected": true
  }],
  "constraints": {"maxEstimatedMinutes": 5, "allowedLevels": ["unit"]}
}`)

	_, err := LoadChangeSet(path)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("error = %v, want unknown field", err)
	}
}

func TestLoadCatalogRejectsDuplicateIDs(t *testing.T) {
	path := writeFixture(t, `[
  {"id":"duplicate","level":"unit","estimatedMinutes":1,"componentKeywords":["checkout"],"coversRiskTags":["validation"],"rationale":"first"},
  {"id":"duplicate","level":"unit","estimatedMinutes":1,"componentKeywords":["checkout"],"coversRiskTags":["validation"],"rationale":"second"}
]`)

	_, err := LoadCatalog(path)
	if err == nil || !strings.Contains(err.Error(), "duplicate test id") {
		t.Fatalf("error = %v, want duplicate test id", err)
	}
}

func writeFixture(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
