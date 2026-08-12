# Decision Model

## Inputs

The CLI reads two strict JSON files:

- a **change set**: identifier, summary, declared input source, changed components, criticality, change type, risk tags, and constraints;
- a **test catalog**: proposed checks, levels, estimates, component keywords, covered risk tags, and rationale.

Unknown JSON fields are rejected. This makes accidental contract drift visible rather than silently changing a recommendation.

## Scoring

Each component-risk pair receives:

```text
risk score = criticality weight + change-type weight
```

| Criticality | Weight | Change type | Weight |
| --- | ---: | --- | ---: |
| low | 1 | config | 2 |
| medium | 3 | ui | 2 |
| high | 5 | logic | 4 |
|  |  | dependency | 4 |
|  |  | api-contract | 5 |
|  |  | data-migration | 5 |
|  |  | security | 5 |

The score represents a prioritization hypothesis, not impact observed in production.

## Selection

A test candidate maps to a risk only if a declared component keyword appears in that component's name or path **and** the candidate explicitly covers the same risk tag. Candidates are selected greedily by currently uncovered risk score per estimated minute, respecting the user-provided time budget.

The report makes deferred risks explicit when no mapped test is permitted, or when mapped tests cannot fit in the remaining budget. It intentionally does not calculate coverage percentages: planned mapping should not be presented as confidence or release readiness.
