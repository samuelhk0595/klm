# Graph engine — CORE P1–P5

## Precedência documental — MCP externo, 2026-09-13

A seção 14 de `GRAPH_AUTHORING_REFINEMENT.md` substitui as recusas preventivas de
MCP externo citadas no histórico abaixo. Permitir MCPs configurados em OpenCode/Codex
sem allowlist por nome/desativação. O lifecycle garantido é do nó e das chamadas:
selar contra novas chamadas e aguardar as iniciadas antes da aceitação. Servidores
compartilhados e tarefas destacadas após a resposta ficam fora da garantia;
“tarefa iniciada” conclui a chamada, não a tarefa. Erro/cancelamento ou callback
perdido sem conclusão comprovada preserva finality incerta; não fabricar resultado
normal nem cancelamento remoto. Grants/sandbox e a decisão de workspace abaixo
permanecem vigentes. Esta nota não comprova implementação, build, teste ou restart.
Detalhes: `GRAPH_ENGINE_PROGRESS_EXTERNAL_MCP_DOCS.md`; o principal coordena o
restart solicitado após build, seguido do reload e reteste pelo usuário.

## Nota de precedência e integração — invocação, 2026-09-13

A decisão posterior do usuário, registrada na seção 13 de
`GRAPH_AUTHORING_REFINEMENT.md`, substitui os requisitos antigos de workspace
citados neste relatório, inclusive o exemplo de `graph_invoke`, o contrato de
retry e a afirmação de que a escolha precisa de `authorizationEventId` próprio:

- `workspace` omitido resolve para `original`, sem pergunta nem referência
  obrigatória à mensagem da escolha. `original` é a pasta registrada do projeto,
  sem path arbitrário. Modos explícitos `original`/`new_worktree` continuam.
- A regra padrão do produto fundamenta essa resolução; omissão não é consentimento
  expresso. `workspace.authorizationEventId` não é obrigatório.
- `authorization` mantém referências `{eventId,text}` para fundamentar a atividade
  e validar evento existente do tipo `user` na conversa invocadora. Rastreabilidade
  não é prova semântica de consentimento nem grant de permissão do harness.
- O modo resolvido continua explícito na run persistida; omissão na entrada não
  significa ausência de modo no registro. O implementador Go integra o default e
  schemas nested completos de authorization/workspace no inventário das ferramentas,
  eliminando a necessidade de adivinhar `path`, `scope` ou IDs de eventos.
- Instruções do orquestrador devem aplicar defaults definidos e perguntar só por
  ambiguidade/informação bloqueante. Retry com artefatos mantém correção concreta,
  decisão fresh/reuse do usuário e associações explícitas W1; W2 permanece vigente.

Esta é somente uma nota documental de integração. As evidências de build e os
marcos abaixo pertencem às rodadas anteriores, não validam esta alteração.
Implementação Go e alinhamento dos prompts seguem com seus responsáveis; restart
será coordenado pelo principal, seguido de reload da aba e reteste pelo usuário.
Mudanças documentais desta rodada: `GRAPH_ENGINE_PROGRESS_INVOCATION_DOCS.md`.

## Fechamento anterior — histórico preservado

Atualizado: 2026-09-13. P1–P5 implementados; integração core/adapter fechada nesta rodada. **`go -C engine build` concluído com sucesso**, sem diagnósticos. A pendência anterior de `graphConversationHooks` foi resolvida pelo adapter.
Nenhum modelo, tarefa real, suíte, commit ou review executado pelo core.

## Fechamento de integração — estado vigente

### Correções e contratos confirmados

