# Quality Change Intelligence Lab

Quality Change Intelligence Lab is a small, deterministic CLI that converts an explicit change description and a test catalog into a risk-prioritized test plan. It demonstrates how Quality Engineering can make change-impact decisions reviewable without claiming that a planned check ran, passed, or guarantees release readiness.

> The provided inputs are synthetic laboratory fixtures. They are not production evidence, client data, or a record of a real release.

## Why it exists

When a change arrives, running every test can be slow and uninformative; relying on intuition is difficult to review. This lab makes the planning assumptions visible:

- component criticality and change type define a transparent risk score;
- a test must match both a component keyword and a declared risk tag;
- a supplied time budget constrains a maximum-risk test plan rather than a static first-match list;
- every catalog entry receives a traceable planning decision: selected, not selected, or excluded by an explicit constraint;
- deferred checks are recorded as residual risk rather than hidden.

The project complements [QuickMetrics](https://github.com/jonasqasoftware/quickmetrics-project): QuickMetrics explains test-run signals after execution, while this lab explains why particular verification should be prioritized before execution.

## Run it

Requirements: Go 1.18 or newer.

```bash
go run ./cmd/qci
```

Generate a reusable planning artifact:

```bash
mkdir -p reports
go run ./cmd/qci --out reports/sample-change-report.md
```

Use a different change description and catalog:

```bash
go run ./cmd/qci \
  --changes path/to/change.json \
  --catalog path/to/test-catalog.json \
  --format json
```

## Example decision

The synthetic fixture describes a high-criticality payment authorization logic change and a medium-criticality checkout UI change. With an 18-minute planning budget, the tool selects the combination of mapped checks with the greatest total risk score. The report identifies whether it used exact optimization or the documented large-catalog fallback, and still states that results, environments, test data, and unmapped interactions remain residual risk.

## Quality Strategy

Context, risks, test layers, decision criteria, evidence boundaries, and residual risks are described in [Quality Strategy](docs/QUALITY_STRATEGY.md). The deterministic rule set is detailed in [Decision Model](docs/DECISION_MODEL.md).

Run the local quality gates:

```bash
gofmt -w .
go vet ./...
go test -race -coverprofile=coverage.out ./...
go build ./...
go run ./cmd/qci --out /tmp/qci-sample-report.md
```

The CI workflow runs these same checks and uploads the generated synthetic planning report as an artifact. GitHub Actions artifacts preserve files produced during a workflow run; they are used here as evidence of a particular CI run, not as a claim about a production release.

## Architecture

```text
change-set JSON + test-catalog JSON
              |
              v
strict decoding -> validation -> transparent risk model -> constrained selection
                                                            |
                                                            v
                                                Markdown or JSON planning report
```

The project uses the Go standard library to keep the decision logic easy to inspect. It deliberately has no database, API integration, dashboard, or AI model: those additions would increase scope before strengthening the core reasoning.

## Limitations

- It does not inspect a real pull-request diff, test-management system, or CI provider.
- Risk tags, component keywords, estimates, and criticality are declared inputs requiring human review.
- A catalog entry means a proposed check, not an implemented or passing test.
- Exact maximum-risk selection is limited to 20 mapped catalog candidates; larger inputs use a labelled deterministic greedy fallback.
- The selected plan is not a release approval, coverage metric, or predictive reliability score.

## Project structure

```text
cmd/qci/              CLI boundary
internal/analysis/    input validation, risk assessment, reporting, and tests
testdata/             clearly labeled synthetic fixtures
docs/                 quality strategy and decision model
.github/workflows/    repeatable quality gates and artifact creation
```

## License

[MIT](LICENSE)
