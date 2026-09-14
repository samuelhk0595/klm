# Graph engine — FRONTEND P6

Atualizado: 2026-09-13. Escopo frontend P6 implementado nos contratos disponíveis
em `GRAPH_ENGINE_PROGRESS_CORE.md`. Validação de comportamento com runtime real
permanece para o usuário após a integração core/adapters.

## Implementado

### Catálogo e seleção

- Catálogo real por projeto compartilhado entre App/seletor e ProjectAuthoring,
  com `useSyncExternalStore`, cache por projeto e deduplicação de requests.
- CRUD invalida o catálogo após a mutação, inclusive se havia uma leitura anterior
  em andamento. Troca de projeto, retorno à área, abertura do seletor e mudança
  de seleção/run recarregam os arquivos reais.
- Falha HTTP mantém o último catálogo conhecido e apresenta erro/Retry. Erros
  parciais de arquivos são exibidos; não são tratados como catálogo vazio nem
  convertem a seleção persistida em None. Seleção indisponível mantém seu ID e
  uma mensagem recuperável. Nenhum grafo fictício é fallback produtivo.
- Seleção por conversa via GET/PATCH `/api/sessions/{id}/graph`, independente de
  Send. PATCH recebe exatamente `{selectedGraphId: string}`; None usa `""`.
- Alterações simultâneas de seleção são serializadas por conversa no cliente;
  a resposta autoritativa é aplicada pela revisão, sem seleção otimista que
  sobrescreva uma invocação mais recente.
- Rename/delete são reconciliados a partir do estado persistido do core por
  SSE/polling e GET após atualização do catálogo. Durante rename, o seletor usa
  o nome atual do catálogo e o canvas usa o snapshot original da run.
- None retorna a Chat e esconde Graph. Não cancela a run nem remove requests.
  Selecionar um grafo apenas disponibiliza Graph, sem abri-la ou executá-lo.

### Projeção e monitor

- DTOs `ConversationGraphState`, `GraphRunProjection` e
  `GraphRequestProjection` adicionados ao cliente, com `Session.graph`.
- Merge da projeção usa sua revisão monotônica, independentemente de
  `Session.updatedAt`. HTTP/SSE antigos são ignorados para o grafo; uma Session
  legada sem `graph` preserva a última projeção conhecida.
- Hidratação e polling GET pequeno da conversa selecionada; polling existente
  `/api/state` continua reconciliando todas as conversas. SSE inclui conversas
  com graph run ativa mesmo quando o chat está idle ou não selecionado.
- `run:null` autoritativo remove atividade/LED e devolve o canvas ao modo readonly
  ocioso, sem inventar conclusão de nós não executados ou histórico de execução.
- `GraphView` desenha `GraphRecord` real; durante run correspondente usa
  `run.snapshot`. Layout ausente, posição não finita ou viewport inválido têm
  fallback de apresentação determinístico, sem mudar a definição executável.
- Todos os IDs ativos/concluídos são considerados, suportando paralelismo e
  revisitas em ciclos. Ativo prevalece sobre conclusão anterior. Join coletando
  tem indicador Collecting separado do Running do agente.
- IDs crus de Choice recebem `choice:` somente no canvas. Agent, Choice, Fork,
  Join e Terminal mostram conclusão real; shimmer e LED reutilizam o estilo atual.
- LED de run somente no seletor do composer, condicionado ao grafo selecionado
  corresponder à run ativa, incluindo starting/ending e espera por usuário.
- Timer CSV, `getNodeActivity`, `onRunningChange` e painéis Run/Input/Output
  removidos do caminho produtivo. Os arquivos antigos de experimento permanecem
  sem importação pelo monitor produtivo.
- Canvas sem run preserva pan/zoom/reset e painéis de configuração readonly,
  incluindo output de Choice terminal, output de Terminal e isolamento por ramo.
- Resolver de apresentação usa harness/model/effort reais e overrides, sem
  defaults de fixtures nem decisão baseada no campo incidental `revision`.
- Não foram redesenhados header da aplicação, dimensões do canvas ou side chat.

### Solicitações humanas

- PermissionCard e QuestionCard existentes são reutilizados acima do composer,
  com origem legível (grafo/nó; Join identificado pelo agente referenciado).
- Respostas usam `GraphRequestProjection.sessionId` e `requestId` reais, pelos
  endpoints existentes. Nenhuma resposta é transformada em mensagem de chat.
- As respostas Session dos nós privados não são inseridas no navegador de sessões;
  o callback reidrata a projeção da conversa proprietária. Também há filtro por
  `role: graph_node` no merge/listagem/inscrições de sessões públicas.
- None mantém a fila de requests. Solicitações de nós não desabilitam o composer
  do orquestrador idle; grants e decisões continuam sendo os expostos pelo core.

### Autoria e templates