- Factory publicada em `graph_adapter.go` com a assinatura acordada e registrada pelo `init()` de `graph_tools.go`.
- `execute` inicializa `t.prompt` com o texto recebido, chama a factory em main/side sem binding e usa o texto completo enriquecido antes de catálogo/bridge/request nativo. Binding privado de nó continua obrigatório.
- Corrigido `graph_scheduler.go`: a notificação pré-bindada agora inicializa **`t.prompt = prompt` com `run-result.md` + dados**, antes de a factory adicionar `orchestrator.md`. Assim, o `payload.Text = t.prompt` de `execute` conserva política e resultado; antes, o prefixo substituía o conteúdo da notificação.
- `Finished` é chamado após persistir o fim da sessão, remover `a.runs`, fechar `t.done` e liberar `app.mu`. O callback dos nós comunica o resultado ao supervisor; nenhuma transição de aceitação despacha destino diretamente.
- O adapter cobre erro anterior à construção dos hooks com fallback para `graphNotificationFinished`; com hooks instalados, usa `Finished`, sem chamar ambos.
- Corrigido o outbox para decidir delivered/pending pelo `GraphAdapterResult` do turn encerrado, sem consultar `Session.Status`, que pode já pertencer a outro turn reservado pelo árbitro. `finishGraphAdapter` já classifica erro, falta de conclusão e interrupção explicitamente.
- Na falha de preparação de uma notificação, o core atualiza o outbox **antes** de liberar seu `WaitGroup`, preservando a ordem com shutdown.
- Alterações Go deste fechamento restritas a `engine/graph_scheduler.go` e comentário do contrato em `engine/main.go`; mudanças anteriores do adapter/frontend preservadas.

### Prompts no modo de desenvolvimento

- Confirmada a presença dos seis arquivos em `engine/prompts/` por listagem de arquivos.
- Loader usa primeiro o diretório do fonte identificado por `runtime.Caller`, depois `prompts/` adjacente ao executável. Para o build atual, `engine/engine.exe` tem `engine/prompts/` como recurso adjacente real; o lookup padrão não depende do cwd do servidor.
- `init()` só registra a factory: não exige leitura de prompt no startup do pacote. Os arquivos continuam relidos quando um envio é preparado.
- `engine/README.md` documenta o modo dev e override explícito, a partir da raiz: `$env:KLM_PROMPTS_DIR = (Resolve-Path .\engine\prompts).Path`. Um override explicitamente inválido continua gerando erro claro, sem fallback silencioso.
- Conferência por código/arquivos; **nenhum servidor foi iniciado ou reiniciado** para validar isso.

### Evidência deste fechamento

1. `gofmt -w "engine/graph_scheduler.go" "engine/main.go"` — sucesso, sem diagnósticos.
2. `Test-Path -LiteralPath "engine"` — `True`.
3. `go -C engine build` — **sucesso, saída vazia, exit code 0**. Uma execução nesta rodada; nenhuma checagem sobreposta ou suíte.

### Limites vigentes e passagem para operações

- Sem bloqueio de compilação ou símbolo de integração pendente. Operações pode prosseguir com o restart combinado; o core não reiniciou servidores.
- Conforme o topo vigente de `GRAPH_ENGINE_PROGRESS_ADAPTER.md`, os três harnesses Windows são admitidos pelo preflight de mecanismos: Pi (gate/extensão), OpenCode (plugin owned) e Codex (selo físico + leitura fria). **`RuntimeValidated` continua false**; build não comprova Choice/continuação nativa com modelo.
- OpenCode exige a versão inspecionada **1.18.30**, handshake e inventário compatíveis. OpenCode/Codex recusam MCP externo incompatível antes do prompt. Configurações que não atendam a essas condições são bloqueios concretos de execução, não falha de compilação.
- Unix permanece recusado para nós por falta de contenção equivalente. Job Object não comprova término de serviços remotos/daemons fora do domínio owned; essas limitações não foram removidas pelo core.
- Validação humana ainda necessária para Choice+continuação, Terminal/workspaces, Fork/Join, solicitações, notificações e restart. Nenhum modelo/grafo/tarefa real foi executado nesta integração.

## Segunda rodada — implementação P2–P5

### P2 — runtime sequencial, sessões e resultado final

- `graph_runtime.go`: supervisor por run, contexto independente do turn, registro antes de efeitos, workers de nós e cancelamento coordenado. `starting/running/ending` retêm o slot por conversa.
- `BindGraphAdapter` é chamado antes de `execute` em sessões `Role:"graph_node"`; callbacks capturam run/activation/turn IDs, não confiam em IDs enviados pelo modelo.
- `ReserveChoice` valida envelope/campos/origem, persiste reserva, devolve erro corretivo lido de `invalid-choice.md` e conta violações por call ID. `Transition` persiste sealed/drained/accepted em ordem, resolve mapping e só então aceita. **Nenhum desses callbacks despacha destino.**
- `Finished` envia um resultado ao supervisor depois de a sessão nativa encerrar. O supervisor marca continuação segura e despacha a conexão real, ou acorda a mesma ativação com `missing-choice.md`. Invalid/missing somam até três; callback repetido é deduplicado por native turn ID.
- Sessão mais recente do nó/run só é reutilizada com `continue_target`, mesmo cwd canônico/harness e `Continuable` confirmado pelo adapter. Correção mantém ativação/sessão; nova run nunca importa a sessão nativa anterior.
- Solicitações humanas existentes entram na projeção da conversa com a origem privada correta. Grants continuam no escopo original do callback.
- Foram criados os seis prompts aprovados em `engine/prompts/`; loader relê a cada envio, com `KLM_PROMPTS_DIR`/recursos adjacentes como opção de distribuição.

