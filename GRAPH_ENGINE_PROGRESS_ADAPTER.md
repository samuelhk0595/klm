# Graph engine — relatório do implementador de adapters (P0)

## Precedência documental — MCP externo, 2026-09-13

A seção 14 de `GRAPH_AUTHORING_REFINEMENT.md` substitui as recusas preventivas de
MCP externo e exigências de cancelamento de servidor descritas nas rodadas abaixo.
Permitir MCPs configurados em OpenCode/Codex sem allowlist por nome/desativação.
A garantia cobre o nó e chamadas iniciadas: impedir novas chamadas após selo e
aguardar conclusão antes de aceitar Choice. Servidor compartilhado e tarefa
destacada além da resposta ficam fora do escopo; “tarefa iniciada” conclui a chamada,
não a tarefa. Erro/cancelamento ou callback perdido sem evidência de conclusão
mantém finality incerta, sem resultado normal ou garantia de cancelamento remoto.
Grants/sandbox permanecem iguais. Esta nota não comprova implementação, build,
teste ou restart da alteração; preserva as evidências históricas abaixo.
Registro documental: `GRAPH_ENGINE_PROGRESS_EXTERNAL_MCP_DOCS.md`. O principal
coordena o restart solicitado após o build; usuário recarrega a aba e retesta.

Atualizado: 2026-09-13. Segunda rodada implementada; build integrado pertence ao core.

## Segunda rodada — estado vigente (substitui o status histórico abaixo)

### Mecanismos implementados

| Harness / plataforma | Mecanismo de finality | Condições e limites |
| --- | --- | --- |
| Pi / Windows | Gate sequencial da extensão → abort após tool results → agent_settled → Job vazio → aceitação | Mesmo mecanismo da primeira rodada; sessões/cwd conferidos. |
| OpenCode / Windows | Plugin owned `tool.execute.before` e `shell.env` → recibo nativo → abort/reconciliação → Job vazio → aceitação | Exige versão inspecionada 1.18.30, handshake e inventário nativo. MCP externo habilitado impede iniciar o prompt. |
| Codex / Windows | Reserva → recibo MCP → interrupt/completed → shutdown nativo por EOF → Job vazio → leitura fria do histórico persistido → selo físico/aceitação | Não depende de PreToolUse. MCP externo impede iniciar o prompt. O resultado permanece **reservado**, sem aceitação, durante o encerramento. |
| Unix | Process group existente | Continua recusado para nós: ausência de contenção equivalente comprovável para descendentes que saem do grupo. |

Os três caminhos Windows têm mecanismos implementados e são admitidos pelo preflight
de capacidade. Isso **não significa validação de uma tarefa real ou de continuação
com modelo**: `RuntimeValidated` continua false. OpenCode/Codex recusam configurações
externas incompatíveis em vez de esconder/desabilitar silenciosamente essas ferramentas.
Não alterei sandbox/read-only, grants, políticas de aprovação ou ferramentas essenciais
de arquivos/comandos para habilitar um adapter.

### Factory acordada com core

```go
var graphConversationHooks func(*app, *turn, Session) (*GraphAdapterHooks, error)
```

`execute` inicializa `t.prompt` com o texto recebido se vazio, chama a factory para
main/side sem binding, usa o `t.prompt` enriquecido e instala os hooks retornados.
Tudo ocorre antes de catálogo, bridge ou requests nativos e sem `a.mu`.
Erro segue o cleanup normal: estado error, solicitações limpas, `a.runs` liberado,
`t.done` fechado, cancelamento e `a.wg.Done`. Um nó privado sem binding é recusado;
não recebe acidentalmente ferramentas de conversa.

A factory pode retornar **hooks e erro juntos**: `Finished` continua instalado para
registrar uma falha de notificação, embora nenhum harness seja iniciado. Se ela
falhar antes de construir os hooks, `execute` chama o callback existente do core
`a.graphNotificationFinished(t.graphNotificationID, result)` como fallback, depois
de liberar o turn/mutex. Assim, a notificação reservada não fica `delivering` para
sempre devido a um erro de leitura de prompt. A integração usa os símbolos P3 já
presentes; não alterei `graph_tools.go`, `graph_scheduler.go` nem `linked.go`.

API existente preservada. Adicionado `GraphAdapterCapability.ProcessSeal` (JSON
`processSeal`) para distinguir selamento físico de gate pré-tool. `CheckGraphAdapter`
aceita mecanismo de gate **ou** de selo físico, sempre com contenção owned.

