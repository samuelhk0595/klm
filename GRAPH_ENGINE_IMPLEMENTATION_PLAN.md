# Graph engine — plano de implementação

Data: 2026-09-13.
Status: plano técnico produzido a pedido do usuário; implementação não iniciada.

Nota de precedência — 2026-09-13: o status acima registra a elaboração histórica
do plano; a entrega posterior está em `GRAPH_ENGINE_IMPLEMENTATION_PROGRESS.md`.
A decisão de invocação da seção 13 de `GRAPH_AUTHORING_REFINEMENT.md` prevalece:
`workspace` omitido usa `original`, sem pergunta nem evento obrigatório de escolha
do workspace. As instruções vigentes abaixo incorporam essa decisão; fresh/reuse
em tentativas com artefatos continua exigindo decisão do usuário.

Decisão posterior de lifecycle — 2026-09-13: a seção 14 do refinamento limita a
garantia ao nó e às chamadas de ferramentas. Permitir MCPs externos configurados
em OpenCode/Codex, sem allowlist por nome nem desabilitá-los; não exigir encerramento
de servidor compartilhado ou tarefa destacada após o retorno da chamada. Selo,
conclusão das chamadas iniciadas e registro de finality incerta permanecem obrigatórios.

## 1. Objetivo e precedência

Implementar a execução real dos grafos autorados no projeto, com Pi, OpenCode e
Codex, e o acompanhamento mínimo na aba Graph existente. Entregar em incrementos
funcionais, mantendo a engine Go independente do cliente.

Fonte dos contratos: `GRAPH_AUTHORING_REFINEMENT.md`, especialmente as seções
5–14. Ler também `AGENTS.md`, `product.md` e `CONTEXT.md`. Este plano escolhe
mecanismos técnicos para esses contratos; não substitui decisões de produto.

O refinamento da execução e do acompanhamento inicial está fechado. Questões
antigas de S1, R1/R2, C1, E1, W1/W2 e A1–A5 não devem ser reapresentadas como
pendências. Uma incompatibilidade técnica concreta deve ser apontada como tal.

### Dentro da entrega

- Invocação assíncrona pelo orquestrador, autorização vinculada à atividade,
  agendamento, tentativas corretivas e uma run ativa por conversa.
- Agent, Choice, Terminal, Fork e Join; sessões por nó/ativação, passagem explícita
  de dados, workspaces, resultados e encerramento de trabalho remanescente.
- Definições capturadas por run, registros persistentes, notificações e tratamento
  de reinício sem retomada automática.
- Catálogo real no seletor, seleção por conversa, LED no input e nós ativos/
  concluídos na Graph existente. Reutilização das perguntas/permissões atuais.
- Adequações de autoria necessárias: output de Terminal e isolamento por ramo.

### Fora da entrega

Forks aninhados; limites de ativações e timeouts de trabalho; limpeza de worktrees;
interrupção/retomada como controles do usuário; múltiplas runs simultâneas na mesma
conversa; entradas ricas/arquivos e `.klm/harness.toml`; validador visual completo;
abas novas de run, histórico navegável, logs/Input/Output, setas de ativações e
novo componente de solicitações na sidebar. Registros internos continuam necessários.

## 2. Base atual e pontos de integração

A análise foi estática, com três subagentes: núcleo/persistência, harnesses/bridge
e cliente/autoria. O CRUD existente foi validado pelo usuário e está não commitado;
preservar essas alterações e os artefatos visuais. Não reutilizar a engine antiga
de `klm-test` nem tratar as fixtures como contratos executáveis.

| Área | Código existente a aproveitar |
| --- | --- |
| Persistência | `engine/store.go`: `diskState`, `Session`, `Event`, `nativeSession`, `commitLocked`, `loadState`, `saveState` |
| Ciclo de vida | `engine/main.go`: `app`, `turn`, shutdown; `app.runs` representa turns de chat, não graph runs |
| Execução nativa | `engine/harness.go`: `app.execute`, `adapter`; `*_interactive.go`, `interactive.go`, `platform_*.go` |
| Ferramentas privadas | `engine/linked_bridge.go`: bridge autenticado e inventário; `linked.go`: reserva de turn e continuação |
| Interações | `engine/permissions.go`, `questions.go`; callbacks já roteados à sessão originadora |
| Arquivos de autoria | `engine/authoring.go`: `GraphDefinition`, `GraphNode`, `GraphBranch`, `validateGraph`, `validateOutput` |
| Catálogos de modelos | `engine/models.go`: catálogo real e seleção efetiva |
| Cliente e estado | `clients/desktop/src/engine.ts`, `App.tsx`: snapshots HTTP, SSE e polling |
| Grafos reais | `features/graphs/ProjectAuthoring.tsx`, `files.ts`: catálogo, conversão e desenho |
| Monitor atual | `GraphView.tsx`, `GraphCanvas.tsx`, `GraphPicker.tsx`: canvas, indicadores e seletor; atividade ainda simulada |
| Cartões reutilizáveis | `features/chat/PermissionCard.tsx`, `QuestionCard.tsx`: recebem `sessionId` e respondem à sessão correta |

