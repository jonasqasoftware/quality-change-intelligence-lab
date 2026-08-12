# Quality Strategy

## Context

Teams often react to code changes by either running every available check or relying on intuition. Both approaches obscure the decision: the first can be slow and noisy; the second can omit a high-impact risk.

Quality Change Intelligence Lab is a local planning tool that turns an explicit change description and a declared test catalog into a deterministic, reviewable test recommendation. The included fixtures are synthetic. They are not a record of a real product change, a release gate, or production test evidence.

## Primary risks

| Risk | Why it matters | Mitigation in this lab |
| --- | --- | --- |
| Hidden prioritization rules | Teams cannot challenge a black-box recommendation. | Criticality and change-type weights are visible in code and documented. |
| Incorrect catalog mapping | A test can be selected for the wrong component or risk. | Matching requires both a component keyword and an exact risk tag. |
| Budget hides material risk | A short test window can defer an important check. | Deferred risks are emitted with an explicit reason. |
| Planning treated as proof | A selected test is mistaken for a passing result. | Every report states its evidence boundary and residual risk. |

## Strategy and test levels

| Level | Purpose | Evidence produced |
| --- | --- | --- |
| Unit | Validate weight, matching, sorting, budget, and residual-risk rules. | Deterministic Go test results. |
| Integration | Load strict JSON input and render Markdown/JSON reports through the CLI boundary. | CLI output from synthetic fixtures. |
| Workflow | Format, vet, race-test, build, generate a sample report, and upload it as a CI artifact. | A GitHub Actions artifact tied to a workflow run. |

## Decision criteria

A recommendation is selected only when it:

1. belongs to an allowed test level;
2. matches both the changed component and the stated risk tag;
3. adds previously uncovered risk; and
4. fits within the supplied time budget.

Candidates are ordered by uncovered risk score per estimated minute. Ties favor greater uncovered risk score, then stable test ID ordering. The scoring model is a transparent heuristic for discussion, not a universal release formula.

## Residual risks and limits

- Inputs are manually declared; this lab does not parse a real Git diff or connect to issue trackers.
- The catalog is a planning model. It does not prove test implementation, execution, or pass status.
- Estimates, component keywords, criticality, and risk tags need human review.
- A report should inform a quality decision alongside execution results, product context, and known incidents; it must not replace them.
