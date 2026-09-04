# Formato das entradas

A ferramenta lê dois arquivos JSON: um **change-set** (a mudança) e um **catálogo de testes** (as verificações candidatas). Ambos usam decodificação estrita — um campo desconhecido faz o carregamento falhar, para que um erro de digitação no contrato apareça imediatamente em vez de mudar a recomendação em silêncio.

## Change-set

Arquivo com um único objeto JSON.

```json
{
  "id": "DEMO-CHG-042",
  "summary": "Strengthen payment authorization handling and update checkout validation",
  "source": "synthetic laboratory fixture — not production evidence",
  "components": [
    {
      "name": "payment authorization",
      "path": "services/payments/authorization.go",
      "criticality": "high",
      "changeType": "logic",
      "riskTags": ["authorization", "financial-integrity"]
    }
  ],
  "constraints": {
    "maxEstimatedMinutes": 18,
    "allowedLevels": ["unit", "api", "e2e", "accessibility"]
  }
}
```

| Campo | Tipo | Obrigatório | Significado |
| --- | --- | --- | --- |
| `id` | string | sim | Identificador da mudança. Aparece no relatório para rastreabilidade. |
| `summary` | string | sim | Descrição curta da mudança, em texto livre. |
| `source` | string | sim | De onde essa descrição veio (ex.: `"synthetic laboratory fixture"`, `"change description reviewed by QE"`). O campo registra a origem declarada da entrada; a ferramenta não verifica sua veracidade. |
| `components` | array de objetos | sim, ao menos 1 | Os componentes alterados por essa mudança. |
| `constraints` | objeto | sim | O orçamento de tempo e os níveis de teste permitidos. |

### `components[]`

| Campo | Tipo | Obrigatório | Valores aceitos | Significado |
| --- | --- | --- | --- | --- |
| `name` | string | sim | livre | Nome do componente, usado no relatório e na correspondência com `componentKeywords` do catálogo. |
| `path` | string | sim | livre | Caminho aproximado do componente no repositório. Também entra na correspondência com `componentKeywords`. |
| `criticality` | string | sim | `low`, `medium`, `high` | Peso de criticidade do componente (ver [Modelo de Decisão](DECISION_MODEL.md)). Qualquer outro valor faz o carregamento falhar. |
| `changeType` | string | sim | `config`, `ui`, `logic`, `dependency`, `api-contract`, `data-migration`, `security` | Peso do tipo de mudança. Qualquer outro valor faz o carregamento falhar. |
| `riskTags` | array de strings | sim, ao menos 1 | livre | Os riscos declarados para esse componente (ex.: `"authorization"`, `"accessibility"`). Cada tag vira um risco individual no relatório. |

### `constraints`

| Campo | Tipo | Obrigatório | Significado |
| --- | --- | --- | --- |
| `maxEstimatedMinutes` | inteiro > 0 | sim | Orçamento total de tempo para o plano selecionado. |
| `allowedLevels` | array de strings, ao menos 1 | sim | Níveis de teste permitidos (ex.: `"unit"`, `"e2e"`, `"accessibility"`). Comparado sem diferenciar maiúsculas/minúsculas com o campo `level` de cada item do catálogo. Um nível fora dessa lista é sempre excluído, mesmo que o teste esteja associado a um risco mapeado. |

## Catálogo de testes

Arquivo com um array de objetos JSON.

```json
[
  {
    "id": "unit-payment-policy",
    "level": "unit",
    "estimatedMinutes": 3,
    "componentKeywords": ["payment", "authorization"],
    "coversRiskTags": ["authorization", "financial-integrity"],
    "rationale": "Fast validation of authorization rules and amount boundaries."
  }
]
```

| Campo | Tipo | Obrigatório | Significado |
| --- | --- | --- | --- |
| `id` | string | sim, único no catálogo | Identificador do teste/verificação proposta. |
| `level` | string | sim | Nível do teste (ex.: `unit`, `api`, `e2e`, `accessibility`, `performance`). Só é considerado se estiver em `constraints.allowedLevels` do change-set. |
| `estimatedMinutes` | inteiro > 0 | sim | Estimativa de duração usada para caber no orçamento. Não é uma medição real de execução. |
| `componentKeywords` | array de strings, ao menos 1 | sim | Palavras comparadas (sem diferenciar maiúsculas/minúsculas) contra `nome + caminho` do componente. Basta uma palavra aparecer como substring para o componente ser considerado correspondente. |
| `coversRiskTags` | array de strings, ao menos 1 | sim | As `riskTags` do change-set que esse teste declara cobrir. Essa associação é uma entrada de planejamento, não evidência de cobertura executada. A comparação é exata (sem diferenciar maiúsculas/minúsculas), não por substring. |
| `rationale` | string | não | Texto livre explicando por que esse teste existe. Aparece no relatório quando o teste é recomendado. |

## Como um teste se conecta a um risco

Cada combinação `componente + riskTag` do change-set vira um **risco** com uma pontuação (criticidade + tipo de mudança). Um teste do catálogo só é considerado associado (mapeado) a um risco quando **as duas condições** são verdadeiras ao mesmo tempo:

1. pelo menos uma palavra de `componentKeywords` aparece dentro de `nome do componente + caminho do componente` (correspondência por substring, sem diferenciar maiúsculas/minúsculas);
2. pelo menos uma tag de `coversRiskTags` é igual a uma `riskTag` do componente (correspondência exata, sem diferenciar maiúsculas/minúsculas).

Se apenas uma das duas condições for satisfeita, o teste não está associado àquele risco. Um teste sem nenhum risco mapeado, ou com um `level` fora de `allowedLevels`, é marcado como `excluded` na tabela de decisões — ver [Como interpretar o resultado](../README.md#como-interpretar-o-resultado) no README.