Os caminhos de features acima são relativos a `clients/desktop/src/`.
`Session.Workspace` é agrupamento visual; não usá-lo como diretório de execução.
`ParentID` identifica side chat hoje; não usá-lo para simular sessões de nós.

## 3. Organização técnica e persistência

Manter o `package main`, armazenamento JSON transacional e mutex existentes.
Não introduzir banco, broker ou framework de workflows nesta entrega.

Arquivos novos sugeridos, criados conforme cada etapa precisar:

- `engine/graph_contract.go`: compilação, conexões, templates e Choices.
- `engine/graph_runtime.go`: registros, transições e despacho de ativações.
- `engine/graph_scheduler.go`: atividades autorizadas, dependências e notificações.
- `engine/graph_tools.go`: operações privadas de orquestrador/nó.
- `engine/graph_workspaces.go`: Git, diretórios, proveniência e reaproveitamento.
- `engine/graph_terminal.go`: PowerShell e coleta do resultado.
- `engine/graph_prompts.go` e `engine/prompts/`: carregamento/composição.

### Registros mínimos

| Registro | Conteúdo |
| --- | --- |
| `GraphActivity` | Conversa, objetivo autorizado, grafo, referências à autorização, ordem/prioridade, dependências, avaliação do resultado, correção e decisão de workspace |
| `GraphRun` | Activity/projeto/conversa, task, snapshot YAML/TOMLs, estado, workspace inicial, resultado/erro e timestamps |
| `GraphActivation` | Run/node/ocorrência, input, workspace, sessão nativa associada, estado, violações, Choice recebida/aceita e output |
| `GraphDelivery` | Conexão, ativação produtora, destinatário, payload e contexto causal; ID para deduplicação |
| `JoinRound`/`JoinArrival` | Rodada, entradas esperadas e recebidas, payloads separados, origem e integração |
| `WorkspaceRecord`/`WorkspaceUse` | Diretório/repositório/branch/base, criação e associações explícitas de usos posteriores |
| `RunNotification` | ID estável, resultado/run, conversa destinatária, estado de entrega e turn correlacionado |

Sessões dos nós reutilizam o transporte/armazenamento nativo, com propriedade
explícita de run/nó e cwd próprio. Ficam fora da lista de conversas e não ganham
ferramentas de leitura do chat. Distinguir seu papel das sessões main/side.

Migrar `state.json` de v1 para v2 com validação de referências e invariantes.
Preservar projetos, sessões, grants, consultas e IDs nativos; não sobrescrever
estado inválido. O executável antigo já recusa versões desconhecidas.

Transições passam por `app.mu` + `commitLocked`; I/O Git, processos e modelos
ocorrem fora do mutex. Persistir a reserva/intenção antes de iniciar o trabalho,
e correlacionar cada retorno com run/ativação/operação. Estender o cancelamento
por falha de armazenamento aos executores de grafo.

Manter contextos de graph run separados dos turns. O término do turn que invocou
um grafo não cancela a run. Reservar o slot da conversa em `starting`, `running`
e `ending`; somente liberar após o encerramento efetivo. Runs distintas continuam
permitidas em conversas diferentes do mesmo projeto.

## 4. Contrato executável, snapshot e dados

### Compilação separada do CRUD

Adicionar `compileGraph` antes do início de execução, preservando Save de autoria
e seus rascunhos. Compilar referências, tipos, conexões e expressões sem depender
do layout. Não exigir associação Fork→Join em campo de configuração.

