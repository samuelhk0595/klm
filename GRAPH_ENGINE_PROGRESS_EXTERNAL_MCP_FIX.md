# Graph engine — correção do limite de chamadas MCP externas

Data: 2026-09-13. Código integrado e binário `engine/engine-mcp.exe` gerado.
Restart fica com o agente operacional. Não alterei relatórios anteriores, README,
documentos canônicos, configurações MCP do usuário ou o estado da engine em execução.

## Problema e contrato aplicado

A run `b9307f7f23e21b42cc3a5c75cb9eab92` falhou no ban genérico de servidor externo
OpenCode (`agentdeck`). Foram removidos os bans de configuração de OpenCode e Codex.

Conforme a correção explicitamente aprovada pelo usuário, a garantia é sobre o nó
owned e suas **chamadas de ferramenta**, não a vida do servidor nem a de tarefas
remotas destacadas. Uma resposta "task started" encerra aquela chamada; a engine
não passa a acompanhar/cancelar a tarefa destacada. A existência de servidor MCP
configurado/habilitado não impede iniciar o nó.

Reserva da Choice continua distinta de aceitação. Resposta da própria ferramenta
Choice é devolvida sem aguardar a si mesma. A continuação só pode ser aceita depois
das respostas das chamadas acompanhadas e do encerramento owned exigido pelo adapter.

## Arquivos desta correção

- Novos: `engine/graph_mcp_calls.go`, `engine/graph_mcp_calls_test.go`.
- Alterados: `engine/graph_adapter.go`, `engine/graph_runtime.go`,
  `engine/opencode_interactive.go`, `engine/codex_interactive.go`,
  `engine/opencode-graph-plugin.mjs`.
- Novo relatório: este arquivo.
- Binário de entrega: `engine/engine-mcp.exe`.

`graph_invocation_contract.go` e as últimas correções de schemas, autorização,
default de workspace, fresh/reuse e provenance foram lidos e preservados.
Não houve alteração de sandbox, grants, policies de aprovação, gate de versão
OpenCode ou configuração dos serviços externos.

## Mecanismo implementado

### Rastreador por ativação

Cada binding de ativação possui um mapa de chamadas. A chave OpenCode contém
sessão nativa + call ID; Codex acrescenta o turn ID. Chamadas de filhos e chamadas
paralelas não colapsam quando reutilizam IDs locais. O binding não é compartilhado
com outra ativação ou nova correção de turn.

Estados internos:

| Estado | Evidência |
| --- | --- |
| `admitted` | Admitida no gate ou observada no protocolo; pode ainda estar esperando permissão. Não é afirmação de bytes enviados. |
| `response` | Callback de retorno real ou envelope nativo de resposta. Inclui uma resposta que apenas comunica início de tarefa remota. |
| `not_started` | Before lançou erro antes de executar, ou rejeição de permissão foi entregue com correlação exata antes do envio MCP. |
| `uncertain` | Falha/aborto local sem prova da resposta. Não é cancelamento remoto confirmado. |

Uma resposta real tardia pode resolver uma dúvida local anterior enquanto o adapter
ainda está encerrando; notificações duplicadas/atrasadas não desfazem uma resposta
já comprovada. Isso não constitui garantia distribuída de exatamente uma vez.

### OpenCode

- O plugin owned admite o nome realmente resolvido pelo harness em
  `tool.execute.before`. O inventário nativo identifica ferramentas locais; nomes
  MCP/custom/resource restantes são acompanhados conservadoramente como chamadas
  externas. Não há allowlist de nomes de servidores.
- O status real retornado ao conectar o MCP owned também informa nomes de servidores
  descobertos. Sua normalização segue `McpCatalog.toolName` da versão 1.18.30;
  isso evita que colisão com ID de ferramenta nativa esconda uma chamada MCP.
  Servidor configurado, inclusive `agentdeck`, não é motivo de recusa.
- Admissão e selamento compartilham o mutex. Before rejeitado depois do selo não
  cria entrada. Se o HTTP de admissão falhar, o plugin lança erro antes de executar
  e tenta informar `not_started`; ausência de confirmação permanece conservadora.
- `tool.execute.after` comunica `response`. Essa operação continua permitida depois
  do selo, para drenar chamadas já admitidas. Não constitui nova chamada de trabalho.
- Eventos nativos de erro são observados **antes** do filtro de UI da sessão raiz,
  incluindo subtasks. Rejeições de permissão correlacionadas com `tool.callID`
  podem comprovar que a chamada ainda não foi enviada.
- Solicitações MCP de sessões filhas conhecidas mantêm o fluxo de perguntas/permissões
  da sessão de nó, sem grants artificiais. A raiz não é a única sessão observada.
- A Choice não libera o interrupt enquanto existem chamadas admitidas pendentes.
  Se a raiz terminar normalmente primeiro, o adapter continua recebendo conclusões
  dos filhos antes de fechar o processo ou aceitar a Choice.

**Limite concreto do protocolo:** na tag v1.18.30, `McpCatalog.convertTool` lança
exceção quando recebe `isError`; `session/tools.ts` só executa o after quando recebe
um valor retornado. Não existe after garantido no caminho de exceção. Um `state.error`
com texto não distingue resposta remota de erro de transporte/abort local. Sem
outra evidência, essa chamada fica incerta — não inventei um callback de erro nem
interpretei qualquer texto de erro como resposta MCP.