### P3 — atividades, fila, assessment e outbox

- `graph_tools.go`: `init()` registra **`graphConversationHooks = graphConversationFactory`**, usando exatamente a assinatura acordada `func(*app,*turn,Session)(*GraphAdapterHooks,error)`.
- Core acrescentou `turn.prompt string` e `turn.graphNotificationID string`. Factory retorna `Node:false`, inventário/handlers e acrescenta o prefixo ao texto completo de `t.prompt`. O consumo pelo adapter e o caminho pré-bindado foram concluídos no fechamento registrado acima.
- Main recebe `graph_catalog`, `graph_activities`, `graph_get_run`, `graph_recent_events`, `graph_invoke`, `graph_assess`, `graph_update_activity`. Side recebe só as quatro consultas, com ownership da conversa principal. Nós recebem exclusivamente a capacidade Choice do binding do adapter.
- Autorizações referenciam eventos `user` existentes. Decisão de workspace tem `authorizationEventId` próprio, podendo apontar à mesma mensagem já explícita. IDs dão rastreabilidade; política semântica permanece no prompt do orquestrador.
- Operações são versionadas/idempotentes; prioridade: correção concreta pronta, prioridade expressa, ordem da solicitação. Dependentes aguardam `succeeded` por assessment, não apenas terminal completion. Atualizar dependências/abandonar/repriorizar exige evento de orientação do usuário. Não há Stop do orquestrador.
- Retry guarda nova task/correction e decisão fresh/reuse; uma revisão-base expressa já existente é preservada se a nova tentativa em worktree a omitir. Falha de preflight sem run não exige uma decisão extra sobre artefatos que nunca foram criados.
- `graph_scheduler.go` reserva captura por conversa, captura definições vigentes e faz preflight de todos os agentes/plataforma. Reconfere as revisões sob `authoringMu` antes de persistir o início e mudar a seleção atomicamente.
- `linked.go::scheduleLinkedLocked` é o árbitro compartilhado: notifications e consultas linked usam o mesmo `a.runs[sessionId]` dos turns de usuário. Nenhuma sobreposição nativa. Preflight inválido gera activity-error persistente; terminal gera outbox na mesma transação da liberação do slot.
- Notificação incerta/falha volta a pending com o mesmo ID e retry de transporte após 30s; não reinicia tarefas. Estado/assessment/operation IDs impedem efeitos duplicados. Restart preserva outbox/fila e interrompe trabalho incerto sem replay.

### P4 — Terminal e workspaces

- `graph_terminal.go`: PowerShell não interativo, script único, cwd herdado e `$payload` via JSON UTF-8 privado em `<data-dir>/graph-inputs/<run>/<activation>/payload.json`.
- stdout/stderr em ordem observada, coleta máxima 64 KiB UTF-8 com flag de truncamento, pipes drenados. Exit não zero segue o mapping; launch/collection/infrastructure failure é falha técnica. Não há timeout de tarefa.
- Shell e comandos Git passam por `PrepareGraphProcess`, `Attach`, `Stop`, `Close` e `Drained`. Incerteza de drenagem é propagada como `graphUnconfirmedError`, retém `ending` e registra `FinalityError`; não libera o slot fingindo término.
- `graph_workspaces.go`: intents persistidos antes de `git worktree add`, argumentos separados, nomes únicos com run/activation/branch e diretórios em `<data-dir>/worktrees/<project>/<workspaceId>`.
- Modo original aceita arquivos/dirty state existentes; new_worktree usa revisão pedida ou HEAD local da origem. Nenhum reset/commit/push/pull/checkout do diretório original/cópia dirty/cleanup implícito.
- Reuse exige source run terminal, mapa completo explícito incluindo initial, IDs usados pelo source run, existência física e repositório/branch managed compatíveis. Associações são validadas contra o snapshot inteiro, inclusive caminhos não visitados. Proveniência e usos posteriores permanecem registrados.