Verificar destino inicial elegível, referências existentes, configurações efetivas
de modelos, destinos completos, saídas terminais e políticas de sessão elegíveis.
Rejeitar padrões já declarados inválidos: caminhos normais paralelos sem Join,
Choices mutuamente exclusivas como entradas obrigatórias distintas, Choice
compartilhada colapsando chegadas independentes e políticas de sessão conflitantes.
Não executar Fork aninhado como se estivesse suportado. Aceitar ciclos ordinários;
não substituir essa validação por exigir um DAG ou provar término de todo grafo.
Complementar verificações estáticas com invariantes durante o despacho.

Ao iniciar cada run, capturar a definição do grafo e todos os agentes utilizados,
incluindo overrides resolvidos e revisões de origem. Reutilizar `authoringMu` e
rechecar revisões de arquivos para detectar edição externa durante a leitura;
uma leitura instável falha de modo explícito, sem montar snapshot misturado.
Não prometer transação externa sobre vários arquivos editados por outro programa.

Capturar layout somente como apresentação opcional; ausência/invalidez de layout
não impede execução. O cliente usa posicionamento determinístico quando necessário.
Renomear/editar/desabilitar no catálogo não altera uma run capturada. Atividades
apenas agendadas e novas tentativas usam as definições vigentes ao efetivo início.

### Adequações do schema

- `GraphNode.output` para Terminal: objeto plano de strings/templates; ausente
  equivale a `{}`, nunca ao input implicitamente encaminhado.
- `GraphBranch.separate_worktree`: booleano opcional, default `true`, inclusive
  em arquivos existentes. Se false, compartilhar cwd/branch herdados e não exigir
  um nome para criar branch. Preservar nomes configurados sem usá-los para checkout.
- Choice autoral continua em YAML; o blocked interno é uma identidade runtime
  `{origin: "engine", id: "blocked"}` distinta de `{origin: "graph", id: slug}`.
  Ambos podem coexistir com o mesmo nome visível.

### Resolvedor único e limitado

Usar parser de referências, sem avaliação arbitrária de código:

| Contexto | Referências |
| --- | --- |
| Choice output | `choice.<field>`, `run.input.task` |
| Terminal output | `payload.<field>`, `command.result`, `run.input.task` |
| Fork branch output | `payload.<field>`, `run.input.task` |
| Script Terminal | Variável PowerShell `$payload`; sem substituição textual de templates no código |

O input inicial pode ser materializado como `{task: run.input.task}` junto das
instruções do nó. Payloads autorados permanecem planos; o contexto de Join pode
ter coleção estruturada própria sem ampliar a linguagem de output do autor.

Escolha técnica inicial: interpolar campo opcional declarado de Choice omitido
como string vazia; referência desconhecida ou campo local ausente é erro explícito
de resolução, nunca texto `null` ou vazamento de `{{...}}`. Alinhar preview e
validação de autoria. Falha de mapping é técnica, distinta de violação na submissão.
Não buscar resultados arbitrários de outros nós nem criar variáveis globais.

## 5. Harnesses, sessões e compromisso da Choice

### Primeiro ponto a comprovar durante a implementação

O código atual de `linked_answer` grava uma resposta e instrui o modelo a terminar;
isso não garante que nenhuma ação ocorra depois. A implementação precisa de uma
barreira real entre receber a Choice e aceitá-la definitivamente.

Sequência: validar → reservar submissão → selar a ativação contra novo trabalho →
aguardar chamadas efetivamente iniciadas e encerrar trabalho do nó → confirmar
conclusão dentro desse limite → persistir aceitação e
continuação. Nunca despachar o próximo nó diretamente do callback da ferramenta.
A reserva não é aceitação. Se o processo cair nessa janela, registrar interrupção;
não continuar a partir de uma aceitação que nunca foi confirmada.

O selo impede novas chamadas, inclusive ferramentas já permitidas. MCP configurado
não é evidência de chamada pendente: correlacionar inícios e conclusões reais.
Resposta “tarefa iniciada” conclui a responsabilidade de lifecycle da chamada,
sem garantir conclusão da tarefa remota destacada ou sucesso do objetivo.
Callback perdido, erro ou cancelamento sem comprovação de conclusão mantém
finality incerta; saída local/abort não comprova cancelamento remoto. Não exigir
encerramento de servidores MCP compartilhados para aceitar Choice.

Preservar coerência do histórico nativo e de sessões que serão continuadas:
interromper o harness com a ferramenta pendente não pode deixar uma sessão que
falha ao reutilizar. O protocolo de resposta/selamento depende de cada adapter.