### Codex

- Removido o ban durante `mcpServerStatus/list`; a verificação de readiness das
  ferramentas owned e a paginação continuam funcionando com servidores externos.
- Eventos `item/started` e `item/completed` de MCP são observados antes do filtro
  raiz/thread/turn usado pela UI, inclusive enquanto o processo está sendo drenado.
  Frames já enfileirados após Done também são consumidos, não descartados.
- `completed`/`failed` com envelope `result.content` comprova resposta, inclusive
  `isError=true`. Status cancelled/failed com somente erro, ou fechamento do processo,
  não comprova que o serviço respondeu/cancelou a chamada.
- Interrupt por Choice aguarda o conjunto conhecido de chamadas pendentes. Uma
  corrida que introduza outra chamada antes do selo físico continua sendo observada;
  não transforma cancelamento dessa chamada em resposta.
- A verificação fria já existente agora também inspeciona threads filhas identificadas
  nos itens `collabAgentToolCall.receiverThreadIds`. Não inicia prompts/modelos.
  Historicamente retornados e atuais são correlacionados por thread/turn e timestamps;
  thread nova do subagent pertence à ativação. Ausência de cobertura, histórico
  inacessível ou item sem resposta em thread legada não claramente correlacionável
  conserva incerteza em vez de anunciar drenagem.
- A vida de tarefas remotas mencionadas no resultado não entra nesse conjunto.
  O Job continua encerrando apenas processos owned; isso não é a prova da resposta MCP.

### Pi

O gate e a sequência sequencial/settled existentes foram preservados. Não havia ban
genérico de configuração MCP nesse adapter. A extensão owned desta integração não
carrega extensões externas arbitrárias; o novo resultado não promove término owned
a confirmação de trabalho remoto. Chamadas de shell e tarefas destacadas não ganham
uma promessa de cancelamento de serviços externos.

## Propagação de incerteza

`GraphAdapterResult` passa a incluir `ToolCallsSettled` e `UncertainToolCalls`.
`ProcessesDrained` continua descrevendo a drenagem owned, separadamente.

Quando faltar resposta/cobertura, o resultado carrega `graphUnconfirmedError`, não
aceita a Choice e não permite continuação. `graph_runtime.go` usa a política existente
de `ending`/`FinalityError` para reter o slot e persiste o motivo com IDs de
ativação/sessão/turn/chamada nos eventos da ativação. O diagnóstico agora descreve
incerteza de chamada, em vez de atribuí-la incorretamente à morte de um servidor.
Payloads, credenciais e resultados integrais não são copiados para esse diagnóstico.

Nada foi reexecutado nem alterado retroativamente na run relatada. A correção será
utilizada em novas execuções depois do restart coordenado.

## Evidência e checagens

Fontes inspecionadas, sem executar modelos:

- OpenCode v1.18.30: `packages/opencode/src/mcp/catalog.ts`, sobretudo a chamada
  `client.callTool`, conversão de `isError` em throw e normalização de nomes;
  ordem dos hooks em `session/tools.ts`/`plugin/index.ts` verificada no trabalho P0.
- Handlers nativos instalados no repo e contratos da versão Codex 0.153.4 já
  inspecionados no P0; foram preservados EOF, leitura fria e validação nativa.
- Código atual de permissões confirma que replies ocorrem fora de `app.mu`;
  o novo bookkeeping não adiciona inversão desse mutex com o do binding.

Checagens executadas nesta correção:

1. `gofmt -w` somente nos arquivos Go alterados.
2. TypeScript no novo caminho do plugin:
   `node C:/Users/Samuel/AppData/Roaming/npm/node_modules/typescript/bin/tsc --noEmit --allowJs --checkJs --strict --skipLibCheck --module esnext --moduleResolution bundler --target es2022 engine/opencode-graph-plugin.mjs`
   — terminou sem erros.
3. `go -C engine test -run '^TestGraphMCPBoundary' -count=1 .`
   — última execução: `ok klm/engine 0.610s`.
   Cinco testes focados cobrem: paralelo/subtask/Choice sem autoespera, recusa no
   before, erro/transport/cancel sem falsa confirmação, resposta de tarefa destacada,
   erro remoto com envelope real, duplicatas, saída com chamada pendente, leitura
   fria e cobertura ausente de filho, e colisão de nome nativo/MCP descoberto.
   Houve novas execuções apenas após mudanças de código/casos; nenhuma suíte completa.
4. `Test-Path -LiteralPath C:/Users/Samuel/Documents/Projects/Personal/klm/engine`
   — `True`.
5. **Build final executado uma única vez**:
   `go -C engine build -o engine-mcp.exe .`
   — terminou com sucesso, sem diagnósticos.

Binário absoluto:
`C:/Users/Samuel/Documents/Projects/Personal/klm/engine/engine-mcp.exe`.

Não rodei grafos reais, modelos, suites completas, review adversarial, commit ou
restart. O processo `engine-next.exe` não foi substituído por esta sessão. O agente
operacional deve reiniciar com o binário acima; a validação seguinte é o reteste
humano com os MCPs configurados, incluindo Choice enquanto há chamadas paralelas.
