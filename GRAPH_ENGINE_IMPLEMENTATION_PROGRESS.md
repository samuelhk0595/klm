# Graph engine — progresso de implementação

Atualizado: 2026-09-13. Implementação P0–P6 entregue por subagentes para validação humana; agente principal atuou como orquestrador. Build/typecheck aprovados, engine reiniciada e cliente aberto no Playwright. P0 tem mecanismos implementados, mas comprovação integrada de Choice/continuação com modelos permanece pendente.

## Correção de MCP externo — estado vigente

Em 2026-09-13, usuário reportou run `b9307f7f23e21b42cc3a5c75cb9eab92` falhando antes do prompt por `Graph execution cannot confirm lifecycle of external OpenCode MCP server: agentdeck`. A causa era ban genérico de configuração externa, não prova de chamada pendente.

Usuário aprovou limitar a garantia ao nó e suas chamadas: permitir MCPs configurados, selar novas chamadas e aguardar as iniciadas antes de aceitar Choice; servidor compartilhado e tarefas destacadas após resposta ficam fora da garantia. Cancelamento/erro sem comprovação da resposta registra incerteza. Regra canônica: seção 14 de `GRAPH_AUTHORING_REFINEMENT.md`; substitui ban descrito em relatórios históricos.

- Implementador `ses_f666b6c91ffep5TvXEp61Jfy0I`: removidos bans OpenCode/Codex, rastreador por ativação/sessão/turn/call, plugin after/gate, eventos/protocolo e propagação de incerteza em ending/FinalityError. Choice não aguarda sua própria resposta em deadlock. Nenhuma allowlist por nome ou alteração de configs/grants/sandbox.
- Documentação `ses_f64129a17ffe2smXGWD4Dt8yuU`: contrato/plano/produto/contexto/READMEs alinhados.
- Relatórios: `GRAPH_ENGINE_PROGRESS_EXTERNAL_MCP_FIX.md` e `GRAPH_ENGINE_PROGRESS_EXTERNAL_MCP_DOCS.md`.
- Checagens: cinco testes focados `go -C engine test -run '^TestGraphMCPBoundary' -count=1 .` aprovados; typecheck específico do plugin aprovado; build `go -C engine build -o engine-mcp.exe .` aprovado.
- Limite conhecido OpenCode 1.18.30: exceção de ferramenta não garante after; sem evidência de resposta real, caso permanece incerto. Tarefa remota destacada não passa a ser monitorada após resposta.
- Restart concluído por `ses_f648c5e94ffe2CrIX8CXYM4SU4`: binário ativo **engine/engine-mcp.exe**, PID observado 12204, http://127.0.0.1:7331, health ok v2, zero turns/runs ativos. Engine anterior saiu graciosamente (exit 0). Ambiente/cwd/args/data-dir mantidos; engine.exe/engine-next.exe são versões anteriores.
- Relatório operacional: `GRAPH_ENGINE_PROGRESS_EXTERNAL_MCP_OPERATIONS.md`. Contagens públicas mantidas (3 projetos, 29 sessões, 542 eventos); projetos iguais. Comparação completa de sessões diferiu; incremento de GraphRevision no shutdown é compatível com isso, mas snapshot anterior não foi retido, portanto não afirmar igualdade integral nem causa única confirmada. Startup/health sem erro.
- Próximo: usuário recarrega http://127.0.0.1:5173/ e retesta com MCPs externos. Não houve navegador/chat/grafo/modelo acionado pelo agente operacional; Choice/continuação nativas permanecem para validação humana.

## Correção após primeiro teste humano — histórico concluído

Em 2026-09-13, o usuário relatou falhas de `graph_invoke` na conversa Testes. A investigação identificou schemas nested incompletos em authorization/workspace combinados com decoder estrito; seis chamadas foram rejeitadas antes de run. A mensagem usada como referência não continha decisão de workspace.

Decisão posterior aprovada: omissão de workspace/modo significa `original`, sem perguntar nem exigir evento que fundamente a escolha. A autorização da atividade continua como `{eventId,text}`; default vem da regra do produto, não de consentimento inferido da omissão. Fresh/reuse em retries com artefatos permanece explícito. Esta decisão substitui exigências de workspace dos relatórios históricos; contrato canônico atualizado na seção 13 do refinamento.

