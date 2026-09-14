# Graph engine — DOCUMENTAÇÃO

Atualizado: 2026-09-13. Documentação de execução e monitor mínimo atualizada;
checagem por leitura dirigida, sem testes ou execução de modelos.

## Arquivos atualizados

- `product.md`: execução assíncrona autorizada, atividades/runs, snapshots,
  sessões privadas, Choice, Terminal/Fork/Join, workspaces e reinício; seleção
  persistida e monitor mínimo real. Escopo e decisões anteriores preservados.
- `CONTEXT.md`: atividade versus run, seleção, Terminal/output, Fork/isolation,
  blocked interno e vocabulário de run ativa/Join coletando. Removida a antiga
  pendência de refinamento do contrato de Fork.
- `README.md`: uso do catálogo real, invocação pelo orquestrador, execução e
  acompanhamento atuais, limites de adapters e links aos relatórios vigentes.
- `clients/desktop/README.md`: catálogo compartilhado, GET/PATCH da seleção,
  projeção/SSE/polling, snapshot, ativos/concluídos, Collecting, LED no seletor,
  requests com sessão originadora e roundtrip de output/isolation.

As antigas descrições de execução futura, catálogo fictício e simulação produtiva
foram substituídas. O histórico CSV/Run/Input/Output ficou resumido como experimento
superado pela seção 12 de `GRAPH_AUTHORING_REFINEMENT.md`. Galeria estática e demais
limites não relacionados continuam identificados como tais.

## Estado e limites registrados

- Pi, OpenCode e Codex têm mecanismos Windows implementados e admitidos pelo
  preflight de capacidade; **`RuntimeValidated=false`** nos três. Isso não comprova
  tarefa real, Choice nativa, finality integrada ou continuação com modelo.
- OpenCode exige **1.18.30** e plugin/handshake owned. OpenCode/Codex recusam MCP
  externo incompatível antes do prompt do nó, sem desabilitá-lo silenciosamente.
- Contenção Windows de processos owned não comprova término de serviços remotos
  ou trabalho delegado externamente. Execução de grafos em Unix não está habilitada.
- None/abrir Graph não inicia nem cancela trabalho; requests persistem na projeção.
  O monitor usa snapshot durante a run correspondente e configuração real ociosa
  depois, sem histórico, painéis de execução ou conclusão fictícia de todos os nós.

## Checagem por leitura dirigida

- Lidos `AGENTS.md`, `product.md`, `CONTEXT.md`, plano completo e os três relatórios
  CORE/ADAPTER/FRONTEND; conferidas as decisões mais recentes das seções 11–12 do
  refinamento. O estado vigente da segunda rodada ADAPTER prevalece sobre as
  recusas históricas de OpenCode/Codex ainda citadas no relatório CORE.
- Inspecionados os diffs preexistentes dos quatro documentos antes dos patches,
  preservando os acréscimos de CRUD e decisões do usuário.
- Leitura focal de `engine/graph_adapter.go` confirmou a declaração do hook acordado
  e os campos/preflight de capacidade, com `RuntimeValidated` sem promoção a true.
  Isso não revalida o build que o CORE registrou como bloqueado em sua rodada.
- Leitura focal de `GraphView.tsx` confirmou desenho real, prioridade active/collecting/
  completed e configuração apenas no modo ocioso.
- Relidas as seções alteradas dos quatro documentos; busca dirigida conferiu que
  referências a fixtures/experimentos não descrevem o estado produtivo atual e que
  versão, plataforma e limites de validação estão explícitos.
- Nenhum teste, build, typecheck, modelo, comando de grafo ou review executado por
  este implementador. O typecheck citado no README desktop é evidência atribuída
  ao relatório FRONTEND, não uma nova checagem desta rodada.

Validação humana restante: fluxo integrado de Choice/continuação nos três harnesses,
Terminal/workspaces/Fork/Join, notificações/restart e seleção/reload/None/requests no
monitor. A documentação registra implementação, não aprovação desses fluxos.

Este implementador alterou somente os quatro documentos solicitados e este relatório.
Código, `engine/README.md`, plano, relatórios CORE/ADAPTER/FRONTEND e progresso
principal não foram editados. Sem commit ou push.