- Terminal possui editor de output plano e preview JSON, preservando Command
  literal. `output` ausente é `{}`, nunca encaminhamento implícito do input.
- Fork possui toggle Separate worktree por ramo; ramo novo e campo ausente usam
  true. False preserva o Git branch configurado, sem exigir ou validar esse nome
  como branch a criar; reativar isolamento restaura o campo e sua validação.
- `graphDrawing`/`canvasDefinition` preservam `output` de Terminal e
  `separate_worktree` em ambos os sentidos do contrato JSON/YAML do core.
- Editor reutilizado com sugestões dos outputs imediatamente conectados e
  inserção explícita de `payload.<field>`, sem lookup global de outros nós.
- Parser/validação/preview compartilham a gramática restrita do core:
  - Choice: `choice.<campo declarado>` e `run.input.task`.
  - Fork: `payload.<campo>` e `run.input.task`.
  - Terminal: os mesmos de Fork e `command.result`.
- Templates malformados e referências de outro contexto são erros explícitos.
  Preview de Choice rejeita campos extras/não-string e required ausente; optional
  declarado omitido resolve string vazia. Payload local ausente falha, sem `null`
  textual nem vazamento de template não resolvido. `$payload`/`$...` no script
  permanecem sintaxe PowerShell, sem interpolação KLM no Command.

## Contrato de integração

| Operação | Contrato consumido |
| --- | --- |
| Catálogo | `GET /api/projects/{projectId}/authoring` → `{agents, graphs, errors}` |
| Seleção/hidratação | `GET /api/sessions/{sessionId}/graph` → `ConversationGraphState` |
| Seleção | `PATCH /api/sessions/{sessionId}/graph`, `{selectedGraphId: "id"}` ou `""` → projeção completa |
| SSE | `/api/sessions/{sessionId}/events` → Session contendo projeção |
| Polling global | `/api/state` → sessões públicas, com `graph` quando disponível |
| Permissão | `POST /api/sessions/{request.sessionId}/permissions/{requestId}`, `{decision}` |
| Pergunta | `POST /api/sessions/{request.sessionId}/questions/{requestId}/reply`, `{answers}` ou `{cancelled:true}` |
| Autoria | Rotas existentes; Terminal `output?: Record<string,string>` e ramo `separate_worktree?: boolean` |

`run.graphId` é a identidade atual do catálogo; `run.snapshot.id` permanece a
identidade capturada. A run é independente da seleção. O cliente não adiciona
rota de execução, stop de graph run, histórico ou controles de retomada.

O `GraphRecord` da projeção contém definição/layout, **não** os TOMLs capturados
dos agentes. Por isso o monitor ativo identifica agentes pelas referências do
snapshot e não apresenta configurações do catálogo vivo como se fossem as da run.
No preview ocioso, as configurações vêm dos agentes reais do catálogo. Isso evita
atribuir modelo/effort incorretos após edição/rename de agente durante execução.

## Arquivos

Em `clients/desktop/src/`:

- `engine.ts`, `App.tsx`.
- `features/chat/GraphPicker.tsx`, `PermissionCard.tsx`, `QuestionCard.tsx`.
- `features/graphs/catalog.ts` (novo), `ProjectAuthoring.tsx`, `GraphsPage.tsx`.
- `features/graphs/files.ts`, `GraphView.tsx`, `GraphCanvas.tsx`, `graph-canvas.css`.
- `features/graphs/NodeConfigurationPanel.tsx`, `agent-node.ts`.
- `features/graphs/choice.ts`, `ChoiceNodePanel.tsx`, `ChoiceOutputEditor.tsx`,
  `ChoiceOutputPreview.tsx`.
- `features/graphs/fork.ts`, `ForkNodePanel.tsx`, `terminal.ts`,
  `TerminalNodePanel.tsx`.

Este relatório é o único arquivo alterado fora de `clients/desktop/**` pelo
responsável frontend. Alterações locais preexistentes de CRUD foram preservadas.
Nenhum Go ou progresso principal foi editado.

## Checagem e validação humana

- `npx tsc --noEmit` em `clients/desktop`: **passou, sem diagnósticos**.
  Executado ao estabilizar e novamente após o ajuste final do nome no seletor
  durante rename; a última execução cobre o código entregue.
- Não foram executados suites, testes de integração, reviews, modelos, comandos
  de grafos, commits, builds Go ou validação de runtime no navegador.
- Validar manualmente após integração: seleção/reload/None; rename/delete durante
  run; catálogo com erro; SSE/reconexão com chat idle; nós paralelos/ciclos/Join
  coletando; requests de nós respondidas com None; roundtrip YAML de Terminal e
  Fork misto; configuração readonly de modelos reais e Choice terminal.

P6 frontend está implementado; compilação TypeScript não é alegação de execução
real dos adapters nem de validação humana desses fluxos.