### OpenCode: plugin real e escopo

- Novo `engine/opencode-graph-plugin.mjs`, embutido via go:embed em
  `opencode_interactive.go`; cada turn de nó recebe arquivo privado temporário.
- Configuração injetada só no processo owned por `OPENCODE_CONFIG_CONTENT`;
  config/plugin existentes são preservados. JSON que não pode ser estendido com
  segurança gera erro explícito.
- Handshake autenticado por `klm/graph/opencode`, fora do inventário do modelo.
  A admissão e a reserva compartilham o mutex; falha de transporte ou selo faz o
  plugin lançar erro **antes** da execução nativa. Só a sessão raiz submete Choice;
  sessões nativas de subtask compartilham o gate da ativação.
- `/global/health` deve confirmar 1.18.30; `/experimental/tool/ids` fixa o inventário
  nativo admitido. Ferramenta nova/desconhecida posterior não ganha admissão implícita.
- `/mcp` é inspecionado antes do prompt: servidor não-owned e não-disabled é erro,
  pois encerrar um cliente MCP não comprova que seu serviço terminou trabalho.
- Leitura da tag v1.18.30 confirmou que `Plugin.trigger` propaga a falha do before
  hook e que `session/tools.ts` o executa antes de `item.execute` e da chamada MCP.
  O plugin também cobre o hook `shell.env` da execução shell nativa.

### Codex: alternativa física e histórico continuável

- CLI instalada: **0.153.4**. `features list` confirma `hooks` stable/enabled.
- A tag rust-v0.153.4 tem `PreToolUse`, inclusive hooks de comando e MCP. Contudo,
  `events/pre_tool_use.rs` deixa `should_block=false` quando o handler falha, retorna
  JSON inválido, não inicia ou ocorre falha de serialização. Hooks managed/required
  verificam carregamento; não mudam essa política de execução. Hooks executor-scoped
  são assíncronos e não aplicam controle. Portanto esse hook **não** virou nosso selo.
- Alternativa implementada: nenhuma aceitação enquanto existir o processo executor.
  Após o tool result correlacionado, solicitar interrupt e esperar o turno terminal;
  fechar stdin para shutdown nativo, drenar frames e aguardar saída, depois terminar
  e conferir o Job. O selo físico cobre ferramentas paralelas/já autorizadas dentro
  do Job; não reimplementa ferramentas do harness.
- A leitura de `core/src/tasks/mod.rs` mostrou um detalhe importante: a emissão de
  `TurnComplete`/`TurnAborted` precede o último `flush_rollout`. Por isso acrescentei
  **verificação fria**, em outro app-server owned, usando somente initialize e
  `thread/read {includeTurns:true}`. Não envia thread/start, thread/resume ou prompt.
  O thread deve existir no storage, conter a ativação terminal e o recibo de Choice
  correspondente. Só depois de fechar/drenar também esse processo é possível aceitar.
- Erro/timeout/ausência do recibo impede aceitação/continuação. As esperas são de
  infraestrutura (shutdown/consulta), não timeout da tarefa ou limite de ativações.
- `mcpServerStatus/list` percorre **todas** as páginas no modo nó; qualquer servidor
  além do owned recusa execução antes de turn/start. Chat/orquestrador conserva o
  comportamento MCP existente.

### Evidência nova — sem modelos nem turns de tarefa

1. Comandos `--version`: OpenCode **1.18.30**, Codex **0.153.4**; `codex features list`.
2. Typecheck específico do novo plugin JS por TypeScript (`--allowJs --checkJs
   --strict --noEmit --skipLibCheck`): sem erros.
3. Probe isolado do plugin: handshake/admissão antes do selo, recusa de tool e shell
   depois do selo, e falha de transporte fail-closed. Retornou todos os cinco flags
   true. Isso verifica o módulo; não é uma execução de tool com modelo.
4. Probe **nativo** de startup OpenCode em diretório/config XDG temporários: chamou
   somente health e inventário, confirmou carregamento real do plugin e handshake.
   Saída: `version=1.18.30, pluginLoaded=true, ownedHandshake=true,
   nativeToolInventory=true, modelCalls=0, taskTurns=0`. Processo foi encerrado e
   diretório temporário removido.
5. Probe Windows inerte por P/Invoke das APIs utilizadas: Job configurado com limite
   de 144 bytes, 1 processo suspenso, **3 processos** depois de retomar/gerar filhos,
   **0 processos** após TerminateJobObject; parent exit confirmado. Cwd somente temp.
   Isso verifica as primitivas/estrutura no Windows atual; **não** compila nem
   exercita o launcher Go integrado. Não foi usado como evidência de MCP externo.