| Harness | Aproveitar | Comprovação necessária |
| --- | --- | --- |
| Pi | `pi-permissions.ts`, tool gate, ferramenta sequencial, `agent_settled` | Selamento antes de bypasses/aprovações, ordem de ferramentas paralelas, abort/settlement e continuidade com cwd persistido |
| OpenCode | MCP owned, `prompt_async`, SSE/reconciliação, abort | Encerramento confirmado ou gate pré-execução que cubra ferramentas já permitidas; idle/step-finish isolado não prova ausência de trabalho |
| Codex | MCP owned, thread/turn, `turn/interrupt`, `turn/completed` | Selamento/encerramento de ferramentas paralelas e já autorizadas; preservar thread e grants nativos |

Essas verificações são o primeiro incremento técnico, não novas perguntas de
produto. Se um harness não oferecer mecanismo suficiente, registrar a limitação
concreta antes de anunciar suporte; não enfraquecer finality ou segurança para
simular conformidade. Nenhuma chamada a modelos foi feita para elaborar este plano.
O critério é por harness: um adapter comprovado pode servir ao primeiro incremento
sequencial enquanto os demais são adequados, sem declarar a entrega completa.

### Validação e correções

Centralizar a validação no engine: Choice disponível, identidade inequívoca,
strings, campos obrigatórios e ausência de extras. Campo opcional pode ser omitido;
campo obrigatório vazio ainda é string, sem requisito de conteúdo não aprovado.
Usar envelope de ferramenta que permita observar/validar payload no engine, em
vez de depender exclusivamente da rejeição de schema pelo harness.

Cada chamada inválida e cada término normal sem Choice contam uma violação.
Deduplicar por chamada/turn; uma rejeição e um término posterior sem Choice são
duas ocorrências reais, não a mesma notificação duplicada. Primeira e segunda
recebem correção; terceira falha a run. Erro técnico, cancelamento e espera por
usuário não são omissões. O contador não reinicia ao trocar de Choice e volta a
zero somente em nova ativação, inclusive com sessão continuada.

`continue_target` consulta a sessão mais recente daquele nó naquela run e compara
cwd canônico. Mesmo cwd permite reuso; diretório diferente ou sessão inexistente
cria nova. Nova run não importa histórico de uma run anterior. Reenvios corretivos
pertencem à ativação atual. `execute` deve distinguir conclusão normal, correção,
erro técnico e encerramento controlado por Choice, sem converter todos em Stop.

## 6. Orquestração, prompts e solicitações humanas

Estender o bridge privado com capacidades por finalidade, sem MCP público novo.
O token owned fixa conversa/run/ativação; não confiar em IDs arbitrários do modelo.
Manter linked tools para main/side; nós de grafo recebem apenas suas capacidades
de execução/Choice, sem acesso à conversa ou poder de invocar/parar outros grafos.

Operações técnicas sugeridas: catálogo resumido, invocar atividade autorizada,
consultar run, eventos recentes de nó, listar/agendar/atualizar atividade e submeter
Choice. Usar os mesmos handlers centrais para MCP de OpenCode/Codex e extensão Pi.
Invocação retorna ID/estado iniciado sem esperar pelo resultado do grafo.

Registrar referência à autorização expressa já existente na conversa. Isso dá
rastreabilidade, não comprovação semântica automática de consentimento nem grant
de permissão do harness. `authorization` mantém referências `{eventId,text}`;
validar evento existente do tipo `user` na conversa invocadora. Essa referência
fundamenta a atividade, não uma escolha obrigatória de workspace.

Publicar schemas completos dos objetos nested de `graph_invoke` nas ferramentas
MCP e Pi: tipos, campos obrigatórios/opcionais, enums, defaults, descrições e mapa
de reuse. A validação estrita não pode depender de o agente adivinhar `path`,
`scope` ou IDs de eventos. `workspace` é opcional, com default `original`, e
`workspace.authorizationEventId` não é obrigatório. Instruções do orquestrador
aplicam os defaults definidos e perguntam só por ambiguidade/informação bloqueante,
incluindo fresh/reuse ainda não decidido em retry com artefatos. Não adicionar
confirmação de workspace por omissão nem modal obrigatório a cada run.
Mutação de atividade usa versão/ID de operação para não duplicar despacho quando
uma notificação é entregue novamente. Registrar assessment por run de origem.