- Implementador: `ses_f64129a2fffeehCxJAwotF89Ar`. Schemas completos de invoke/update, default normalizado, referência de workspace opcional, validação antecipada da autorização e erros com caminho completo. Prompt com exemplo mínimo e aplicação de defaults.
- Documentação: `ses_f64129a17ffe2smXGWD4Dt8yuU`. Refinamento/plano/produto/contexto/READMEs alinhados.
- Relatórios: `GRAPH_ENGINE_PROGRESS_INVOCATION_FIX.md` e `GRAPH_ENGINE_PROGRESS_INVOCATION_DOCS.md`.
- Checagem focada: `go -C engine test -run '^TestGraphInvocation(Decoder|Defaults|Schema)$' -count=1 .` passou. Sem I/O Git/harness/run.
- Build: `go -C engine build -o engine-next.exe .` passou. Binário separado porque `engine.exe` está em execução.
- Restart concluído por `ses_f648c5e94ffe2CrIX8CXYM4SU4`: engine anterior saiu graciosamente (exit 0); binário ativo **engine/engine-next.exe**, PID observado 34108, http://127.0.0.1:7331, health ok v2, zero turns/runs ativos. `engine.exe` ainda é o binário anterior; próximos reinícios devem usar/recompilar a versão corrigida.
- Ambiente/cwd/argumentos/data-dir preservados; comparação pública manteve 3 projetos, 27 sessões, 497 eventos. Relatório: `GRAPH_ENGINE_PROGRESS_INVOCATION_OPERATIONS.md`.
- Próximo: usuário recarrega http://127.0.0.1:5173/ e retesta na conversa. Não houve interação com navegador nem envio de chat/grafo pelo agente operacional.

## Fontes e regras

- Plano: `GRAPH_ENGINE_IMPLEMENTATION_PLAN.md` (P0–P6).
- Contratos: `GRAPH_AUTHORING_REFINEMENT.md`; contexto: `product.md`, `CONTEXT.md`, `AGENTS.md`.
- Preservar todas as alterações locais preexistentes de CRUD/UI/docs e artefatos Playwright.
- Sem commit/push, suítes completas, revisão adversarial ou execução de tarefas reais/modelos para validação automática.
- Verificação proporcional: build Go e typecheck frontend pelos responsáveis, quando o incremento estiver integrado.
- Pedido final: reiniciar a engine em execução e abrir o cliente web no Playwright.

## Responsabilidades e coordenação

O orquestrador mantém este arquivo. Cada subagente registra seu relatório em `GRAPH_ENGINE_PROGRESS_<AREA>.md`, incluindo arquivos, interfaces, checagens e limitações. Antes de modificar arquivo de outro responsável, devolver a necessidade ao orquestrador.

| Etapa | Responsável | Estado | Escopo |
| --- | --- | --- | --- |
| P0 | `ses_f666b6c91ffep5TvXEp61Jfy0I` | Implementado Windows; validação nativa completa pendente | Pi gate/settled, OpenCode plugin owned, Codex selo físico + leitura fria; limitações abaixo |
| P1 | `ses_f666b6c76ffe1Op1mPbYhi4PlN` | Implementado; build integrado passou | Contratos/schema/snapshot/persistência v2 e DTOs disponíveis |
| P2–P5 | core `ses_f666b6c76ffe1Op1mPbYhi4PlN` | Implementado; build integrado passou | Runtime, scheduler, ferramentas, Terminal, workspaces, Fork/Join |
| P6 | `ses_f64d7063dffe3wa38x8TI7R1Fl` | Implementado; typecheck passou | Catálogo/seleção/projeção/monitor/autoria/solicitações |
| Documentação | `ses_f64935357ffelICg2GZiKIJkEU` | Concluída | product/CONTEXT/READMEs alinhados; relatório DOCS |
| Entrega local | `ses_f648c5e94ffe2CrIX8CXYM4SU4` | Concluída | Engine reiniciada, migração/startup saudáveis e Playwright aberto |

## Fronteiras iniciais de arquivos

