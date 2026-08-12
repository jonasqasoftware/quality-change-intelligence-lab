package main

import (
	"strings"
	"testing"

	"github.com/jonasqasoftware/quality-change-intelligence-lab/internal/analysis"
)

func TestRenderJSON(t *testing.T) {
	output, err := render("json", analysis.Assessment{ChangeID: "CHG-1"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), `"changeId": "CHG-1"`) {
		t.Fatalf("output = %s", output)
	}
}

func TestRenderRejectsUnsupportedFormat(t *testing.T) {
	_, err := render("xml", analysis.Assessment{})
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("error = %v", err)
	}
}