Atividade agendada não é run. O scheduler considera dependências e decisões de
workspace antes de reservar execução. Conclusão normal não é sucesso semântico:
o orquestrador avalia Choice/output antes de liberar dependentes. Correção concreta
pronta tem prioridade; independentes autorizadas seguem ordem dos pedidos enquanto
outra atividade aguarda usuário. Preservar a decisão expressa fresh/reuse e não
inventar contagem fixa de retries. Não dar ferramenta Stop ao orquestrador.

`RunNotification` funciona como caixa de saída persistida. Resultado terminal e
liberação do slot são persistidos juntos após encerrar todo trabalho. Chat idle
recebe turn de continuação; chat ocupado recebe depois. Unificar a arbitragem desse
turn com consultas linked para manter um único turn por sessão. Reentrega incerta
após restart usa o mesmo ID: garantia realista é pelo menos uma vez, com deduplicação
de efeitos, não uma transação exatamente uma vez entre arquivos locais e harness.

Criar os prompts aprovados: `orchestrator.md`, `agent-node.md`, `join-node.md`,
`invalid-choice.md`, `missing-choice.md`, `run-result.md` em `engine/prompts/`.
Instruções adicionais ficam em arquivos próprios; não espalhar prosa em handlers.
Reler a cada envio durante desenvolvimento. Separar dados variáveis da prosa;
snapshots de YAML/TOML não congelam esses prompts internos. Falha de carregamento
é explícita. Para pacote distribuído, prever os mesmos arquivos como recursos,
sem transformar empacotamento em projeto paralelo nesta entrega.

Perguntas/permissões dos nós usam os callbacks já existentes. Projetar solicitações
pendentes no estado da conversa proprietária, com node/run/session/request IDs.
Reutilizar `PermissionCard`/`QuestionCard` acima do composer, com origem legível e
resposta à sessão nativa correta, inclusive ao selecionar None. Não bloquear o
orquestrador enquanto outro nó aguarda. Grants de sessão pertencem ao nó, não ao
chat pai; grants de projeto/harness mantêm o escopo existente. Não criar sidebar
nova, converter resposta em mensagem comum ou enfraquecer sandbox/aprovações.

## 7. Terminal, workspaces e paralelismo

### Terminal

Executar o texto configurado como um único script PowerShell não interativo,
com cwd herdado e `$payload` carregado por canal de dados separado (por exemplo,
JSON UTF-8 em arquivo privado da ativação lido pelo wrapper). Não colar valores
no script nem dividir por `;`. `cd` afeta apenas esse script.

Capturar stdout e stderr em `command.result` na ordem observada, sem prometer
ordenação universal entre streams. Proposta inicial de coleta: até 64 KiB UTF-8
com truncamento explícito e metadado, continuando a drenar os pipes até terminar.
É limite técnico de coleta, não timeout ou limite de ativações. Evitar persistir
state.json a cada byte de progresso; logs nativos normalizados continuam limitados.

Exit não zero e stderr continuam pelo output mapeado. Falha ao iniciar processo
ou erro de infraestrutura da coleta falha a run. Não abrir terminal interativo,
responder automaticamente a prompts ou adicionar timeout de tarefa.

### Workspaces

Separar recurso físico de seus usos. Manter repo/cwd/branch/base e IDs de criação
de run, ativação e ramo. Criar worktrees em diretório gerenciado por projeto dentro
dos dados da engine, com nomes legíveis + run/activation/branch IDs, verificando
colisões. Executar Git com argumentos separados; nunca fazer checkout/reset no
diretório original para criar isolamento. Diretório compartilhado pode ser não-Git
se nenhuma operação Git de isolamento/integração exigir repositório.

Na invocação, respeitar modo explícito `original`/`new_worktree`; `workspace`
omitido resolve para `original`, a pasta registrada do projeto, sem path arbitrário.
Persistir o modo resolvido. A regra padrão do produto fundamenta essa resolução,
sem inventar consentimento expresso ou exigir referência à mensagem da escolha.
`workspace.authorizationEventId` não é obrigatório. Nova worktree usa commit
solicitado ou HEAD local atual da origem. Fork fixa uma única base por ativação antes das
criações isoladas; ramo sem isolamento mantém o cwd/branch herdado e seu dirty state.
Não fazer commit, push, pull, cópia de alterações ou cleanup implicitamente.

