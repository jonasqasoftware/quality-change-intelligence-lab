# Quality Change Intelligence Lab

<p align="center">
  <img src="docs/brand/social-preview.png" alt="Jonas Davila: engenharia de qualidade com evidência, clareza e método" width="100%">
</p>

[![CI](https://github.com/jonasqasoftware/quality-change-intelligence-lab/actions/workflows/quality.yml/badge.svg)](https://github.com/jonasqasoftware/quality-change-intelligence-lab/actions/workflows/quality.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-1e2024.svg)](LICENSE) [![Jonas Davila](https://img.shields.io/badge/portfolio-jonasdavila.com.br-a8432f.svg)](https://jonasdavila.com.br/en/)

> **In English:** A Go CLI for risk-based test planning. Given a described change, candidate checks and a time budget, it selects the checks that cover the most risk and reports what was left out. It plans; it does not execute.

Ferramenta de Quality Engineering para priorizar verificações antes da execução, considerando riscos declarados e um orçamento de tempo. O resultado deixa explícito o que foi priorizado, o que ficou de fora e quais riscos continuam sem verificação.

> Os exemplos deste repositório são fixtures sintéticos, criados para este laboratório — não são evidência de produção, dados de cliente, nem o registro de um release real.

## O problema

Diante de uma mudança, é comum escolher entre dois extremos: rodar todo o conjunto de testes disponível — lento e nem sempre informativo — ou confiar na experiência de quem revisa — difícil de questionar depois. Nos dois casos, a decisão em si costuma ficar invisível: ninguém consegue apontar, depois, por que um teste específico entrou ou ficou de fora do escopo verificado.

## O que a ferramenta faz

A partir de uma descrição estruturada da mudança e de um catálogo de verificações candidatas, a ferramenta:

- calcula uma pontuação de risco por componente alterado, combinando criticidade e tipo de mudança;
- casa cada verificação do catálogo com os riscos que ela cobre, usando palavra-chave de componente e tag de risco declaradas;
- seleciona, dentro do orçamento de tempo informado, o conjunto de verificações que cobre o maior risco possível;
- registra, para cada item do catálogo, se ele foi selecionado, não selecionado ou excluído — e por quê;
- lista os riscos que ficaram sem verificação dentro do orçamento (risco residual).

A ferramenta **não**:

- executa testes;
- lê automaticamente um pull request ou inspeciona um diff;
- prevê defeitos;
- mede cobertura de código;
- aprova um release;
- substitui a decisão de quem faz QA/QE.

## Quando usar

- Para decidir, de forma revisável, quais verificações priorizar quando o tempo não permite rodar tudo — deixando claro para quem revisa o motivo de cada teste incluído ou excluído.
- Para discutir e documentar, com um time, os critérios de priorização de risco antes de uma mudança entrar em revisão.

## Quando não usar

- Como substituto de execução de testes, de gate de release ou de métrica de cobertura — a ferramenta planeja, não executa.
- Como fonte automática de verdade sobre o que mudou — ela depende de um change-set escrito manualmente, não lê o diff real do repositório.
- Isoladamente, sem revisão humana das entradas (criticidade, tipo de mudança, tags de risco) e sem considerar resultados de execução, contexto de produto e incidentes conhecidos.

## Começando em poucos minutos

Requisitos: Go 1.18 ou mais recente.

```bash
go run ./cmd/qci
```

Esse comando roda com os fixtures sintéticos incluídos no repositório e imprime o relatório em Markdown no terminal.

Para salvar o relatório em um arquivo:

```bash
mkdir -p reports
go run ./cmd/qci --out reports/sample-change-report.md
```

Para gerar em JSON em vez de Markdown:

```bash
go run ./cmd/qci --format json
```

## Exemplo prático

O fixture sintético descreve duas mudanças: uma alteração de lógica em autorização de pagamento (criticidade alta) e uma alteração de UI no formulário de checkout (criticidade média), com um orçamento de 18 minutos.

```text
mudança
  ↓
catálogo de verificações
  ↓
orçamento de tempo
  ↓
priorização por risco
  ↓
plano selecionado + risco residual
```

Executando `go run ./cmd/qci`, o plano selecionado usa 16 dos 18 minutos de orçamento e cobre os quatro riscos mapeados com três testes: `unit-payment-policy`, `a11y-checkout-form` e `e2e-checkout-validation`. Nenhum risco ficou adiado.

<details>
<summary>Ver o relatório completo (saída real, em inglês — o CLI ainda não tem saída localizada)</summary>

```markdown
# Change Intelligence Report

> Input source: synthetic laboratory fixture — not production evidence
> This is a deterministic planning report. It does not claim that tests ran, passed, or establish release readiness.

## Change

- **ID:** `DEMO-CHG-042`
- **Summary:** Strengthen payment authorization handling and update checkout validation
- **Time budget:** 18 minutes
- **Selected plan:** 16 minutes
- **Selection method:** `exact-maximum-risk-coverage`
- **Selection note:** Exact optimization evaluated all valid combinations of 4 mapped candidate(s).

## Prioritized risks

| ID | Component | Risk | Score | Decision basis |
| --- | --- | --- | ---: | --- |
| `R-01-01` | payment authorization | `authorization` | 9 | high criticality (5) + logic change (4) |
| `R-01-02` | payment authorization | `financial-integrity` | 9 | high criticality (5) + logic change (4) |
| `R-02-01` | checkout form | `input-validation` | 5 | medium criticality (3) + ui change (2) |
| `R-02-02` | checkout form | `accessibility` | 5 | medium criticality (3) + ui change (2) |

## Recommended test plan

| Test | Level | Estimate | Newly addressed risks | Why selected |
| --- | --- | ---: | --- | --- |
| `unit-payment-policy` | unit | 3 min | R-01-01, R-01-02 | Fast validation of authorization rules and amount boundaries. |
| `a11y-checkout-form` | accessibility | 4 min | R-02-02 | Checks keyboard and automated accessibility assertions for the form. |
| `e2e-checkout-validation` | e2e | 9 min | R-02-01 | Checks the critical checkout validation journey in a browser. |

## Decision trace for catalog entries

| Test | Level | Estimate | Status | Mapped risks | Decision reason |
| --- | --- | ---: | --- | --- | --- |
| `a11y-checkout-form` | accessibility | 4 min | selected | R-02-02 | Selected by exact-maximum-risk-coverage within the supplied time budget. |
| `api-payment-authorization-contract` | api | 7 min | not-selected | R-01-01, R-01-02 | Not selected because its explicitly mapped risks are already addressed by the selected plan. |
| `e2e-checkout-validation` | e2e | 9 min | selected | R-02-01 | Selected by exact-maximum-risk-coverage within the supplied time budget. |
| `performance-payment-baseline` | performance | 20 min | excluded | — | Excluded because test level "performance" is not permitted by the supplied constraints. |
| `unit-payment-policy` | unit | 3 min | selected | R-01-01, R-01-02 | Selected by exact-maximum-risk-coverage within the supplied time budget. |

## Deferred risks and residual risk

All catalog-mapped risks fit inside the selected plan. This is not proof that all change risk is covered.

A selected test plan demonstrates planned verification only. Test execution results, environment conditions, data quality, and untested interactions remain residual risk.

## Limits

- Matching uses explicit component keywords and risk tags; it does not inspect a real repository diff.
- Estimates are planning inputs, not observed execution duration.
- A catalog entry is a proposed verification activity, not pass/fail evidence.
- Human review is required before using this plan for release decisions.
```

</details>

`api-payment-authorization-contract` (7 min) e `performance-payment-baseline` (20 min) não entraram no plano: o primeiro porque os riscos que cobre já foram atendidos por um teste mais barato (`unit-payment-policy`); o segundo porque o nível `performance` não está entre os níveis permitidos pelo change-set. Essa é exatamente a rastreabilidade que a ferramenta se propõe a entregar — ver [Como interpretar o resultado](#como-interpretar-o-resultado).

## Usando seus próprios dados

```bash
go run ./cmd/qci \
  --changes caminho/para/sua-mudanca.json \
  --catalog caminho/para/seu-catalogo.json \
  --format json
```

O change-set descreve a mudança (componentes alterados, criticidade, tipo de mudança, tags de risco e orçamento). O catálogo descreve as verificações candidatas (nível, estimativa, palavras-chave de componente e tags de risco cobertas). O contrato completo de campos, valores aceitos e a regra de correspondência entre teste e risco estão documentados em [Formato das Entradas](docs/FORMATO-DAS-ENTRADAS.md) — não é necessário ler o código-fonte para descobrir o que cada campo espera.

Um campo desconhecido em qualquer um dos dois arquivos faz o carregamento falhar com uma mensagem de erro, em vez de ser ignorado silenciosamente.

## Como interpretar o resultado

- **`selected`**: o teste está no plano escolhido dentro do orçamento informado.
- **`not-selected`**: o teste é elegível (casa com um risco mapeado), mas ficou fora do plano — porque o risco que ele cobre já foi atendido por outro teste, ou porque a regra de seleção preferiu um conjunto diferente. Não selecionado **não** significa inútil.
- **`excluded`**: o teste não tem uma correspondência explícita de palavra-chave de componente e tag de risco com a mudança declarada, ou seu nível não está entre os permitidos.
- **Risco residual / riscos adiados**: riscos que nenhum teste selecionado cobre dentro do orçamento. A ferramenta lista isso de forma explícita em vez de escondê-lo.
- **Plano selecionado ≠ software pronto para release.** O relatório registra uma decisão de planejamento — o que foi priorizado e por quê — não uma execução, nem um resultado de teste, nem uma aprovação.

## Como funciona

```text
JSON de mudança + JSON de catálogo de testes
              |
              v
decodificação estrita -> validação -> modelo de risco transparente -> seleção com restrição de orçamento
                                                                        |
                                                                        v
                                                        Relatório em Markdown ou JSON
```

O modelo completo de pontuação e seleção — pesos de criticidade e tipo de mudança, regra de correspondência, otimização exata e fallback guloso — está descrito em [Modelo de Decisão](docs/DECISION_MODEL.md).

O projeto usa apenas a biblioteca padrão do Go, para manter a lógica de decisão fácil de inspecionar — sem banco de dados, API, dashboard ou IA, que aumentariam o escopo sem fortalecer esse raciocínio central.

## Estratégia de qualidade

Contexto, riscos, níveis de teste, critérios de decisão, limites de evidência e riscos residuais estão descritos em [Estratégia de Qualidade](docs/QUALITY_STRATEGY.md).

Rode os quality gates locais:

```bash
gofmt -w .
go vet ./...
go test -race -coverprofile=coverage.out ./...
go build ./...
go run ./cmd/qci --out /tmp/qci-sample-report.md
```

O workflow de CI roda essas mesmas verificações e publica o relatório sintético gerado como artefato — evidência de uma execução de CI específica, não de um release de produção.

## Limitações

- Não inspeciona um diff real de pull request, sistema de gestão de testes ou provedor de CI.
- Tags de risco, palavras-chave de componente, estimativas e criticidade são entradas declaradas que exigem revisão humana.
- Um item do catálogo representa uma verificação proposta, não um teste implementado ou com execução comprovada.
- A seleção exata de risco máximo é limitada a 20 candidatos mapeados; entradas maiores usam o fallback guloso determinístico documentado.
- O plano selecionado não é uma aprovação de release, uma métrica de cobertura, nem uma previsão de confiabilidade.
- Os pesos de criticidade e tipo de mudança são fixos no código; ajustá-los hoje exige alterar e recompilar o CLI, não um arquivo de configuração.

## Estrutura do projeto

```text
cmd/qci/              fronteira do CLI
internal/analysis/    validação de entrada, modelo de risco, geração de relatório e testes
testdata/             fixtures sintéticos claramente identificados
docs/                 estratégia de qualidade, modelo de decisão e formato das entradas
.github/workflows/    quality gates repetíveis e geração de artefato
```

## Projetos relacionados

- [QuickMetrics](https://github.com/jonasqasoftware/quickmetrics-project): atua depois da execução, interpretando sinais de resultados de teste. Este laboratório atua antes da execução, priorizando o que verificar diante de uma mudança e de um orçamento.
- **AIMA**: explora decisões de Quality Engineering em um escopo mais amplo (análise, impacto, métricas e ação). Quality Change Intelligence Lab se concentra especificamente na priorização de verificações diante de uma mudança e de um orçamento limitado — os dois projetos são independentes.

## Licença

[MIT](LICENSE)