### P5 — Fork, Join, causalidade e deadlock

- Fork calcula uma única revisão-base para suas criações isoladas, persiste-a e prepara os ramos. Ramo compartilhado mantém exatamente o workspace herdado e ignora o nome Git configurado. Outputs dos ramos são resolvidos explicitamente; não há duplicação implícita de input.
- Supervisor inicia os nós reservados dos ramos em paralelo. Entregas são únicas por conexão/produtor; Choice compartilhada não se transforma em conexões artificiais.
- Join coleta todas as conexões, preserva payloads separados e aceita `{}` como chegada. Correlaciona por segmento causal e rodada; provenance de Fork fica em campo separado, sem exigir igualdade de Fork ID.
- Rodadas sobrepostas/chegadas independentes duplicadas/origens ambíguas/policies incompatíveis são erros explícitos. O agente só inicia após todas as chegadas. Terminal não vota policy; default é new.
- Join sem branch retorna à origem; com branch cria nova integração do HEAD local atual dessa origem. Prompt contém inputs separados e metadata Git/workspaces atual. Integração/conflitos são trabalho do agente; não há merge automático.
- Após Join, contexto/cwd seguem pela integração e uma nova identidade causal distingue o próximo ciclo. Ciclos ordinários seguem sem limite artificial. Fork aninhado continua explicitamente rejeitado.
- Quando restam apenas entradas faltantes sem produtor ativo/reservado, falha com conexões faltantes. Agente esperando pergunta/permissão mantém worker ativo e não é considerado deadlock.
- Terminal autoral normal exige reconvergência e ausência de trabalho pendente. Engine blocked/falha encerra o restante, preserva artefatos e só publica após settlement dos workers.

### Arquivos acrescentados/alterados nesta rodada

- Novos: `engine/graph_runtime.go`, `graph_scheduler.go`, `graph_tools.go`, `graph_prompts.go`, `graph_terminal.go`, `graph_workspaces.go` e os seis `engine/prompts/*.md`.
- Integração: `engine/main.go`, `engine/graph_records.go`, `engine/linked.go`, `engine/README.md` e este relatório. `graph_adapter.go`, `harness.go`, bridge/adapters/platform e frontend não foram editados pelo core.
- A base P1 abaixo continua sendo o contrato HTTP/frontend; não foram introduzidas abas/histórico/painéis de execução novos.

### Pendências da segunda rodada — situação após o fechamento

- A ausência anterior de `graphConversationHooks` foi resolvida; o build integrado agora passou. O uso de `turn.prompt` nas notificações pré-bindadas foi corrigido sem definições duplicadas ou stubs.
- A recusa de OpenCode/Codex mencionada na leitura inicial era o estado histórico P0. O estado vigente admite os três mecanismos Windows com as condições descritas no topo deste relatório; Unix e as limitações de serviços remotos permanecem.
- Gofmt/parsing aplicado aos arquivos core, sem erros. Nenhum grafo/modelo/processo de tarefa foi executado como verificação. Falta validação humana nativa de Choice+continuação, Terminal/Git/workspaces, Fork/Join e notificações/restart.

### Comandos/evidência da segunda rodada — histórico preservado

1. `gofmt -w` nos arquivos Go de propriedade do core — concluído sem diagnóstico de parsing. Nenhum arquivo de adapter foi formatado pelo core.
2. `Test-Path -LiteralPath "engine"` — `True`.
3. **Único build:** `go -C engine build` — falhou com:

```text
# klm/engine
.\graph_tools.go:13:15: undefined: graphConversationHooks
```

O compilador não reportou outros erros naquela execução. O build daquela rodada não foi aprovado e não foi repetido enquanto o adapter editava seus arquivos. **Esse bloqueio foi superado pelo build bem-sucedido do fechamento de integração descrito no topo**, após nova autorização do usuário e publicação do hook.

### Ferramentas de mutação — argumentos concretos

`graph_invoke`:

```json
{
  "operationId": "stable-client-operation-id",
  "graphId": "saved-graph-slug",
  "objective": "Authorized objective",
  "task": "Self-contained run task",
  "authorization": [{"eventId": "existing-user-event-id", "text": "Expressly authorized scope"}],
  "workspace": {
    "mode": "original",
    "authorizationEventId": "existing-user-event-id"
  },
  "dependencies": [],
  "priority": 0
}
```

Resposta real: `{activity, started, runId?, runStatus?}`. Sem slot/dependências prontas, permanece atividade em fila; `started` só é true se houver run ativa persistida. Erro de preflight é erro de ferramenta e também diagnóstico de atividade persistido, nunca uma run fictícia.

`graph_assess`: `{activityId, runId, operationId, expectedVersion, satisfied:boolean, reason}`. `graph_update_activity`: `{activityId, operationId, expectedVersion, action, task?, correction?, workspace?, userEventId?, priority?, dependencies?}`. Actions: `retry`, `await_user`, `abandon`, `reprioritize`, `change_dependencies`.

`workspace` de retry mantém o contrato P1: `mode`, `attempt:"fresh"|"reuse"`, `authorizationEventId`, `baseRevision?`, `sourceRunId?`, `reuse?:{association:workspaceId}`. Reuse requer `initial`. Revisão explícita diferente pode ser enviada; omissão preserva a base expressa anterior para nova worktree, ou usa HEAD quando nunca houve base expressa.

Consulta `graph_get_run` expõe run/estado/resultado, ativações, rounds/deliveries e workspaces/uses; `graph_recent_events` aceita `{runId,nodeId,limit?}` e limita a 50 eventos (15 por omissão), sem payloads volumosos dos eventos. IDs arbitrários de outra conversa são recusados.

---

## Contrato P1 preservado (referência de integração)

## Contratos disponíveis agora para frontend

### Catálogo e autoria

- `GET /api/projects/{projectId}/authoring`: contrato existente `{agents, graphs, errors}`. Os `graphs` são `GraphRecord` (`id`, `revision`, `definition`, `layout`). Erros de catálogo não equivalem a lista vazia.
- `GraphNode.output?: Record<string,string>`: somente Terminal; ausência significa `{}` e não encaminha input.
- `GraphBranch.separate_worktree?: boolean`: ausência **true**, inclusive YAML antigo. `false` preserva `git_branch` configurado, mas não exige nome válido/não vazio nem o utiliza para checkout. `GraphBranch.Isolated() bool` aplica o default.
- Fork output aceita `{{payload.campo}}` e `{{run.input.task}}`. Terminal output também aceita `{{command.result}}`. Choice output aceita `{{choice.campo}}` e `{{run.input.task}}`.
- Rename/delete de grafo atualiza seleção persistida por outbox recuperável entre os arquivos de autoria e `state.json`. Snapshot em execução permanece intacto.

### Seleção e projeção

`GET /api/sessions/{sessionId}/graph` retorna `ConversationGraphState`.

`PATCH /api/sessions/{sessionId}/graph` recebe **exatamente**:

```json
{"selectedGraphId":"implementation-review"}
```

None usa `{"selectedGraphId":""}`; `null`/campo ausente são 400. Resposta 200 é a projeção inteira. Só conversa principal; grafo deve existir e estar enabled para uma nova seleção. Selecionar None/abrir Graph não executa nem cancela trabalho.

```ts
type ConversationGraphState = {
  sessionId: string
  selectedGraphId: string // "" = None; sempre presente nesta projeção
  revision: number // monotônica global persistida; comparar antes de aplicar
  run: GraphRunProjection | null
  requests: GraphRequestProjection[]
}
type GraphRunProjection = {
  id: string
  graphId: string // identidade atual de seleção/catálogo; pode ser "" após delete
  active: boolean
  status: 'starting' | 'running' | 'ending'
  revision: number
  snapshot: GraphRecord // DTO de autoria já existente; definição/layout capturados
  activeNodeIds: string[]
  completedNodeIds: string[]
  completedChoiceIds: string[] // IDs crus; prefixar choice: apenas no canvas
  collectingJoinIds: string[] // Join coletando entradas, não agente executando
}
type GraphRequestProjection = {
  runId: string
  graphId: string
  nodeId: string
  nodeName: string
  activationId: string
  sessionId: string // usar ESTE ID na resposta do cartão
  requestId: string
  kind: 'permission' | 'question'
  permission?: Permission // contrato existente integral
  question?: QuestionRequest // contrato existente integral
}
```

