# MCP externo — entrega operacional

Data: 2026-09-13. Startup às **16:19:54**, horário local (19:19:54 UTC).

## Estado entregue

- **Engine saudável:** http://127.0.0.1:7331 — PID **12204**.
- **Binário ativo:** `C:\Users\Samuel\Documents\Projects\Personal\klm\engine\engine-mcp.exe`, já compilado e aprovado pelo implementador.
- **Cliente preservado:** http://127.0.0.1:5173/ — Vite PID **8628**.
- Engine pronta para reload/reteste manual. Nenhum acesso ao navegador nesta operação.
- `engine.exe` e `engine-next.exe` não foram sobrescritos. Próximos reinícios devem considerar que o binário desta entrega é **engine-mcp.exe**, redescobrindo os processos vivos em vez de reutilizar os PIDs deste relatório.

## Descoberta e configuração

Listener e processo alvo redescobertos com `Get-NetTCPConnection`, `Get-CimInstance Win32_Process` e `GetOwner`. Cwd, argumentos e ambiente lidos dos parâmetros do processo Windows. Nenhum PID foi obtido de arquivos de estado/lock.

| Campo | Antes | Depois |
| --- | --- | --- |
| PID | 34108, confirmado no listener vivo | 12204 |
| Binário | `engine/engine-next.exe` | `engine/engine-mcp.exe` |
| Criação UTC | 2026-09-13T18:15:55.8330220Z | 2026-09-13T19:19:54.3178930Z |
| Owner | `DESKTOP-50TK4D0\Samuel` | Igual |
| Addr | `127.0.0.1:7331` | Igual |
| Cwd | `C:\Users\Samuel\Documents\Projects\Personal\klm\engine\` | Igual |
| Data-dir explícito | `C:\Users\Samuel\AppData\Roaming\klm\engine` | Igual |
| Ambiente | 79 entradas; sem overrides `KLM_*` | Igualdade integral confirmada em memória |

Comando atual:

```powershell
C:\Users\Samuel\Documents\Projects\Personal\klm\engine\engine-mcp.exe -addr 127.0.0.1:7331 -data-dir C:\Users\Samuel\AppData\Roaming\klm\engine
```

Argumentos extraídos com `CommandLineToArgvW` e repassados, substituindo apenas o executável. Valores do ambiente não foram exibidos ou persistidos. `Test-Path` confirmou data-dir, state existente, novo binário, prompts e diretórios pais dos artefatos antes da operação.

Vite preservado: `node node_modules/vite/bin/vite.js --host 127.0.0.1 --port 5173 --strictPort`, cwd `clients/desktop`, owner Samuel, pai 16008. `VITE_ENGINE_URL` ausente; mantém o endpoint padrão da engine.

## Encerramento e startup

1. Health e estado público consultados antes do shutdown: **runningSessions=0**, **activeGraphRuns=0**, nenhuma sessão pública running.
2. Console verificado por `AttachConsole` / `GetConsoleProcessList`: somente engine descoberta e helper temporário. O helper exige console exclusivo da engine; não usa PID alvo hardcoded.
3. Processo, criação, owner e listener revalidados imediatamente antes de sinalizar. Health ocioso também reconsultado imediatamente antes do sinal.
4. **`CTRL_BREAK_EVENT`** enviado ao console exclusivo; engine anterior encerrou graciosamente com **exit code 0**, confirmado por handle. Nenhum force-kill.
5. Novo processo criado com `CREATE_NEW_CONSOLE`, mesmo cwd/ambiente/argumentos e stdio separado. Log: `2026/09/13 16:19:54 KLM engine listening on http://127.0.0.1:7331`.
6. Listener confirmado no PID 12204 e executável confirmado como `engine-mcp.exe`. Após o helper retornar, nova consulta confirmou processo em background atendendo health e state normalmente.

## HTTP e comparação pública

Health final:

```json
{"activeGraphRuns":0,"runningSessions":0,"status":"ok","version":2}
```

| Item | Antes | Depois |
| --- | ---: | ---: |
| Projetos | 3 | 3 |
| Sessões | 29 | 29 |
| Eventos | 542 | 542 |
| Sessões running | 0 | 0 |

- Objetos públicos completos de projetos: **iguais**.
- Objetos públicos completos de sessões: **comparação retornou diferente**. Portanto, não se afirma igualdade integral das sessões nem de seus eventos apenas pelas contagens.
- O helper encerrou com a asserção diagnóstica `State differences detected; report for investigation`, **depois** de iniciar e verificar a nova engine. A engine continuou saudável, como confirmado por HTTP/listener em chamada independente.
- Diagnóstico limitado do caminho de shutdown: `engine/main.go:180–183` faz commit quando existe histórico de GraphRuns, mesmo sem run ativo; `engine/store.go:240` incrementa `GraphRevision` nesse commit. A projeção pública usa essa revisão, podendo mudar objetos de sessão sem alterar mensagens. Isso é uma explicação compatível com a diferença, **não prova de que foi o único campo alterado**.
- O snapshot anterior existia apenas na memória do helper e não foi retido após seu encerramento. Não foi possível produzir um diff campo a campo desta operação. Nenhum dump de mensagens privadas foi gravado; nenhum novo restart foi feito apenas para repetir a comparação.
- Estado permaneceu v2; startup aceitou os dados existentes. Nenhum erro de carregamento/persistência observado. Não houve manipulação direta de state/lock nem fallback para diretório vazio.

## Artefatos e limites

- Relatório novo: `GRAPH_ENGINE_PROGRESS_EXTERNAL_MCP_OPERATIONS.md`. Progresso principal e relatórios dos implementadores preservados.
- Helper adaptado via `apply_patch`: `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-operations-20260913.py`.
- Logs desta operação:
  - `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-engine-external-mcp-20260913.stdout.log`
  - `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-engine-external-mcp-20260913.stderr.log`
- Nenhuma alteração de código/configuração da aplicação, build ou teste nesta entrega. Nenhum navegador, chat, run ou modelo acionado.
- Evidência cobre restart ocioso, startup, ambiente/argumentos preservados e health/state. A correção de lifecycle MCP/Choice será validada pelo usuário no reteste manual. A diferença na comparação completa das sessões permanece registrada acima para o orquestrador.
