# Graph engine — entrega operacional

Atualizado: 2026-09-13, após startup às 13:55:43 (horário local).

## Estado entregue

- Engine reiniciada e saudável: **http://127.0.0.1:7331**, PID **29204**.
- Cliente existente reutilizado: **http://127.0.0.1:5173/**, Vite PID **8628**.
- Playwright permanece aberto na aba KLM, título **KLM | Agent Workspace**.
- Nenhum bloqueio de startup, migração ou carregamento inicial encontrado.

## Processos, argumentos e ownership

Ownership foi consultado em processos vivos via `Get-CimInstance Win32_Process` / `GetOwner`, e relacionado aos listeners via `Get-NetTCPConnection`. Nenhum PID foi obtido de `engine.lock` ou de estado persistido.

| Papel | PID | Comando / contexto |
| --- | --- | --- |
| Engine anterior | 10844 | `C:\Users\Samuel\AppData\Local\go-build\07\0776ca9378c7b05afffc666ae752d5cce836464ed72074a063fa14ae6b5cec7f-d\engine.exe`, sem flags |
| Launcher anterior | 31000 | `go run .`, pai 36640 (`cmd.exe /k "go run ."`) |
| Engine atual | 29204 | `C:\Users\Samuel\Documents\Projects\Personal\klm\engine\engine.exe -addr 127.0.0.1:7331 -data-dir C:\Users\Samuel\AppData\Roaming\klm\engine` |
| Vite preservado | 8628 | `node node_modules/vite/bin/vite.js --host 127.0.0.1 --port 5173 --strictPort`, pai 16008 (`cmd.exe`) |

- Usuário confirmado dos processos anteriores: `DESKTOP-50TK4D0\Samuel`.
- Cwd real da engine anterior e da nova: `C:\Users\Samuel\Documents\Projects\Personal\klm\engine`.
- Cwd real do Vite: `C:\Users\Samuel\Documents\Projects\Personal\klm\clients\desktop`.
- **Data-dir preservado:** `C:\Users\Samuel\AppData\Roaming\klm\engine`. Derivado do `APPDATA` do processo anterior, que não usava flags. Diretório e `state.json` existentes foram confirmados antes do shutdown.
- Ambiente real da engine anterior lido dos parâmetros do processo Windows e copiado diretamente em memória ao novo processo: 79 entradas, sem overrides `KLM_*`. Valores do ambiente não foram gravados em arquivos ou exibidos.
- `VITE_ENGINE_URL` ausente no processo Vite: usa o endpoint padrão `http://127.0.0.1:7331`, confirmado pelas requisições do navegador.
- Binário usado foi o `engine/engine.exe` já compilado pela implementação. `engine/prompts/` existente confirmado; nenhum override necessário.

## Shutdown, startup e migração

1. Antes do shutdown, HTTP health retornou `status=ok`, `version=1`, `runningSessions=0`.
2. Console da engine inspecionado por `AttachConsole` / `GetConsoleProcessList`: somente engine 10844, launcher Go 31000 e seu terminal 36640, além do helper temporário. Todos os processos alvos foram identificados. Ownership, comando, criação e listener da engine foram revalidados imediatamente antes do sinal.
3. Enviado **`CTRL_C_EVENT`** ao console próprio, com o helper protegido contra o sinal. Engine anterior terminou com **exit code 0**, confirmado por handle do processo. Nenhum force-kill foi necessário.
4. Nova engine iniciada em processo separado com `CREATE_NEW_CONSOLE`, mesmo cwd e ambiente, endereço e data-dir explicitados. O launcher operacional retornou; a engine continuou ouvindo no próprio PID.
5. Startup registrou `KLM engine listening on http://127.0.0.1:7331`. Poll HTTP confirmou versão 2 e o listener confirmou PID 29204.
6. Migração v1 → v2 aceita no startup sobre o estado existente. Evidência: health anterior v1, novo startup saudável v2 e estado público preservado; `loadState` em `engine/store.go` persiste a migração antes de liberar o startup. Não foi feito dump do arquivo privado nem comparação de suas seções internas.

### Evidência HTTP

Health final:

```json
{"activeGraphRuns":0,"runningSessions":0,"status":"ok","version":2}
```

Comparação de `/api/state` antes/depois, realizada em memória sem imprimir conteúdo das conversas:

| Item | Antes | Depois |
| --- | ---: | ---: |
| Projetos | 3 | 3 |
| Sessões | 27 | 27 |
| Eventos | 479 | 479 |
| Sessões em execução | 0 | 0 |

- Objetos públicos de projetos: iguais.
- Objetos públicos de sessões: iguais, desconsiderando apenas o novo campo de projeção `graph`.
- Nenhum diretório vazio alternativo, reset, remoção de lock ou escrita manual de estado foi usado.

## Evidência Playwright

- Aba KLM existente recarregada em `http://127.0.0.1:5173/`; documento HTTP **200**.
- Espera limitada pelo botão **New session** visível concluiu normalmente.
- Snapshot inicial de acessibilidade capturado com profundidade 3: rail **Projects**, sidebar **Sessions**, navegação **New session / Agents / Graphs**, área principal, **Conversation**, **Message composer** e **Session Status Bar** presentes.
- Navegação persistida do navegador restaurou o projeto **Authoring validation**. Não foi selecionado outro projeto, sessão ou grafo.
- Console da navegação atual: **0 errors, 0 warnings** (3 mensagens totais).
- Instrumentação durante o carregamento inicial: nenhuma exceção `pageerror`, requisição falha ou resposta HTTP >= 400 observada naquele carregamento.
- Recursos HTTP da engine observados depois de carregar: `/api/state`, seleção de grafo da sessão, catálogo de autoria e consultas automáticas de modelos/quota do cliente, todos com status **200**. Nenhuma consulta adicional de modelos foi disparada manualmente.
- Aba KLM deixada aberta. As demais abas existentes foram preservadas.

## Artefatos e limites

- Arquivo de entrega no repositório: `GRAPH_ENGINE_PROGRESS_OPERATIONS.md`.
- Helper operacional e logs somente no temp preaprovado:
  - `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-operations-20260913.py`
  - `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-engine-20260913.stdout.log`
  - `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-engine-20260913.stderr.log`
- `Test-Path` dos pais executado antes de criar esses arquivos e este relatório. O helper é específico desta operação, com verificações do processo anterior; não é um script genérico para reinícios futuros.
- Nenhuma configuração ou código de aplicação alterado; progresso principal preservado.
- Build/typecheck já aprovados não foram repetidos. Nenhuma suíte, revisão, tarefa real, envio de chat, seleção de grafo ou execução de grafo realizada.
- Esta evidência cobre encerramento ocioso, startup/migração do estado real e carregamento básico. Execução nativa, Choice/continuação, Terminal/workspaces/Fork/Join, solicitações e restart durante execução continuam para validação humana conforme os relatórios de implementação.