Reaproveitamento é mapa explícito fornecido pelo orquestrador. Chave técnica proposta:
`nodeId + occurrence + branchId?`, mais workspace inicial; valor é workspace ID
registrado. Isso permite identificar usos futuros antes de gerar activation IDs.
Registrar cada consumo no novo run; não selecionar por nome/recência. Validar
existência e compatibilidade com snapshot/repositório. Ambiguidade real volta ao
orquestrador; não descartar silenciosamente associações de grafo editado.

Tentativa fresh com worktree nova preserva a anterior e usa a base aprovada.
Exceção W2: no diretório original, iniciar sobre os arquivos existentes, sem
reset ou isolamento automático. Agentes podem ajustar/remover alterações conforme
a tarefa. A engine não promete um filesystem limpo nesse modo.

### Fork e Join

Criar entregas explícitas por conexão, com payload e proveniência causal. Identidade
de conexão vem do YAML (Choice→destino, Terminal→destino, Fork/ramo→destino), não
do layout. Não transformar uma Choice compartilhada em duas conexões artificiais.

Join mantém rodada e entradas separadas; `{}` conta como chegada. Correlacionar
entrega com a rodada/contexto causal, nunca apenas com a última rodada aberta.
Notificação duplicada é idempotente; segunda chegada independente indevida ou
rodadas sobrepostas são violações do grafo. Não exigir mesmo Fork de origem nem
adicionar agrupamento OR. Após rodada concluída, ciclos exigem entradas novas.

O agente só começa após todas as entradas. Fornecer os payloads identificados por
conexão e, separadamente, worktrees/branches/commits. Resolver política de sessão
por concordância das Choices recebidas; omissão é new, Terminal não vota e somente
Terminals resulta em new. A coleta de inputs não é uma execução antecipada do agente.

Join sem output branch integra na origem do trabalho paralelo; com output branch,
cria worktree própria a partir do HEAD atual dessa origem e a continuação fica ali.
Cada criação tem identidade única. O agente faz a integração e resolve conflitos;
se não conseguir, pode selecionar blocked. Não substituir isso por merge automático.

Quando só restarem dependências sem produtor executando ou executável, falhar com
diagnóstico das entradas faltantes (J8). Espera por usuário não é deadlock.
Se a compilação encontrar proveniência de origem ambígua em uma forma não aninhada,
isolar o caso concreto antes de inventar uma regra de escolha pela primeira chegada.

## 8. Encerramento e reinício

Em falha técnica ou blocked interno: fechar despacho, cancelar o trabalho do run,
confirmar término dos processos/ferramentas controlados e então publicar resultado
e liberar slot. Escolha terminal autoral só termina normalmente após reconvergência;
seu nome não decide falha/blocked. Preservar arquivos e worktrees.

Reforçar a propriedade dos processos: `killTree` atual e retornos por timeout de
espera dos adapters não comprovam término. No Windows, preferir contenção por Job
Object com encerramento dos filhos ao fechar; integrar à criação antes de trabalho
efetivo. Nas plataformas existentes, usar a primitiva equivalente já compatível
com o launcher, sem ampliar a linguagem PowerShell para outras shells.

Aplicar a seção 14: permitir MCPs externos configurados sem allowlist por nome ou
desativação. A engine responde pelo nó e pelas chamadas efetivamente iniciadas,
aguardando sua conclusão antes da aceitação. Não responde pelo encerramento do
servidor MCP compartilhado nem por tarefas destacadas além da resposta da chamada.
Isso substitui a regra preventiva que recusava configurações externas sem chamadas
pendentes demonstradas. Erro/cancelamento ou perda de callback sem comprovação de
conclusão registra incerteza, sem aceitação/resultado normal fictício. Matar o
harness não comprova cancelamento remoto. Grants/sandbox permanecem iguais.
Não matar um PID persistido sem comprovar sua propriedade/identidade.

Shutdown marca/encerra runs interrompidas. Restart não reexecuta ativações, scripts
ou criações Git incompletas. Preservar registros e artefatos, finalizar o diagnóstico
de interrupção, restaurar notificações e atividades agendadas. Solicitações humanas
cujo processo terminou ficam invalidadas como hoje; não reapresentar callbacks mortos.
Uma atividade dependente não ganha sucesso por causa da interrupção de sua origem.

## 9. Cliente mínimo e alinhamento de autoria

