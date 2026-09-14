# Graph invocation — entrega operacional

Data: 2026-09-13. Startup confirmado às **15:15:56**, horário local (18:15:56 UTC).

## Estado entregue

- **Engine saudável:** http://127.0.0.1:7331 — PID **34108**.
- **Binário em execução:** `C:\Users\Samuel\Documents\Projects\Personal\klm\engine\engine-next.exe`, produzido pelo build aprovado em `GRAPH_ENGINE_PROGRESS_INVOCATION_FIX.md`.
- **Cliente preservado:** http://127.0.0.1:5173/ — Vite PID **8628**.
- Pronta para o usuário recarregar a aba e repetir o teste. Nenhuma interação com o navegador nesta entrega.
- **Atenção para a próxima operação:** o processo atual executa `engine-next.exe`. `engine.exe` não foi sobrescrito e continua sendo o binário anterior. Redescobrir listener/processo vivo em futuros reinícios, sem confiar nos PIDs históricos deste relatório.

## Processo real e configuração preservada

Listener, executável, comando, owner e timestamp de criação foram consultados em processos vivos via `Get-NetTCPConnection`, `Get-CimInstance Win32_Process` e `GetOwner`. Cwd, argumentos e ambiente foram lidos dos parâmetros do processo Windows; nenhum PID veio de `state.json` ou `engine.lock`.

| Campo | Antes | Depois |
| --- | --- | --- |
| PID | 29204, redescoberto pelo listener | 34108 |
| Executável | `engine/engine.exe` | `engine/engine-next.exe` |
| Criação UTC | 2026-09-13T16:55:42.8996030Z | 2026-09-13T18:15:55.8330220Z |
| Owner | `DESKTOP-50TK4D0\Samuel` | `DESKTOP-50TK4D0\Samuel` |
| Addr | `127.0.0.1:7331` | Igual |
| Cwd | `C:\Users\Samuel\Documents\Projects\Personal\klm\engine\` | Igual |
| Data-dir explícito | `C:\Users\Samuel\AppData\Roaming\klm\engine` | Igual |
| Ambiente | 79 entradas, sem overrides `KLM_*` | Igualdade integral confirmada em memória |

Comando atual:

```powershell
C:\Users\Samuel\Documents\Projects\Personal\klm\engine\engine-next.exe -addr 127.0.0.1:7331 -data-dir C:\Users\Samuel\AppData\Roaming\klm\engine
```

Os argumentos foram extraídos do comando vivo com `CommandLineToArgvW` e repassados, trocando apenas o executável. Ambiente foi copiado em memória e comparado no novo processo, sem imprimir ou persistir seus valores. Diretório de dados, `state.json`, binário e `engine/prompts/` existentes foram confirmados com `Test-Path` antes de parar a engine.

Vite mantido com `node node_modules/vite/bin/vite.js --host 127.0.0.1 --port 5173 --strictPort`, cwd `clients/desktop`, pai 16008 e owner Samuel. `VITE_ENGINE_URL` ausente: endpoint padrão preservado.

## Encerramento e startup

1. Health antes da parada: `status=ok`, `version=2`, **runningSessions=0**, **activeGraphRuns=0**. `/api/state` também não apresentou sessões running. Os valores foram reconsultados imediatamente antes de sinalizar.
2. Console inspecionado por `AttachConsole` / `GetConsoleProcessList`: exclusivo da engine descoberta, além do helper temporário. O helper não contém PID alvo hardcoded; exige que o console pertença somente ao processo obtido do listener. Processo completo, criação, owner e listener revalidados antes do sinal.
3. Primeira tentativa com `CTRL_C_EVENT` não encerrou em 25 segundos. Nenhum replacement foi iniciado nessa tentativa. Listener e health continuaram disponíveis e ociosos. A causa do Ctrl+C não ter efeito não foi comprovada.
4. Nova tentativa com **`CTRL_BREAK_EVENT`**, após repetir descoberta/ownership/health, encerrou graciosamente com **exit code 0**, confirmado por handle do processo. Nenhum force-kill foi usado.
5. `engine-next.exe` iniciado com `CREATE_NEW_CONSOLE`, stdio separado e logs no temp. O helper protege a si mesmo ao sinalizar e limpa a opção de ignorar Ctrl+C antes de iniciar o novo processo.
6. Log: `2026/09/13 15:15:56 KLM engine listening on http://127.0.0.1:7331`.
7. Health e listener confirmados durante startup e novamente após o launcher retornar: PID 34108 saudável, executando o caminho correto em background.

## Saúde e preservação de dados

Health final HTTP:

```json
{"activeGraphRuns":0,"runningSessions":0,"status":"ok","version":2}
```

Comparação em memória de `/api/state`, sem dump das conversas:

| Item | Antes | Depois |
| --- | ---: | ---: |
| Projetos | 3 | 3 |
| Sessões | 27 | 27 |
| Eventos | 497 | 497 |
| Sessões running | 0 | 0 |

- Objetos públicos completos de projetos e sessões, incluindo projeção `graph`: **iguais**.
- Estado permaneceu v2. Startup aceitou os dados existentes; nenhum erro de carregamento/persistência observado.
- Não houve leitura completa ou comparação das seções privadas do arquivo persistido. Nenhuma manipulação direta de state/lock, reset ou diretório vazio alternativo.

## Artefatos e limites

- Novo arquivo no repositório: `GRAPH_ENGINE_PROGRESS_INVOCATION_OPERATIONS.md`. Relatórios de implementação e progresso principal não foram editados.
- Helper anterior adaptado via `apply_patch`: `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-operations-20260913.py`.
- Logs desta operação:
  - `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-engine-invocation-20260913.stdout.log`
  - `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-engine-invocation-20260913.stderr.log`
- Pais dos novos arquivos confirmados com `Test-Path`. Nenhum ajuste de código ou configuração da aplicação foi necessário.
- Build/testes não repetidos. Nenhuma recarga de navegador, envio de chat, consulta manual de modelo ou execução de grafo.
- Evidência limitada a restart ocioso, startup, configuração preservada e health/state. O usuário fará o reteste funcional de graph invocation após recarregar a aba.
