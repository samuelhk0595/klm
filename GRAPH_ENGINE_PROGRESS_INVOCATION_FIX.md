# Graph invocation — correção de contrato e defaults

Data: 2026-09-13. Implementado e compilado para validação humana.

## Arquivos desta entrega

- `engine/graph_invocation_contract.go`: DTOs das duas ferramentas, schemas aninhados, diagnóstico de argumentos, validação de referências da atividade e normalização de workspace.
- `engine/graph_tools.go`: publicação dos schemas, workspace opcional no invoke, validação antecipada de autorização e defaults/regras do retry.
- `engine/graph_records.go`: contrato legado documentado, validação persistida de autorização com erros acionáveis e diagnósticos de associações.
- `engine/store.go`: normalização no carregamento e antes da escrita do estado v2.
- `engine/graph_runtime.go`: normalização do próximo estado antes das invariantes e persistência.
- `engine/graph_scheduler.go`: normalização e validação antes do preflight/reserva, inclusive fresh/reuse quando existe tentativa anterior.
- `engine/graph_workspaces.go`: validação antes da criação, preservação do workspace inicial já associado e diagnósticos de reutilização.
- `engine/prompts/orchestrator.md`: exemplo mínimo correto, defaults do produto, referência de autorização da atividade e regras concisas de retry.
- `engine/graph_invocation_contract_test.go`: testes pequenos de decoder/default/schema, sem executar grafos.
- Este relatório; binário gerado em `engine/engine-next.exe`.

## Defaults exatos

- `graph_invoke` sem `workspace`, ou com `workspace: {}`, normaliza para `mode: "original"` antes de persistir a atividade.
- `workspace` presente sem `mode` também normaliza para `original`. Um retry com modo omitido usa `original`, não herda automaticamente `new_worktree` da tentativa anterior.
- `original` usa a pasta cadastrada do projeto (`Project.Folder`) com os arquivos existentes. Não aceita caminho arbitrário na ferramenta. Reutilização explícita continua usando os IDs/proveniência registrados.
- O agente não deve perguntar nem emitir mensagem justificando esse default.
- `workspace.authorizationEventId` é opcional para ambos os modos. Campo ausente ou string vazia legada é aceito; se não vazio, continua exigindo referência a uma mensagem user real da conversa.
- `attempt` não recebe default. Retry depois de um run exige `fresh` ou `reuse`, mesmo quando `mode` foi omitido. Uma correção de preflight sem run anterior pode omitir workspace/attempt. O retry continua exigindo correção concreta e tarefa autocontida.
- `reuse` continua exigindo `sourceRunId` e `reuse.initial`, além das associações explícitas necessárias para Fork/Join. Fonte, IDs, prontidão, repositório e proveniência seguem as validações existentes.
- Em `new_worktree`, revisão não especificada usa HEAD local da pasta de origem na criação; retry mantém a revisão anteriormente solicitada conforme a regra existente. Não há reset, commit, limpeza ou cópia implícita de arquivos modificados.
- Atualizações que não sejam retry preservam a decisão de workspace salva.

## Schema e erros

- `graph_invoke.authorization` publica array não vazio de objetos estritos `{eventId: string, text: string}`, ambos obrigatórios, com descrições de origem e escopo.
- `authorization` continua obrigatório para a **atividade**. O engine verifica presença, texto não vazio e referência a evento user na conversa proprietária; não compara semanticamente a mensagem com a tarefa nem cria grants de ferramentas/harness.
- `graph_invoke.workspace` e `graph_update_activity.workspace` compartilham schema completo de `mode`, `attempt`, `authorizationEventId`, `baseRevision`, `sourceRunId` e `reuse`.
- Objetos fechados rejeitam campos extras; `reuse` permite associações dinâmicas com valores string contendo IDs registrados. Enums de modo, tentativa e ação são publicados.
- O decoder estrito permanece ativo. Verificações específicas das duas ferramentas fornecem caminho completo e formato/campos aceitos, por exemplo `workspace.path`, `workspace.event_id`, `authorization[0].scope` e `workspace.reuse.initial`.
- Valores explícitos inválidos de modo/tentativa não viram default. `mode: ""`, `mode: null`, workspace nulo e tipos incompatíveis são rejeitados na fronteira da ferramenta.
- Autorização ausente, referência inexistente/não-user e texto inválido recebem erros acionáveis antes da gravação, em vez de uma invariável genérica de `state.json`.
- `graph_update_activity` mantém a autorização da atividade existente; não recebe um novo campo `authorization` nem passa a autorizar trabalho novo por `userEventId`.

## Compatibilidade e persistência

- Mantida a versão 2 do estado. Nenhuma migração destrutiva ou descarte de registros.
- Registros antigos com modo ausente/vazio recebem somente `original` no carregamento e antes da escrita. Modos explícitos válidos, referências legadas, decisões fresh/reuse, revisões, IDs e proveniência são preservados.
- Valores de modo inválidos não vazios continuam sendo recusados; a compatibilidade com modo vazio no armazenamento não torna `mode: ""` válido em novas chamadas.
- Scheduler valida a decisão antes de preparar/reservar um run e grava o workspace normalizado. A criação inicial valida o modo e respeita o ID inicial já associado.
- O estado real da engine em execução não foi aberto ou regravado por esta validação. Leitura/gravação v2 e execução integrada permanecem para validação humana após reinício.

## Verificação realizada

- `gofmt` nos arquivos Go tocados.
- Uma execução focada: `go -C engine test -run '^TestGraphInvocation(Decoder|Defaults|Schema)$' -count=1 .` — **passou**, `ok klm/engine 0.618s`.
- Os testes exercitam formatos/aliases incorretos, tipos e enums explícitos inválidos, autorização referenciada, default sem aprovação de workspace, preservação de metadados e exigência de fresh/reuse. Não fazem I/O de persistência, Git, harness nem execução de run.
- `Test-Path` confirmou o diretório pai `engine`.
- Build Go único: `go -C engine build -o engine-next.exe .` — **passou**, sem diagnósticos.
- Não foram executadas suítes completas, integrações complexas, modelos ou grafos reais. Nenhum commit/push.

## Entrega para operações

Binário corrigido:

`C:\Users\Samuel\Documents\Projects\Personal\klm\engine\engine-next.exe`

A engine em execução (`engine/engine.exe`) não foi parada nem substituída. O agente de operações deve usar o novo binário no reinício combinado, preservando os argumentos e data-dir existentes. O usuário recarregará a aba e validará o fluxo real, incluindo invoke sem workspace, ferramenta com schema atualizado, mensagens de erro, autorização e retries.
