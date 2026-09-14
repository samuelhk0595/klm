# Graph engine — documentação do limite de MCP externo

Data: 2026-09-13. Decisão aprovada aplicada à documentação; adapter implementa em
paralelo com seu responsável.

## Diagnóstico e decisão

A recusa `cannot confirm lifecycle ... agentdeck` vinha de uma regra preventiva
genérica sobre configuração MCP externa em OpenCode/Codex, sem chamadas pendentes
demonstradas. O contrato agora permite MCPs configurados, sem allowlist por nome
nem desabilitação. `agentdeck` não recebe uma exceção particular.

A engine garante lifecycle do nó e das chamadas de ferramentas: impedir novas
chamadas após selo e aguardar as efetivamente iniciadas antes de aceitar Choice.
Não garante encerrar servidor MCP compartilhado ou tarefa remota destacada após
o retorno. Resposta “tarefa iniciada” conclui a responsabilidade pela chamada,
sem comprovar término da tarefa ou sucesso do objetivo.

Erro/cancelamento ou callback perdido sem evidência de conclusão mantém finality
incerta. Encerrar processo local/cliente, esvaziar Job ou solicitar cancelamento
não comprova cancelamento remoto nem autoriza resultado normal fictício.
Grants, sandbox e aprovações permanecem iguais.

## Arquivos alterados

- `GRAPH_AUTHORING_REFINEMENT.md`: nova seção 14 datada com diagnóstico e
  precedência; delimitação vigente de Choice e J7 ao nó/chamadas.
- `GRAPH_ENGINE_IMPLEMENTATION_PLAN.md`: nota de precedência, sequência de
  selo/conclusão, limite externo e tratamento de incerteza.
- `product.md`: substituição do ban preventivo pelo limite aprovado e precisão
  do requisito de aceitação de Choice.
- `CONTEXT.md`: termos de finality de nó/chamada, tarefa externa destacada e
  finality incerta, distinguindo retorno da ferramenta de conclusão da tarefa.
- `README.md`, `engine/README.md`, `clients/desktop/README.md`: alinhamento
  pontual do ban/limite de lifecycle; contrato separado de comprovação runtime.
- `GRAPH_ENGINE_PROGRESS_ADAPTER.md`, `GRAPH_ENGINE_PROGRESS_CORE.md`: somente
  pequenas notas no topo indicando precedência sobre os requisitos antigos.
  Evidências e marcos históricos preservados.
- `GRAPH_ENGINE_PROGRESS_EXTERNAL_MCP_DOCS.md`: este relatório.

Default `workspace` omitido → `original`, autorização da atividade, schemas nested,
fresh/reuse e demais correções recentes preservados. Código, prompts e progresso
principal não foram editados por este subagente.

## Checagem e coordenação

Checagem exclusivamente textual por leitura e busca dirigida das instruções de
MCP/lifecycle nos documentos do escopo. Nenhum teste, review, build, typecheck,
restart, execução de modelo/grafo ou commit realizado nesta rodada documental.
Evidências anteriores nos relatórios não validam a mudança atual.

O responsável pelo adapter implementa em paralelo. O usuário pediu restart ao final:
o principal o coordena após integração e build. Depois, o usuário recarrega a aba e
retesta com MCPs externos configurados. A validação deve distinguir chamadas que
retornaram (inclusive “tarefa iniciada”) de chamadas sem conclusão comprovada,
preservando selo, Choice, finality incerta e permissões existentes. Nenhum restart
ou comportamento novo é declarado realizado/validado por este relatório.
