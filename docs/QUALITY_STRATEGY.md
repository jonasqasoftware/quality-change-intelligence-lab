# Estratégia de Qualidade

## Contexto

Diante de uma mudança, times costumam reagir de duas formas: rodar todo o conjunto de testes disponível, ou confiar na intuição de quem revisa. As duas escondem a decisão: a primeira pode ser lenta e ruidosa; a segunda pode deixar passar um risco de alto impacto sem que ninguém consiga questionar por quê.

Quality Change Intelligence Lab é uma ferramenta local que transforma uma descrição de mudança e um catálogo de testes declarados em uma recomendação determinística e revisável. Os fixtures incluídos são sintéticos. Eles não são um registro de uma mudança real de produto, um gate de release, nem evidência de teste em produção.

## Riscos principais

| Risco | Por que importa | Mitigação nesta ferramenta |
| --- | --- | --- |
| Regras de priorização ocultas | Um time não consegue contestar uma recomendação de caixa-preta. | Os pesos de criticidade e tipo de mudança são visíveis no código e documentados em [Modelo de Decisão](DECISION_MODEL.md). |
| Mapeamento incorreto no catálogo | Um teste pode ser selecionado para o componente ou risco errado. | A correspondência exige palavra-chave de componente **e** tag de risco explícita ao mesmo tempo. |
| Orçamento escondendo risco relevante | Uma janela de tempo curta pode adiar uma verificação importante. | Para catálogos pequenos, a seleção é exata (maximiza o risco contemplado pelo plano); riscos adiados são sempre listados com o motivo. |
| Omissões opacas do catálogo | Quem revisa não consegue saber por que um teste disponível não foi proposto. | O relatório registra uma decisão para cada item do catálogo: selecionado, não selecionado, ou excluído. |
| Planejamento confundido com prova | Um teste selecionado é confundido com um resultado aprovado. | Todo relatório declara seu limite de evidência e o risco residual. |

## Estratégia e níveis de teste

| Nível | Propósito | Evidência produzida |
| --- | --- | --- |
| Unitário | Validar peso de risco, correspondência, seleção por combinação exata, orçamento e comportamento do fallback. | Resultados determinísticos dos testes em Go. |
| Integração | Carregar entrada JSON estrita e renderizar relatórios em Markdown/JSON pelo CLI. | Saída do CLI a partir dos fixtures sintéticos. |
| Workflow | Formatar, vetar, testar com detecção de race condition, compilar e gerar um relatório de exemplo como artefato de CI. | Artefato do GitHub Actions vinculado a uma execução do workflow. |

## Critérios de decisão

Uma recomendação só é selecionada quando ela:

1. pertence a um nível de teste permitido;
2. corresponde tanto ao componente alterado quanto à tag de risco declarada;
3. adiciona risco ainda não contemplado pelo plano; e
4. cabe no orçamento de tempo fornecido.

O algoritmo de seleção — limite de candidatos para busca exata, regra de desempate e fallback guloso para catálogos maiores — está detalhado em [Modelo de Decisão](DECISION_MODEL.md). O modelo de pontuação é uma heurística transparente para apoiar discussão, não uma fórmula universal de release.

## Critérios de sucesso da estratégia

- Toda recomendação selecionada corresponde a um risco mapeado por palavra-chave de componente e tag de risco — verificável pelos testes unitários de correspondência.
- O mesmo change-set e catálogo sempre produzem o mesmo plano — determinismo verificável por execução repetida.
- Todo item do catálogo aparece no relatório com um status e um motivo (`selected`, `not-selected` ou `excluded`); nenhum item fica sem rastro.
- Todo risco não contemplado pelo plano aparece na lista de riscos adiados, com o motivo do adiamento.
- Os quality gates (`gofmt`, `go vet`, `go test -race`, `go build`) passam antes de qualquer alteração ser considerada íntegra.

## Riscos residuais e limites

- As entradas são declaradas manualmente; esta ferramenta não interpreta um diff real do Git nem se conecta a rastreadores de issues.
- O catálogo é um modelo de planejamento. Ele não prova implementação, execução ou aprovação de um teste.
- Estimativas, palavras-chave de componente, criticidade e tags de risco exigem revisão humana.
- Um relatório é insumo de planejamento para uma decisão de qualidade, não a decisão em si; deve ser interpretado junto com resultados de execução, contexto de produto e incidentes conhecidos, nunca isoladamente.