- Compartilhar o catálogo real de `ProjectAuthoring`/`files.ts` com o chat, removendo
  fallback produtivo para fixtures. Recarregar/invalidate após CRUD e ao trocar
  projeto; erro de carga não significa catálogo vazio ou seleção None.
- Persistir `selectedGraphId` na conversa por operação de seleção, fora do envio
  de mensagem. Início real atualiza seleção atomicamente. Reconciliar rename/delete
  com referências da sessão, preservando a identidade do snapshot ativo.
- Adicionar projeção pequena de run por conversa: ID, graph ID, active informado
  pela engine, revisão monotônica, snapshot/desenho e IDs ativos/concluídos. Dados
  volumosos de eventos/payloads não precisam ser transmitidos para esse monitor.
- O estado ativo da run não depende de haver uma ferramenta executando: esperas
  por usuário/inputs e encerramento ainda ocupam seu slot. A projeção distingue
  a coleta iniciada de Join da execução de seu agente, sem abrir painel de detalhes.
- Reutilizar SSE/polling e hidratação inicial. Inscrever conversas com graph run
  ativo mesmo quando o orquestrador estiver idle; ignorar snapshots antigos.
- `GraphView` recebe desenho real e progresso. Remover timer CSV, `getNodeActivity`
  e `onRunningChange` do caminho produtivo. Abrir a aba não controla lifecycle.
- Usar `GraphCanvas` e o shimmer/LED existente; adaptar Completed aos tipos que
  ainda não o mostram. Estado ativo prevalece sobre conclusão de ativação anterior;
  suportar paralelismo. Traduzir IDs de Choice para `choice:{id}` somente no canvas.
- Com run correspondente ativa, desenhar snapshot. Sem ela, manter configuração
  readonly atual. Ao terminar, limpar atividade/LED e retornar ao modo ocioso,
  sem histórico nem marcar todo o grafo completed. None esconde Graph, sem parar run.
- LED de execução fica no seletor do input. Não criar abas extras nem inspeção de
  execução Run/Input/Output. Preservar header, dimensões e side chat.
- Em autoria, adicionar output de Terminal e toggle por ramo de Fork, com roundtrip
  YAML. Reutilizar editores/tokens e alinhar parser/preview. Mostrar também output
  de Choice terminal no painel readonly; não inserir blocked interno editável.
- Tirar defaults fictícios de apresentação de snapshots reais: não depender do
  campo incidental `revision` para decidir se modelo/effort são reais.
- Conectar os cartões existentes às solicitações dos nós conforme seção 6,
  mantendo as sessões privadas fora do navegador de sessões.

## 10. Sequência de entrega e dependências

Cada incremento deixa arquivos compiláveis. O primeiro fluxo real usa um grafo
sequencial pequeno; ele não representa entrega completa antes das demais etapas.

| Etapa | Trabalho e critério de saída |
| --- | --- |
| P0 — Lifecycle dos adapters | Comprovar barreira Choice, término de ferramentas/processos e sessão continuável em cada harness. Implementar as menores extensões em `turn`/`adapter`/bridge. Registrar capacidades efetivamente verificadas. |
| P1 — Contratos e persistência | Schema/output/isolation, compilador, resolvedor, snapshots, v2 e registros. Carregar estado v1 sem perda; compilar definição sem layout e rejeitar contratos inválidos explicitamente. |
| P2 — Primeira execução sequencial | Agent→Choice→Agent ou Choice terminal, sessões privadas, três violações, prompts relidos, invocação assíncrona e resultado ao chat. Integrar solicitações aos cartões existentes antes da validação com agentes. |
| P3 — Ciclo de vida da atividade | Fila persistida, avaliação semântica pelo orquestrador, dependências, retries/fresh/reuse, notificações idle/busy e E1. Nenhuma execução concorrente indevida na mesma conversa. |
| P4 — Terminal e workspaces | PowerShell/payload/result, criação única, base correta, diretório compartilhado, mapa de reuso e exceção W2. Não alterar o repo original como efeito implícito. |
| P5 — Paralelismo e Join | Fork misto, entregas/rodadas, integração e continuação corretas; ciclos sequenciais, J7/J8 e bloqueios estruturais sem suporte aninhado. |
| P6 — Monitor mínimo e autoria | Catálogo/seleção real, snapshot/projeção SSE, LED, ativos/concluídos, reconexão e ajustes dos formulários; preservar o layout existente. |