- `GET /api/state`: sessões públicas incluem `selectedGraphId?: string` e `graph?: ConversationGraphState` nas conversas principais. Sessões privadas de nós são excluídas.
- `GET /api/sessions/{sessionId}/events`: SSE continua retornando **Session**, agora com o mesmo campo `graph`; todos os commits já notificam os listeners, inclusive mudanças de nós enquanto chat está idle. Fazer a inscrição também quando `session.graph.run?.active`.
- `GET .../graph` permite hidratação/polling pequeno independente dos eventos de chat.
- Respostas de criação/patch/message/stop em `api.go` também trazem a view. Endpoints legados de settings/side em outros arquivos podem retornar Session sem `graph`: manter a última projeção conhecida quando o campo estiver ausente; só `graph.run:null` explícito encerra a projeção.
- `run` é somente a run ativa, independentemente da seleção. None conserva `run` e `requests`; esconder canvas não esconde solicitações.
- Ao terminar, `run:null`; não manter LED ou inventar completed global. Active prevalece sobre completed de uma ativação anterior.
- Durante rename, `run.graphId` acompanha a identidade atual; `run.snapshot.id` conserva o ID capturado. Desenhar `snapshot` quando selectedGraphId corresponde a run.graphId. Layout ausente/inválido é `{positions:{}}` para posicionamento determinístico no cliente.
- Requests retornam pelos endpoints existentes: `POST /api/sessions/{request.sessionId}/permissions/{requestId}` e `POST /api/sessions/{request.sessionId}/questions/{requestId}/reply`; corpo existente dos cartões. Não são mensagens de chat.
- `GET /api/graph-runs/{runId}` retorna resumo real abaixo ou 404. Nenhuma rota HTTP inicia execução nesta fase; invocação continuará exclusiva das ferramentas do orquestrador.

```ts
type GraphRunSummary = {
  id: string; activityId: string; conversationId: string; graphId: string
  status: 'starting' | 'running' | 'ending' | 'completed' | 'blocked' | 'failed' | 'interrupted'
  active: boolean; revision: number
  result: null | {
    kind: 'completed' | 'blocked' | 'failed' | 'interrupted'
    choice?: {origin: 'graph' | 'engine'; id: string}
    output?: Record<string,string>; error?: string
  }
  createdAt: string; endedAt?: string
}
```

## Assinaturas exatas para runtime e adapter

```go
func compileGraph(g GraphDefinition, agents []AgentRecord,
    validateModel func(EffectiveGraphAgent) error) (*CompiledGraph, error)
func (a *app) captureGraphSnapshot(ctx context.Context,
    projectID, graphID string) (GraphSnapshot, *CompiledGraph, error)
func compileGraphSnapshot(snapshot GraphSnapshot) (*CompiledGraph, error)
func validateGraphSnapshot(snapshot GraphSnapshot) error
func compileOutput(output map[string]string, kind string,
    input map[string]ChoiceField) (CompiledOutput, error)
func resolveGraphOutput(output CompiledOutput,
    ctx GraphOutputContext) (map[string]string, error)
func (c *CompiledGraph) availableChoices(nodeID string) map[ChoiceIdentity]GraphChoice
func (c *CompiledGraph) validateChoice(nodeID string, identity ChoiceIdentity,
    raw map[string]any) (GraphChoice, map[string]string, error)
func (a *app) graphSlotAvailableLocked(conversationID string) error
func graphRunActive(status string) bool
func graphActivationActive(status string) bool
func validateGraphRecords(d *diskState) error
func interruptGraphState(d *diskState, cause string) bool
func (d *diskState) graphProjection(conversationID string) ConversationGraphState
func (d *diskState) sessionView(id string) *Session
func (d *diskState) graphActivity(id string) *GraphActivity
func (d *diskState) graphRun(id string) *GraphRun
func (d *diskState) graphActivation(id string) *GraphActivation
func (d *diskState) graphWorkspace(id string) *WorkspaceRecord
func (d *diskState) activeGraphRun(conversationID string) *GraphRun
```