6. `gofmt` nos arquivos deste escopo. **Nenhum build Go**, suíte, review, commit,
   modelo ou tarefa de grafo real executado.

Scripts temporários dos probes ficam no diretório preaprovado
`C:/Users/Samuel/AppData/Local/Temp/opencode/`:
`klm-job-smoke.ps1`, `klm-opencode-plugin-smoke.mjs`,
`klm-opencode-startup-smoke.mjs`. Nenhuma fixture foi inserida no projeto do usuário.

### Limites que permanecem

- Job Object controla processos owned; não cancela daemon preexistente, serviço
  remoto ou trabalho disparado por um comando fora desse domínio. O flag
  `ExternalWorkConfirmed` não é promovido a true por abort, EOF, MCP response ou PID.
- A recusa explícita de servidores MCP externos é necessária enquanto não houver
  contrato de settlement/cancelamento desses serviços. A segurança nativa não é
  reduzida para contornar isso. Código arbitrário de plugins e comandos pode ter
  efeitos externos; nenhum adapter promete transação/rollback ou contenção de rede.
- Falta executar tarefas descartáveis autorizadas para validar Choice nativa,
  continuação de sessão e o fluxo integrado dos três harnesses. O caminho Codex de
  leitura fria/persistência foi implementado a partir da API/código inspecionados,
  mas ainda não foi exercitado com um histórico contendo uma Choice real.
- Core é responsável pelo build integrado. O fallback de outbox usa seu método P3
  existente; todos os arquivos fora do ownership continuam preservados.

---

## Registro da primeira rodada — histórico, não status atual

## Estado real

Há implementação de lifecycle, bridge por finalidade, hooks de integração e barreira
Choice. **P0 não está comprovado integralmente nos três harnesses.** Pi/Windows tem
o caminho de barreira implementado; OpenCode/Codex permanecem recusados para executar
nós de grafo pelo preflight, pois os protocolos atualmente integrados não têm gate
pré-execução cobrindo ferramentas já autorizadas. Chat, consultas linked e ferramentas
de orquestrador continuam disponíveis nos três harnesses.

Não executei modelos, tarefas reais, builds Go, suítes, reviews, commits ou pushes.
Nenhum campo foi exigido em `app`, `turn`, `Session` ou `nativeSession` nesta rodada.

## Arquivos de responsabilidade alterados

- Novo `engine/graph_adapter.go`.
- `engine/harness.go`.
- `engine/linked_bridge.go`.
- `engine/interactive.go`.
- `engine/pi_interactive.go` e `engine/pi-permissions.ts`.
- `engine/opencode_interactive.go`.
- `engine/codex_interactive.go`.
- `engine/platform_windows.go` e `engine/platform_unix.go`.
- Este relatório.

Não alterei `main.go`, `store.go`, `api.go`, `authoring.go`, demais `graph_*.go`,
frontend ou documento principal de progresso.

## API entregue ao core (package main)

```go
type GraphAdapterHooks struct {
    Node          bool
    Tools         []map[string]any
    Call          func(context.Context, string, json.RawMessage) (any, error)
    ReserveChoice func(context.Context, string, json.RawMessage) (any, error)
    Transition    func(context.Context, string) error
    Finished      func(GraphAdapterResult)
}

func BindGraphAdapter(t *turn, hooks GraphAdapterHooks) error
func UnbindGraphAdapter(t *turn)
func CheckGraphAdapter(harness string) error
func GraphAdapterCapabilities() []GraphAdapterCapability
func CanonicalGraphDirectory(path string) (string, error)

type GraphOwnedProcess interface {
    Attach() error
    Stop() error
    Close() error
    Drained() bool
}
func PrepareGraphProcess(cmd *exec.Cmd) (GraphOwnedProcess, error)
```

### Integração de uma ativação

1. Core cria/reserva a sessão privada e seu `turn`, guarda em `a.runs[sessionID]`,
   persiste intenção e chama `BindGraphAdapter` **antes** de `go a.execute(...)`.
   `Node: true`; ownership run/node/ativação fica capturado nos callbacks, não no
   JSON fornecido pelo modelo. Reservar `a.wg` como nas chamadas atuais de execute.
