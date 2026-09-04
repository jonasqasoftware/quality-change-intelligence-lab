# Modelo de Decisão

Este documento descreve, em detalhe, as regras que o CLI aplica para transformar um change-set e um catálogo de testes em um plano priorizado. O contrato completo de campos está em [Formato das Entradas](FORMATO-DAS-ENTRADAS.md); aqui o foco é a lógica de pontuação e seleção.

## Entradas

O CLI lê dois arquivos JSON estritos:

- um **change-set**: identificador, resumo, origem declarada, componentes alterados, criticidade, tipo de mudança, tags de risco e restrições;
- um **catálogo de testes**: verificações propostas, nível, estimativa, palavras-chave de componente, tags de risco que declara contemplar e justificativa.

Campos JSON desconhecidos são rejeitados. Isso torna visível qualquer divergência acidental do contrato, em vez de deixá-la mudar uma recomendação silenciosamente.

## Pontuação de risco

Cada par componente/tag de risco recebe:

```text
pontuação do risco = peso da criticidade + peso do tipo de mudança
```

| Criticidade | Peso | Tipo de mudança | Peso |
| --- | ---: | --- | ---: |
| low | 1 | config | 2 |
| medium | 3 | ui | 2 |
| high | 5 | logic | 4 |
| | | dependency | 4 |
| | | api-contract | 5 |
| | | data-migration | 5 |
| | | security | 5 |

A pontuação representa uma hipótese de priorização, não um impacto observado em produção. Os pesos usados para calculá-la são escolhas deste laboratório, feitas para tornar essa hipótese explícita e reproduzível — não um padrão universal, nem uma probabilidade estatística de defeito, um impacto financeiro observado ou uma escala validada pelo setor.

## O que a pontuação significa

A pontuação serve apenas para ordenar e priorizar candidatos dentro deste modelo. Ela não representa:

- probabilidade estatística de defeito;
- severidade observada em produção;
- impacto financeiro previsto;
- porcentagem de cobertura;
- confiança ou prontidão de release.

## Correspondência entre teste e risco

Um teste do catálogo só é elegível para se associar a um risco quando, ao mesmo tempo: (1) uma de suas `componentKeywords` aparece como substring em `nome do componente + caminho do componente`, e (2) uma de suas `coversRiskTags` é igual a uma `riskTag` daquele componente. Ambas as comparações ignoram maiúsculas/minúsculas. Um teste cujo `level` não está em `allowedLevels` é sempre excluído, mesmo que esteja associado a um risco mapeado.

## Seleção do plano

Para um catálogo de até **20 candidatos mapeados** (testes associados a ao menos um risco e cujo nível é permitido), o CLI avalia todas as combinações possíveis que cabem no orçamento e escolhe a de maior pontuação total de risco contemplado pelo plano. Em caso de empate, prevalece o plano com menor tempo total; se ainda houver empate, um critério de desempate determinístico baseado nos IDs dos testes decide o resultado — o mesmo change-set e catálogo sempre produzem o mesmo plano. O relatório rotula este método como `exact-maximum-risk-coverage`.

Essa busca exata é deliberadamente limitada para proteger a execução local: sua complexidade cresce exponencialmente com o número de candidatos. Catálogos mapeados maiores usam um método guloso determinístico, rotulado `deterministic-greedy-fallback` no relatório. Ele escolhe repetidamente, entre os testes que ainda cabem no tempo restante, o que tem a maior razão entre risco ainda não contemplado pelo plano e minutos estimados — desempatando por ID de teste. É uma aproximação transparente, não uma garantia de otimalidade.

O relatório torna explícitos os riscos adiados quando nenhum teste permitido está associado a eles, ou quando os testes associados a eles não cabem no orçamento restante. Deliberadamente, ele não calcula uma porcentagem de cobertura: um mapeamento planejado não deve ser apresentado como confiança ou como prontidão para release.

## Rastro de decisão

Para cada item do catálogo, o relatório registra um destes três estados de planejamento:

- `selected`: incluído no plano que coube no orçamento;
- `not-selected`: elegível, mas fora do plano porque o plano escolhido já contempla seus riscos mapeados, ou porque a regra de seleção documentada preferiu outro conjunto de testes;
- `excluded`: fora dos níveis de teste permitidos, ou sem uma correspondência explícita de palavra-chave de componente e tag de risco.

O rastro explica apenas uma decisão de planejamento. Ele não implica que um teste exista, tenha sido executado ou tenha passado.