P0 e P1 podem avançar em paralelo. Após definir DTOs, autoria/seleção de P6 pode
avançar com contratos reais enquanto P2–P5 são implementados. Core e adapter devem
ter dono coordenador: não editar `execute`, bridge e scheduler em paralelo sem
acordar interfaces. O caminho crítico é P0→P2 e P1→P2→P3/P4→P5→P6.

## 11. Validação proporcional

Na elaboração deste plano: somente leitura de código/documentos, análise estática
e conferência do texto. Nenhum modelo, grafo, build ou suíte foi executado.

Durante implementação, seguir `AGENTS.md`: menor checagem relevante, sem suítes
completas, testes de integração complexos ou revisão adversarial automática.
Para alterações Go, `go -C engine build`; para TypeScript, `npx tsc --noEmit`
em `clients/desktop`. Não executar ambos repetidamente quando só uma área mudou.
Testes pequenos do resolvedor/transições podem ser usados se houver necessidade
concreta; não criar um framework de testes como pré-requisito da entrega.

Roteiro de validação humana, com tarefas descartáveis e execução autorizada:

1. Selecionar/abrir/None não executa; pedido autorizado inicia e o chat continua livre.
   Após restart coordenado e reload da aba, retestar invocação sem `workspace`:
   usa a pasta original sem perguntar ou exigir evento da escolha; schemas nested
   descrevem os campos reais. Autorização da atividade continua vinculada a evento
   `user` da conversa; retry com artefatos mantém a decisão fresh/reuse.
2. Agent escolhe saída, destino recebe mapping e nenhuma ação ocorre após aceitação;
   verificar isso nos três harnesses, inclusive sessão continuada e ferramentas.
3. Erros de payload/omissões misturados chegam a três por ativação; built-in e
   Choice autoral blocked coexistem e produzem resultados distintos.
4. Pergunta/permissão de nó é respondível com origem correta enquanto chat continua;
   grants e side chat mantêm seu comportamento.
5. Terminal trata quotes/semicolons no payload como dados, mantém `cd` local e
   segue após exit não zero. Falha ao iniciar é técnica.
6. Fork misto/Join entregam inputs separados, base e destino corretos; ciclo usa
   novas chegadas e sessão nova quando cwd muda. Blocked/falha encerra os irmãos.
7. Run ativa usa snapshot apesar de edições; tentativa nova usa definições atuais.
   Reuse é explícito; fresh no original aceita o dirty state sem reset.
8. Resultado acorda chat idle ou aguarda ocupado. Dependência não confunde conclusão
   com sucesso. Restart preserva notificações/fila sem repetir scripts ou retomar run.
9. Graph mostra todos os ativos e os concluídos reais; LED/reload/reconexão/None
   não inventam execução. CRUD, layout e configuração readonly permanecem utilizáveis.

O grafo ilustrativo `Graph execution reference` contém incompatibilidades antigas;
alinhar uma cópia ou criar exemplo descartável válido antes de usá-lo para executar.
Não executar seus comandos atuais como validação automática deste plano.

## 12. Riscos técnicos e ponto de retomada

1. **Finality e término real:** principal dependência P0; transporte MCP e pedido de
   abort existentes não bastam como evidência. Inclui histórico continuável e conclusão
   das chamadas iniciadas; servidores compartilhados e tarefas destacadas além
   da resposta estão fora da garantia, conforme seção 14.
2. **Estado local versus harness/Git:** não há transação distribuída. Usar intenções,
   IDs de operação, reconciliação e interrupção explícita; nunca replay automático
   de efeito incerto ou promessa de entrega exatamente uma vez.
3. **Proveniência/rodadas/reuse:** manter IDs explícitos. Se surgir caso concreto
   não coberto pelas regras, descrever o grafo/caso antes de inventar semântica.
4. **state.json monolítico:** suficiente para a primeira entrega, mas evitar gravação
   de cada byte/delta. Otimização de armazenamento fica para necessidade observada.

Próxima sessão de implementação: ler este plano e as regras canônicas, verificar
alterações locais e começar P0/P1. Manter atualizados os marcos implementados versus
pendentes. A aprovação para produzir este plano e usar subagentes não é autorização
para commit/push nem para executar tarefas reais dos grafos.

Referência durável no AI Memory: `plans/graph-engine-implementation.md`.
Este arquivo contém o plano completo; a nota de status aponta para ele e para sua
cópia na memória, substituindo o antigo ponto de retomada após Join.
