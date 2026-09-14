# Graph engine — documentação da invocação

Data: 2026-09-13. Decisão atual aplicada aos documentos; implementação Go segue
em paralelo com seu responsável.

## Decisão registrada

- `workspace` omitido usa `original`, sem pergunta ou referência obrigatória à
  mensagem de escolha. O fundamento é o default do produto, não consentimento
  expresso inferido da omissão.
- Modos explícitos `original`/`new_worktree` continuam; original é a pasta
  registrada do projeto, sem path arbitrário. `workspace.authorizationEventId`
  não é obrigatório.
- `authorization` continua como referências `{eventId,text}` que fundamentam a
  atividade e apontam a evento `user` existente na conversa invocadora. Isso dá
  rastreabilidade, não prova semântica nem grant do harness.
- Schemas nested de authorization/workspace devem ser completos e coerentes com
  a validação estrita. O diagnóstico registrado foi a falta dessas descrições,
  levando o agente a adivinhar `path`, `scope` e IDs de eventos.
- Prompts devem aplicar defaults definidos e perguntar só por informação ou
  ambiguidade bloqueante. A decisão fresh/reuse do usuário em retries com artefatos,
  correção concreta, associações W1 e exceção W2 continuam vigentes.

## Arquivos alterados

- `GRAPH_AUTHORING_REFINEMENT.md`: nova seção 13 datada e nota de precedência;
  substituída a instrução vigente de perguntar pelo workspace omitido. Clarificado
  que fresh/reuse se resolve para tentativas com artefatos, sem reconfirmar decisão
  já fornecida. Marcos fechados permanecem como histórico.
- `GRAPH_ENGINE_IMPLEMENTATION_PLAN.md`: precedência da decisão posterior,
  contrato de autorização/schema/default, persistência do modo resolvido e roteiro
  de reteste humano. Status da elaboração identificado como histórico.
- `product.md`: regra de produto do workspace original por omissão, limites de
  autorização/rastreabilidade e dever de aplicar defaults.
- `CONTEXT.md`: referência de autorização da atividade, workspace de execução e
  decisão fresh/reuse definidos separadamente.
- `engine/README.md`: contrato datado de invocação e campos nested, distinguindo
  decisão aprovada de comprovação de implementação em execução.
- `GRAPH_ENGINE_PROGRESS_CORE.md`: apenas nota inicial de precedência/integração
  que substitui requisitos antigos e identifica o fechamento anterior como histórico;
  exemplos, contratos e evidências anteriores preservados.
- `README.md`: ajuste pontual de “agreed workspace” para o default aprovado,
  preservando fresh/reuse em novas tentativas com artefatos.
- `GRAPH_ENGINE_PROGRESS_INVOCATION_DOCS.md`: este registro.

## Checagem e próxima validação

Leitura dirigida de `AGENTS.md`, produto, contexto, refinamento, plano, READMEs e
relatórios CORE, DOCS e progresso principal; conferência textual das instruções
de workspace/autorização nos documentos do escopo. Alterações locais anteriores
preservadas por patches pontuais.

Nenhum código ou prompt foi editado nesta rodada. Nenhum teste, review, build,
typecheck, restart, execução de grafo/modelo ou commit foi realizado por este
subagente. Relatórios adapter/frontend e progresso principal não foram editados.

O implementador Go integra o default e schemas nested completos. O principal
coordena alinhamento dos prompts e restart; depois o usuário recarrega a aba e
retesta a invocação sem workspace, os modos explícitos e a preservação das decisões
de retry. Este relatório não afirma que o runtime novo já foi validado.