- Adapter: `harness.go`, `*_interactive.go`, `interactive.go`, `linked_bridge.go`, `platform_*.go`, extensão Pi e novos helpers de lifecycle. Não editar `main.go`/`store.go`/`api.go` sem coordenação; informar hooks necessários.
- Core: `authoring.go`, `store.go`, `main.go`, `api.go` e novos `graph_*.go` (exceto helpers de adapter acordados).
- Frontend: `clients/desktop/**` e seu relatório; nenhum Go.

## Interfaces / decisões de integração

- Interfaces exatas disponíveis nos relatórios `GRAPH_ENGINE_PROGRESS_CORE.md` e `GRAPH_ENGINE_PROGRESS_ADAPTER.md`.
- Core usa `BindGraphAdapter` antes de execute; somente callback Finished após aceitação permite continuação. Adapter pode oferecer factory para turn público, evitando edições concorrentes de execute.
- DTO frontend: `Session.graph` / GET e PATCH `/api/sessions/{id}/graph`; autoria contém output de Terminal e separate_worktree opcional default true.
- Slot de graph run separado dos turns de chat; sessões privadas com ownership explícito, nunca ParentID/Workspace visual.
- Nenhuma capacidade de finality será declarada comprovada apenas por abort/MCP/compilação.

## Verificações e retomada

- Baseline: `git status --short` consultado; muitos arquivos preexistentes modificados e não rastreados.
- Typecheck frontend `npx tsc --noEmit` passou; extensão Pi e plugin OpenCode tiveram checagens específicas aprovadas.
- Build core na segunda rodada falhou por `undefined: graphConversationHooks` enquanto adapter ainda implementava a interface. Símbolo agora publicado; próxima rodada verifica integração e build.
- Probes sem modelos: plugin OpenCode (gate/fail-closed), startup OpenCode com handshake/inventário, primitivas Job Object Windows com descendentes chegando a zero. Launcher Go integrado e Choice/continuação reais ainda não exercitados.
- OpenCode/Codex agora têm mecanismos Windows implementados e preflight admite capacidade; OpenCode exige versão inspecionada 1.18.30/plugin. MCPs externos incompatíveis são recusados antes de prompt nesses adapters. Unix continua sem contenção equivalente para nós.
- `RuntimeValidated=false` nos adapters: tarefas reais/Choice/continuação, Terminal/workspaces/Fork/Join e notificações/restart exigem validação humana. Não afirmar settlement de serviços externos/daemons por causa de Job Object.
- Integração final core: `go -C engine build` passou sem diagnósticos. Corrigido prompt completo de notificações pré-bindadas e correlação do Finished/outbox antes da liberação do WaitGroup.
- Documentação alinhada ao estado atual por subagente; relatório `GRAPH_ENGINE_PROGRESS_DOCS.md`.
- Operações concluídas: engine anterior encerrou graciosamente (exit 0); nova engine `engine/engine.exe` PID 29204 em http://127.0.0.1:7331. Vite preservado em http://127.0.0.1:5173/ (PID 8628).
- Data-dir preservado: `C:/Users/Samuel/AppData/Roaming/klm/engine`. Startup migrou v1→v2; health ok e comparação pública antes/depois preservou 3 projetos, 27 sessões e 479 eventos. Não foi alegada comparação das seções privadas do state.json.
- Playwright aberto no cliente; carregamento inicial sem erros/warnings de console, pageerror ou falhas HTTP observadas. Nenhum envio de chat, seleção ou execução de grafo.
- Relatório operacional: `GRAPH_ENGINE_PROGRESS_OPERATIONS.md` com argumentos, ownership, reinício e evidências.

## Retomada após esta entrega

1. Ler este documento e os relatórios CORE / ADAPTER / FRONTEND / OPERATIONS antes de repetir trabalho. Os relatórios CORE/ADAPTER contêm seções históricas: o estado vigente no topo prevalece.
2. Próxima aceitação é humana: Choice e continuação nos três harnesses, solicitações dos nós, Terminal/output, Fork/Join/workspaces, retries/notificações e restart durante run. Não executar automaticamente tarefas reais nem suítes completas.
3. P0 não está comprovado integralmente por build/probes: manter `RuntimeValidated=false` até haver evidência apropriada. Manter limites concretos Windows/OpenCode/MCP externo descritos acima.
4. Preservar alterações preexistentes e implementação não commitada. Nenhum commit/push realizado. Não repetir checagens aprovadas sem novas mudanças/falhas.