2. `ReserveChoice(ctx, callID, raw)` valida e persiste a reserva. O envelope anunciado
   é `{choice: {origin, id}, payload: {...}}`; o schema de transporte é permissivo
   para deixar identidade, obrigatórios, tipos, extras e correção sob responsabilidade
   do core. O callback pode retornar dados de diagnóstico/recibo. Erro devolve feedback
   à ferramenta e deixa a barreira aberta. Core controla contagem/deduplicação e a
   terceira violação; não há outro contador nos adapters.
3. `Transition(ctx, "sealed")`, depois `"drained"` e `"accepted"`, permite persistir
   as fases. Cada callback deve confirmar sua transação antes de retornar nil.
   Nenhum callback deve despachar o destino. Não fazer callback reentrante no binding.
4. `Finished(result)` roda **depois** da finalização da sessão, remoção do turn de
   `a.runs`, fechamento de `t.done` e liberação de `a.mu`. Só aqui agendar continuação
   ou correção. Todos os callbacks são invocados sem `a.mu`.
5. `execute` remove automaticamente o binding; `UnbindGraphAdapter` só é necessário
   quando uma reserva é abandonada antes de iniciar `execute`.

`GraphAdapterResult` contém `Outcome`, `ChoiceState`, `NativeSettled`,
`ChoiceToolSettled`, `ProcessesDrained`, `Continuable`, `ExternalWorkConfirmed` e
`Error`. Outcomes: `completed`, `missing_choice`, `choice`, `failed`, `interrupted`,
`finality_unconfirmed`. Só `choice` com `ChoiceState == "accepted"` permite seguir
uma Choice. `missing_choice` só aparece após término nativo normal; core contabiliza
a omissão, acorda a mesma ativação até a segunda violação e falha na terceira.

`Continuable` informa settlement/encerramento necessário para reutilização; **não é
evidência de uma continuação realmente executada nesta sessão de desenvolvimento**.
Core ainda deve selecionar a sessão mais recente do mesmo nó/run/harness e comparar
cwd canônico. Outra run, outro cwd ou sessão inexistente exigem sessão nova.

### Ferramentas do orquestrador

Usar `Node: false`, `Tools` com inventário MCP e `Call` para handlers centrais.
Nomes adicionais precisam ser únicos e começar com `graph_`; não podem sobrescrever
`graph_submit_choice`. Linked tools permanecem para main/side. O prefixo linked
de instruções não é inserido em nós privados. A extensão Pi recebe o mesmo inventário
do bridge; OpenCode/Codex verificam as ferramentas do inventário efetivo.

Bindings não são persistidos nem derivados de IDs enviados pelo modelo. Um token
de nó não pode descobrir/ler/consultar linked chats adivinhando nomes de ferramentas.
O gate Pi é método autenticado `klm/graph/gate`, não ferramenta do modelo.

### Terminal/processos

`PrepareGraphProcess(cmd)` ocorre antes de `cmd.Start()`. Para `CommandContext`,
atribuir `cmd.Cancel = owner.Stop`; após Start, chamar `owner.Attach()` imediatamente.
Se Attach falhar: Stop, Wait e Close. No término normal: Wait, Close e conferir
Drained. A espera de confirmação de infraestrutura é limitada; expiração não
significa término comprovado e não deve liberar o slot da run como sucesso.

## O que a barreira implementa

- Reserva serializada, independente de aceitação, com recibo aleatório correlacionado.
- Repetição do mesmo request não aceita/submete novamente; outra Choice após reserva
  é rejeitada. Identificação do resultado nativo exige o recibo da reserva, evitando
  confundir conclusão de uma chamada anteriormente inválida com a Choice reservada.
- Pi consulta o gate antes de bypass interno/permissões e novamente depois da espera
  por aprovação; selamento é checado também localmente. Choice é ferramenta sequencial.
- Depois da reserva, ferramentas posteriores do lote são bloqueadas. A extensão
  pede abort em `turn_end`, quando os tool results do lote já foram acrescentados ao
  histórico. Não interrompe dentro do callback MCP com a ferramenta pendente.
- Aceitação só após `agent_settled`, recibo de ferramenta correspondente e drenagem
  dos processos owned, seguida dos commits `drained` e `accepted`.
- Falha/stop/interrupção não viram Choice ou omissão. Reserva sem confirmação não
  vira aceitação. Não há dependência da disponibilidade do orquestrador para finalizar.
- Pi confere cwd canônico no header JSONL e recusa substituir silenciosamente um
  arquivo nativo perdido ou um ID nativo diferente ao continuar a sessão.