- `captureGraphSnapshot`: chamar **fora** de `app.mu`; utiliza `authoringMu`, captura YAML/TOMLs e SHA-256, resolve overrides, valida catálogo real, relê todas as fontes e rejeita mudança observada. Não promete transação de editor externo sobre vários arquivos.
- `compileGraph`: valida execução adicional ao CRUD; callback de catálogo obrigatório com agentes. Recompilar snapshot usa configurações já capturadas, sem consultar catálogo atual.
- `CompiledGraph`: campos `Definition GraphDefinition`, `Agents map[string]EffectiveGraphAgent` por node ID, `Connections map[string]GraphConnection`, `Incoming/Outgoing map[string][]string`, `Outputs map[string]CompiledOutput`.
- `EffectiveGraphAgent`: `AgentID, Harness, Model, Effort, Prompt string`.
- `GraphConnection`: `ID, Kind string`, `Sources []string`, `ChoiceID, BranchID, To, Session string`.
- IDs executáveis: `choice:<choiceId>`, `terminal:<nodeId>`, `fork:<nodeId>:<branchId>`. Choice compartilhada tem **uma** conexão com várias Sources. Incoming é ordenado e Join aguarda todas.
- `GraphOutputContext`: `Task string`, `Choice/Payload map[string]string`, `CommandResult *string`, `ChoiceInput map[string]ChoiceField`. Campo Choice opcional declarado e omitido resolve vazio; payload ausente é erro técnico. Não há substituição no script PowerShell nem referências globais de outros nós.
- `ChoiceIdentity{Origin:"engine", ID:"blocked"}` sempre disponível para agent/join, contrato `reason:string` obrigatório. Distinto de `ChoiceIdentity{Origin:"graph", ID:"blocked"}`. Map não deve ser serializado diretamente em JSON; converter para lista/envelope ao expor ferramentas.

### Ownership e lifecycle

```go
// Novos campos Session persistidos (sem usar ParentID nem Workspace visual):
Role string              // "graph_node" para nó; legado main/side fica como está
GraphRunID string
GraphNodeID string
ExecutionCWD string      // absoluto/canônico, fornecido ao execute
SelectedGraphID string   // conversa principal; "" = None
Graph *ConversationGraphState // SOMENTE view HTTP/SSE, nunca persistir

// app:
graphRuns map[string]*graphExecution // chave run ID
// graphExecution:
ctx context.Context
cancel context.CancelFunc
done chan struct{}
```

- `app.runs map[string]*turn` continua sendo o transporte dos turns nativos (chat ou sessão privada). `app.graphRuns` é lifetime independente, derivado de `app.ctx`, não do turn invocador.
- Persistir `GraphRun{Status:"starting", Workspace:decisaoDaInvocacao}` e atividade `running` atomicamente antes de iniciar I/O; reservar slot por estado persistido. `Workspace.Mode` é obrigatório em cada run (original/new_worktree), separado da decisão mutável da atividade para tentativas futuras. Seleção no início real deve entrar no mesmo commit.
- P0 não precisa editar essas estruturas. Adapter deve consumir `Session.GraphRunID/GraphNodeID/ExecutionCWD` e receber a associação da ativação por contrato combinado no P2; uma sessão pode ser reutilizada por ativações diferentes.
- Não dar linked/chat-history/invoke/stop tools a `Role:"graph_node"`. Rotas públicas de messages/stop/side/settings/events já recusam sessões privadas; callbacks permission/question continuam acessíveis à UI.
- `GraphChoiceSubmission` possui `OperationID, TurnID string`, `Choice ChoiceIdentity`, `Payload map[string]string`, `ReceivedAt, AcceptedAt string`. Reservar em ativação `sealing`; **AcceptedAt só após P0 confirmar término**. Reserva não despacha destino.
- `Violations int` e `ViolationIDs []string` pertencem à ativação; IDs de call e omission distintos. O validador exige unicidade e tamanho igual ao contador (0–3). Correção/contagem/settlement ainda pertencem ao runtime P2.
- Armazenamento inválido/erro de persistência cancela turns e contextos graph, deixando engine read-only. I/O Git/harness nunca sob `app.mu`.

## Registros persistentes v2

Definições completas com tags JSON em `engine/graph_records.go`:

| Slice diskState | Tipo | Relações e estado |
| --- | --- | --- |
| GraphActivities | GraphActivity | conversa/projeto/grafo, autorização por evento user, objetivo, ordem/prioridade, dependências, assessment por run, correção, decisão fresh/reuse, versão/operation IDs |
| GraphRuns | GraphRun | activity, task, GraphSnapshot, slot starting/running/ending, result, workspace inicial, timestamps; CatalogGraphID é alias separado do GraphID capturado |
| GraphActivations | GraphActivation | run/node/occurrence, input/output, workspace/session, causal/fork/branch/Join round, violations, submission reservada/aceita, metadados Terminal |
| GraphDeliveries | GraphDelivery | produtor/conexão/destinatário, payload separado, workspace/origem/causalidade, target activation/Join round; pending/consumed/interrupted |
| GraphJoinRounds | JoinRound | entradas esperadas, `[]JoinArrival{ConnectionID,DeliveryID}`, causalidade, origem/integração, policy; collecting/running/completed/interrupted |
| GraphWorkspaces | WorkspaceRecord | recurso físico e origem criadora, operação única; reserved/ready/failed/interrupted |
| GraphWorkspaceUses | WorkspaceUse | associações explícitas recurso/run/activation e reutilização |
| GraphNotifications | RunNotification | outbox persistido, ID estável, turn e tentativas; pending/delivering/delivered |
| GraphCatalogChanges | GraphCatalogChange | outbox de rename/delete de autoria e reconciliação de seleção |

- `GraphWorkspaceDecision`: `Mode` original/new_worktree; `Attempt` fresh/reuse; `AuthorizationEventID`, `BaseRevision`, `SourceRunID`; `Reuse map[string]string` com chave `initial` ou `node:<id>:<occurrence>[:branch:<id>]` e valor workspace ID.
- v1 é lido/validado integralmente e escrito atomicamente como v2. Campos desconhecidos/estado inválido são recusados antes de qualquer overwrite. Projetos/sessões/grants/consultas/eventos/native IDs são preservados.
- Restart converte runs/ativações/operações incompletas em interrupted, cria notificação persistida e mantém atividades agendadas. `delivering` volta a `pending` com o mesmo ID. Nunca reexecuta script, Git, ativação ou sessão nativa.
- Shutdown aplica interrupção depois de aguardar `app.wg`; P2/P0 devem registrar todos os executores e garantir settlement real antes desse wait terminar.
- Invariantes verificam IDs/referências, snapshots, um slot/conversa, uma ativação ativa/sessão nativa, Choice válida antes de completar agente, entregas únicas, entradas Join separadas/rodadas não sobrepostas, provenance de workspace e outbox terminal.

## Arquivos do core

- Modificados: `engine/authoring.go`, `engine/store.go`, `engine/main.go`, `engine/api.go`.
- Criados: `engine/graph_contract.go`, `engine/graph_snapshot.go`, `engine/graph_records.go`, este relatório.
- Mudanças preexistentes preservadas; arquivos exclusivos do adapter/frontend e progresso principal não editados.

## Registro histórico das lacunas ao encerrar P1

- Ao encerrar P1 não existia execução simulada nem invocação stub. Dispatcher/runtime/scheduler/ferramentas/prompts/Terminal/workspaces/Fork/Join foram implementados na segunda rodada descrita acima; esta seção preserva o registro do ponto P1.
- Projeções/endpoints leem dados persistidos reais; permanecerão vazios enquanto runtime não produzir registros.
- Topologia rejeita destinos incompletos, caminhos paralelos terminais, Fork aninhado, entradas Choice exclusivas, colapso de chegadas e policies incompatíveis. Casos com reconvergência de origem ambígua são erro explícito; ciclos ordinários não são transformados em DAG. P5 ainda aplica readiness/deadlock/causalidade no despacho.
- Reinício preserva outbox, mas a entrega nativa ao orquestrador é trabalho P3. P1 não afirma garantia de término de processos externos/MCP.
- `gofmt` aplicado aos sete arquivos Go do core, sem erros de sintaxe reportados. Imports, assinaturas e pontos de integração conferidos por inspeção. Build Go ficará para a rodada coordenada após alterações do adapter; nenhuma compilação ou execução da migração foi alegada como validada.
- Validação humana após integração: roundtrip de output/isolation, seleção/None e rename/delete/reload; depois o runtime permitirá observar requests, LED e progresso real.
