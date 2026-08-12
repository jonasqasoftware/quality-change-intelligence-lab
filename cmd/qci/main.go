package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jonasqasoftware/quality-change-intelligence-lab/internal/analysis"
)

func main() {
	changesPath := flag.String("changes", "testdata/sample-change.json", "path to a change-set JSON file")
	catalogPath := flag.String("catalog", "testdata/test-catalog.json", "path to a test catalog JSON file")
	format := flag.String("format", "markdown", "output format: markdown or json")
	outputPath := flag.String("out", "", "optional report output file")
	flag.Parse()

	change, err := analysis.LoadChangeSet(*changesPath)
	if err != nil {
		fail(err)
	}
	catalog, err := analysis.LoadCatalog(*catalogPath)
	if err != nil {
		fail(err)
	}

	assessment := analysis.Assess(change, catalog)
	output, err := render(*format, assessment)
	if err != nil {
		fail(err)
	}
	if err := write(*outputPath, output); err != nil {
		fail(err)
	}
}

func render(format string, assessment analysis.Assessment) ([]byte, error) {
	switch format {
	case "markdown":
		return []byte(analysis.MarkdownReport(assessment)), nil
	case "json":
		output, err := json.MarshalIndent(assessment, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("encode JSON report: %w", err)
		}
		return append(output, '\n'), nil
	default:
		return nil, fmt.Errorf("unsupported format %q; use markdown or json", format)
	}
}

func write(path string, output []byte) error {
	if path == "" {
		_, err := os.Stdout.Write(output)
		return err
	}
	return os.WriteFile(path, output, 0o644)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "qci: %v\n", err)
	flag.CommandLine.SetOutput(io.Discard)
	os.Exit(1)
}