- OpenCode/Codex têm reconhecimento do recibo, interrupção nativa posterior ao tool
  result, confirmação nativa separada, gate nas aprovações e drenagem owned. Isso
  **não** cobre execução nativa já permitida; por isso não foram habilitados para nós.

## Contenção e limites concretos

### Windows

Launch suspenso (`CREATE_SUSPENDED`), Job Object com `KILL_ON_JOB_CLOSE`, atribuição
antes de retomar o processo (`NtResumeProcess`), sem flags de breakaway. Falha de
atribuição impede executar o harness. Cancelamento termina o Job e Close consulta
`ActiveProcesses == 0`; o retorno de taskkill/timeout não é usado como comprovação.
Os handles não são herdados/persistidos. Não há kill de PID recuperado de state.json.

Isso cobre os processos pertencentes ao Job, inclusive descendentes. Não é sandbox
de rede e não comprova término de serviço externo acionado por shell, MCP remoto,
daemon preexistente ou outro trabalho delegado fora do Job. O campo
`ExternalWorkConfirmed` permanece false. A engine não tem transação de cancelamento
com esses serviços e este relatório não anuncia essa garantia.

O uso das APIs de Job/suspensão está implementado, mas ainda requer exercício nativo
local autorizado; não foi validado rodando processos nesta rodada.

### Unix

Mantida finalização por process group. Um descendente pode sair do grupo com setsid;
portanto `Drained()` não afirma contenção e o preflight de nós de grafo recusa a
plataforma por enquanto. Não introduzi suporte fictício equivalente a Job Object.

### OpenCode e Codex

O bridge MCP próprio observa apenas suas chamadas. Callbacks de aprovação não são
um hook universal anterior a ferramentas já permitidas. `session.idle`, `/abort`,
`turn/interrupt` e `turn/completed` isolados não demonstram esse gate nem término de
serviços externos. `CheckGraphAdapter` recusa nós desses harnesses antes do prompt
da tarefa; não executa trabalho para só depois descobrir que não pode aceitar o
resultado. A integração de orquestrador continua utilizável neles.

## Evidência e checagens realizadas

- Documentos obrigatórios lidos, incluindo `GRAPH_AUTHORING_REFINEMENT.md` inteiro.
- Leitura estática do Pi **0.85.1** instalado: `agent-loop.js` demonstra seleção do
  caminho sequencial quando qualquer ferramenta do lote é sequencial, gates antes de
  executar e publicação dos resultados; `agent-session.js` documenta/persiste tool
  results antes de `turn_end`, além de emitir `agent_settled` após o run.
- `gofmt -w` aplicado somente aos arquivos Go deste escopo: parsing/formatação sem
  erro. Isso não é checagem de tipos/imports de package nem build aprovado.
- Typecheck isolado de `engine/pi-permissions.ts` com `tsc --noEmit`, strict,
  skipLibCheck e paths apontando para o SDK Pi instalado. A primeira execução detectou
  inferência demasiado estreita de `crypto.randomUUID()`; acrescentei `callId: string`.
  A segunda execução terminou sem erros. Config temporário fora do repo:
  `C:/Users/Samuel/AppData/Local/Temp/opencode/klm-adapter-pi-tsconfig.json`.
- **Build Go não executado**, respeitando edição simultânea do core. Nenhum novo
  símbolo depende de campos ainda ausentes do core; integração/build global pendentes.
- Nenhuma chamada a modelos, teste de continuação nativa, suíte ou review realizado.

## Segunda rodada / pendências para o orquestrador

1. Core pluga os hooks e seu ownership/snapshot/counter persistente; pode usar Pi no
   Windows como primeiro caminho implementado, condicionado à validação humana.
2. Coordenar build Go único quando o core estabilizar. Não declarar P0 integralmente
   comprovado a partir de gofmt/typecheck.
3. Integrar inventários/handlers de orquestrador em turns main/side e notificações;
   API acima evita editar `execute` em paralelo.
4. Cobrir gate pré-execução real de OpenCode (possível plugin owned ainda não
   implementado) e verificar alternativa concreta no Codex antes de habilitar nós.
5. Validar Pi Choice + tentativa de ferramenta posterior + continuação no mesmo cwd,
   e Job Object com descendentes, somente quando a execução nativa for autorizada.
6. Serviços externos exigem contrato próprio de cancelamento/settlement; sem isso,
   não liberar/afirmar encerramento confirmado de trabalho delegado a esses serviços.

Não considero a entrega completa dos três adapters alcançada nesta rodada.
