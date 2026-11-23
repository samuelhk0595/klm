# Plano de implementação — 0.6.2

Status: planejado; nenhuma correção implementada por este documento.

## 1. Objetivo e decisões do usuário

Entregar execução de graphs no Linux e corrigir as diferenças de comportamento que prejudicam automações e conversas comuns. A referência Linux desta entrega é o engine headless em Linux amd64, com a base de instalação glibc já suportada. Preservar o funcionamento do Windows.

Decisões confirmadas nesta conversa:

1. Exibir a causa correta das falhas de inspeção do OpenCode, com diagnóstico suficiente para uma próxima investigação.
2. Remover a trava especial de mensagens das automações: seu composer, fila e Send now devem se comportar como os de uma conversa comum.
3. Usar uma fonte consistente de autenticação do `gh` em sessões manuais e automações.
4. Tratar a seleção de graphs na configuração da automação como autorização de execução, dispensando nova autorização por mensagem para esses graphs.
5. Implementar suporte completo aos fluxos existentes de execução de graphs no Linux, incluindo agentes, Terminal, Fork/Join e workspaces.
6. Implementar armazenamento protegido de credenciais MCP no Linux.
7. Corrigir o caminho de subagentes nativos do Pi no Linux.
8. Implementar seletor nativo de pastas no Linux quando houver ambiente gráfico.
9. Adicionar Bash e sh aos nós Terminal, preservando graphs PowerShell existentes.

Os itens 6–8 correspondem aos pontos 1, 3 e 4 da última lista de lacunas apresentada ao usuário. Não confundir essa numeração com a lista original de problemas de automação.

### Limites desta versão

- Não incluir a integração autenticada do Activity Board no Linux, cliente desktop Linux, instaladores Linux equivalentes ao NSIS, hospedagem web pelo engine, ARM64 ou suporte de instalação Alpine/musl.
- Não adicionar sandbox de sistema, cgroups obrigatórios, privilégios root ou dependência de sessão gráfica para execução headless.
- Não ampliar autorização de graphs para criação de sessões, instalação de skills, mudanças de Library ou permissões de ferramentas.
- Não executar deploy, instalar dependências na VPS, migrar sua base real ou gerar/publicar releases durante a implementação sem solicitação específica.
- O plano não agenda graphs nem autoriza a execução dos graphs do projeto nesta conversa.

## 2. Evidências e pontos de alteração

| Frente | Causa encontrada | Pontos principais |
| --- | --- | --- |
| Diagnóstico OpenCode | Qualquer erro do subprocesso de inspeção vira a mensagem sobre Node; resolução usa PATH diretamente | `engine/opencode_library.go`, `models.go`, `opencode_interactive.go`, `harness_discovery.go` |
| Conversas de automação | Frontend força steering; backend bloqueia envio imediato e despacho da fila; servidor OpenCode retido é confundido com finalização não confirmada | `clients/desktop/src/App.tsx`, `features/chat/MessageQueue.tsx`, `engine/sqlite_execution_queue.go`, `sqlite_execution_finish.go`, `sqlite_automations_worker.go`, `opencode_interactive.go` |
| Autenticação gh | Perfil OpenCode substitui XDG_CONFIG_HOME, redirecionando também a busca de configuração do gh | `engine/opencode_library.go`, `opencode_interactive.go`, `models.go` e ambientes dos demais adaptadores |
| Autorização de graph | Lista da automação apenas restringe graphs; prompt e backend continuam exigindo evento humano | `engine/prompts/automation-run.md`, `prompts/orchestrator.md`, `graph_invocation_contract.go`, `sqlite_graph_tools.go`, `sqlite_graph_scheduler.go` |
| Graphs Linux | Capacidade de controle de processos é anunciada apenas no Windows | `engine/graph_adapter.go`, `platform_linux.go`, `linux_supervisor.go`, `harness_discovery.go` |
| Shells | Terminal só resolve PowerShell e não persiste seleção de shell | `engine/graph_terminal.go`, `authoring.go`, `graph_contract.go`, `graph_snapshot*.go`, `clients/desktop/src/features/graphs/` |
| Credenciais MCP | Implementação não Windows retorna erro para proteção/desproteção | `engine/mcp_secrets_other.go`, `mcp_connections.go`, `mcp_oauth.go` |
| Pi nativo | Caminho pi-subagents recusa Linux explicitamente | `engine/pi_subagents.go`, `pi_interactive.go`, `pi-subagent-factory.mjs` |
| Seletor de pastas | Capacidade nativa é constante false fora do Windows | `engine/directory_browser_unix.go`, `directory_browser.go`, `platform_linux.go` e endpoint do picker |

As causas específicas do erro Node e da diferença de ambiente entre duas sessões na VPS não foram reproduzidas. As evidências acima são de inspeção do código; não equivalem a validação de runtime.

## 3. Diagnóstico correto da inspeção OpenCode

### Implementação

- Passar à inspeção a identidade de Node resolvida para o harness, inclusive nas consultas ao catálogo. Usar a resolução comum quando essa identidade não estiver disponível; não procurar silenciosamente outro Node por um caminho independente.
- Separar disponibilidade de Node, disponibilidade de `node:sqlite`, abertura do banco, compatibilidade de schema, consulta de organização, timeout e cancelamento.
- Manter a inspeção read-only e a recusa a configurações de organização não suportadas. Ausência do banco continua sendo um caso próprio, sem fingir que houve inspeção.
- O script deve produzir um resultado estruturado com códigos conhecidos. Erros inesperados de processo devem informar etapa, exit code/sinal quando disponível e identificador de diagnóstico.
- Registrar caminho/versão do Node efetivamente utilizados, duração e categoria da falha. Capturar saída de erro de forma limitada e sanitizada, sem conteúdo de auth.json, registros do banco, tokens ou ambiente completo.
- Só mencionar incompatibilidade de Node quando houver evidência dela. Uma consulta inválida ou timeout deve ser apresentada como tal, com texto de UI em inglês.

### Aceite

Uma falha de módulo SQLite, erro de schema e timeout devem ter diagnósticos distintos. Cancelamento solicitado pelo usuário não deve aparecer como Node desatualizado. O mesmo erro deve ser compreensível no catálogo e no início de uma conversa.

## 4. Conversas de automação com comportamento normal

### Remoção das travas

- Retirar a coerção de `mode = steer`, a imposição de `steeringOnly` e a ocultação de Send now motivadas apenas pelo status da automação.
- Retirar as rejeições equivalentes dos endpoints de envio/fila e a exclusão indiscriminada dessas conversas pelo agendador de mensagens.
- Usar estado real do turno para exibir envio, fila, steering e Stop, como nas conversas comuns. Um registro de automação `running` sem turno executando não pode manter o composer preso.
- Reaproveitar as regras comuns de edição do composer para o próximo turno; não conservar um bloqueio global apenas por origem em automação. Capturas já em execução continuam imutáveis, como em qualquer turno.
- Preservar idempotência das mensagens, FIFO, tratamento de entrega incerta e um único turno de harness ativo por conversa. Remover a trava de produto não significa executar dois turnos concorrentes na mesma sessão.

### Transições de controle

| Ação humana | Resultado esperado |
| --- | --- |
| Steering durante um turno | Entrega no próximo limite do modelo; mantém o turno e sua captura atual |
| Mensagem em fila | Entrada aceita pela fila comum; no próximo despacho disponível tem prioridade sobre lembrete/continuação automática ainda não iniciados |
| Send now | Interrompe o turno pelo fluxo comum, aguarda seu encerramento confirmado e inicia a mensagem humana |
| Envio quando não há turno ativo | Inicia uma conversa comum, mesmo que tenha sobrado metadado `running` de automação |
| Stop | Usa o encerramento comum; a origem em automação não exige um fluxo especial para liberar a conversa |

Quando uma mensagem humana iniciar um novo turno e houver uma automação ainda dona da conversa, encerrar essa propriedade de forma transacional como `interrupted_by_user`, preservando o histórico. Uma mensagem apenas enfileirada não deve interromper antecipadamente o turno atual. Se a automação já terminou normalmente, preservar seu resultado.

Lembretes, continuações e decisões tardias devem validar o run/turno proprietário e não reativar uma automação substituída. A mensagem humana não herda automaticamente sua autorização de graphs nem suas ferramentas de classificação. Graphs independentes já aceitos não são cancelados implicitamente por Stop/Send now do chat; suas notificações permanecem atribuídas à atividade original.

### Finalização OpenCode

- Corrigir a distinção entre turno nativo concluído, servidor ocioso retido e processos de trabalho pendentes. Não marcar uma automação como bloqueada apenas porque seu servidor continua aberto.
- No encerramento da automação, obter a confirmação de encerramento necessária pelo domínio de processos já existente. Preferir fechar o runtime nesse limite quando a distinção não puder ser demonstrada com os sinais atuais; não atribuir `processesDrained = true` sem evidência.
- Manter o requisito mais forte de encerramento dos processos de nós de graph e de filhos pertencentes a um lote.
- Registros antigos presos precisam poder ser encerrados pelo envio humano/Stop normal, sem replay automático do objetivo nem alteração silenciosa de resultados históricos.

### Aceite

Enviar após falha, durante execução e após conclusão da automação; usar steering, fila, Send now e Stop. Não pode haver mensagem eternamente em Next turn, lembrete concorrente, perda de input nem duas execuções simultâneas. Automação sem intervenção deve classificar e finalizar normalmente.

## 5. Fonte consistente de autenticação do gh

- Resolver o diretório de configuração do gh a partir do ambiente original do engine, antes de aplicar o perfil isolado do harness.
- Respeitar a precedência nativa: `GH_CONFIG_DIR` explícito; senão `XDG_CONFIG_HOME/gh`; senão o diretório padrão do sistema/usuário, incluindo AppData no Windows e `~/.config/gh` no Linux.
- Projetar o resultado como caminho absoluto em `GH_CONFIG_DIR` nos ambientes pertinentes de Pi, OpenCode e Codex. Usar o mesmo resolvedor nos processos auxiliares de catálogo que compartilhem esse ambiente.
- Não copiar credenciais para diretórios temporários, não criar um segundo login e não alterar o HOME global ou os arquivos originais do gh.
- Preservar tokens e host explicitamente configurados segundo a precedência do próprio gh. Uma fonte consistente não deve substituir a conta/host escolhido pelo usuário.
- Preservar a exclusividade de configuração/Library do OpenCode: corrigir o efeito colateral sobre gh sem reabrir a descoberta nativa que o perfil controla.
- Não ampliar por esse caminho o ambiente dos servidores MCP independentes, que têm sua própria configuração de credenciais.

### Aceite

Sessão manual e automação no mesmo engine/harness encontram a mesma configuração e conta esperada, sem pedir para trocar escopo. Verificar também XDG personalizado, GH_CONFIG_DIR explícito e reinício. Não exibir tokens nos diagnósticos.

## 6. Autorização de graphs pela automação

### Modelo de autorização

- A configuração salva pelo usuário passa a autorizar a execução dos graphs selecionados dentro do objetivo daquela automação. Não adicionar uma segunda confirmação para os mesmos graphs.
- Na admissão de cada run, capturar a origem dessa autorização: projeto, automation ID, revisão/configuração capturada, automation run ID e graphs selecionados.
- Persistir nas atividades uma autorização discriminada por origem: mensagem humana existente ou configuração de automação. Não fabricar eventos `user_message`, IDs ou mensagens humanas de sistema.
- Manter compatibilidade de leitura com atividades antigas que contêm apenas `{eventId, text}`.
- Para uma chamada do turno proprietário de automação, permitir omitir referências de mensagem: o engine verifica o graph na configuração capturada e vincula a autorização. O modelo não pode fornecer um automation ID arbitrário como prova.
- Fora desse contexto, continuar exigindo autorização humana por mensagem. A seleção comum do picker não vira autorização automática.
- Alterações da seleção valem para novos runs; não modificar a captura de um run já admitido. Remover um graph impede novas autorizações em runs posteriores.

### Persistência e compatibilidade

- Atualizar schemas das ferramentas, decoders estritos, registros, projeções SQL, validação no carregamento, scheduler, consultas de atividades e descrições das ferramentas.
- O scheduler revalida a origem persistida, e não depende de um evento humano inexistente. Uma atividade já admitida mantém sua proveniência mesmo após encerramento do chat/automação proprietário.
- Automações já salvas com seleção explícita passam a usar a nova semântica nos próximos runs. Não iniciar retroativamente graphs de runs concluídos ou falhos, nem fabricar grants em registros incompletos.
- Recuperações explícitas usam sua configuração capturada. Não transformar recuperação de downtime em replay automático.
- Atualizar `automation-run.md`, `orchestrator.md`, orientações de colaboração pertinentes e `product.md` conjuntamente. A orientação atual de que a lista nunca autoriza execução deve ser removida especificamente para automações.
- Preservar seleção de workspace: primeira execução no diretório original por padrão. A autorização de graph não escolhe automaticamente fresh/reuse para um retry; políticas existentes de retry continuam sendo verificadas. Não ampliar esta entrega para uma política nova de retries autônomos.

### Aceite

Uma automação executa um graph selecionado sem perguntar autorização novamente. Graph não selecionado, outro projeto e chamada fora do turno proprietário não recebem essa autorização. Persistência/reinício preservam proveniência e idempotência sem duplicar trabalho. Perguntas reais de ferramentas e permissões continuam seguindo a política normal; autorizar o graph não fornece respostas a elas.

## 7. Graphs no Linux

- Reutilizar a supervisão Linux existente, com subreaper/pidfd, em um domínio exclusivo por ativação. Não substituir a verificação de plataforma por um simples `windows || linux` sem integrar a prova de encerramento.
- Verificar disponibilidade dos mecanismos necessários antes de iniciar trabalho. Expor no catálogo a razão concreta de indisponibilidade; manter `runtimeValidated` fiel às evidências disponíveis.
- Preservar a sequência de Choice reservada, bloqueio de novas ferramentas, conclusão das chamadas em andamento, encerramento nativo/processos e aceitação durável.
- Pi/OpenCode mantêm seus hooks de bloqueio de ferramentas. Codex precisa confirmar encerramento do domínio de processos como barreira antes da aceitação da Choice, equivalente à função do Job no Windows.
- Integrar todos os executores: agentes, filhos pertencentes ao nó, Terminal, comandos Git de workspaces, Fork/Join e continuação de sessões nativas quando configurada.
- Confirmar encerramento de descendentes, inclusive processos que criam outra sessão/grupo. Não usar apenas PID, grupo de processo ou observação de `/proc` como prova de encerramento.
- Perda do supervisor, timeout de encerramento ou chamadas MCP com resposta incerta devem continuar produzindo falha/incerteza, sem liberar dependências como se houvesse sucesso.
- Manter processos MCP e trabalho remoto dentro de suas fronteiras reais: encerrar um processo local não comprova que trabalho remoto terminou.
- Atualizar capacidades e textos atualmente presos a Windows/Job em `GraphAdapterCapabilities` e no inventário de harnesses, distinguindo mecanismo implementado de runtime validado.

### Aceite

Em Linux, executar as topologias existentes: agente → agente, Terminal, Choice terminal, Fork/Join, workspace original e worktree isolada, além de continuação nativa. Verificar separadamente Pi, OpenCode e Codex. A próxima etapa só pode começar após as confirmações exigidas. Não deixar processos órfãos nem encerrar processos de outra conversa.

## 8. Terminal com PowerShell, Bash e sh

### Contrato

- Adicionar `shell` ao nó Terminal com valores `powershell`, `bash` e `sh`, persistidos em YAML/JSON, canvas e snapshot do run.
- Campo ausente em graph/snapshot antigo continua significando **PowerShell em qualquer plataforma**. Não reinterpretar scripts existentes como Bash ao abrir no Linux.
- Novos nós devem salvar o shell explicitamente: PowerShell no engine Windows e sh no engine Linux. A escolha se baseia no engine, não no sistema do navegador. Bash permanece selecionável.
- Usar o shell selecionado sem fallback silencioso. Se faltar o executável, falhar no preflight com nome do nó, shell e instrução de recuperação adequada.

### Execução

- Manter `KLM_GRAPH_PAYLOAD_FILE` como arquivo JSON de entrada. PowerShell conserva `$payload`; Bash/sh recebem o caminho do arquivo, sem converter valores do payload em código, `eval` ou variáveis de shell arbitrárias.
- Executar script autoral em arquivo privado por processo supervisionado, com cwd do workspace. Não interpolar tarefa/payload na linha de comando.
- Bash usa modo não interativo sem arquivos de inicialização; sh usa execução não interativa. Neutralizar mecanismos de inicialização como BASH_ENV/ENV nesses executores, sem remover o ambiente de trabalho necessário.
- Não introduzir dependência obrigatória de jq/Python para executar comandos simples; scripts que precisem interpretar JSON podem declarar suas próprias ferramentas.
- Preservar stdout/stderr, exit code, truncamento e a semântica atual de saída não zero. Não acrescentar opções como `set -e` que alterem silenciosamente o script.
- Atualizar validação, importação/exportação, snapshots e visualização da configuração para que o shell sobreviva a round trips. Reutilizar componentes do design system; labels em inglês.

### Aceite

Executar Bash e sh com espaços, Unicode e payload contendo aspas/metacaracteres sem execução involuntária. Confirmar exit code e resultado mapeado. Graph antigo PowerShell mantém comportamento; abrir/salvar no editor não troca seu shell. Graph Linux sem nós PowerShell não exige instalar PowerShell.

## 9. Credenciais MCP protegidas no Linux

### Solução proposta para ambiente headless

- Manter DPAPI no Windows. Implementar backend Linux com criptografia autenticada AES-256-GCM e chave aleatória persistente do engine, sem depender de keyring gráfico/D-Bus desbloqueado.
- Armazenar a chave fora do SQLite, em arquivo privado do diretório de dados do engine, com diretório 0700, arquivo 0600 e proprietário correto. Aplicar criação exclusiva/atômica e verificações de links/propriedade para não sobrescrever uma chave existente.
- Não transportar a chave em argumentos, variáveis de ambiente dos harnesses, snapshots, respostas de ferramentas ou logs.
- Usar envelope versionado com nonce novo por gravação e associação ao domínio de credenciais MCP. Validar integridade ao ler; nunca retornar plaintext de um blob inválido.
- Adaptar o acesso às credenciais para receber contexto do armazenamento/chave do engine. Evitar uma chave global implícita que misture instâncias dev/release ou testes.
- Não regenerar chave ausente quando já houver ciphertext dependente dela. Retornar erro recuperável que distinga chave ausente, permissão incorreta, formato incompatível e integridade inválida.
- Cobrir segredos manuais, variáveis de servidores stdio, OAuth inicial e refresh de tokens. IDs de credencial continuam sendo as únicas referências nos snapshots.
- Documentar backup/restauração do par base + chave, com proteção equivalente e sem incluir a chave em exportações comuns. Blobs DPAPI não são portáveis para Linux; exigir nova autenticação quando necessário.
- Essa proteção é contra leitura indevida de arquivos/base, não isolamento de processos já executando como o próprio usuário do engine. Não prometer uma fronteira de segurança inexistente.

### Aceite

Salvar/testar/reutilizar uma conexão manual e uma OAuth após reiniciar o engine em Linux headless. Segredos não aparecem em plaintext no SQLite, snapshots ou logs. Chave ausente ou incorreta produz erro claro sem apagar credenciais. Windows continua lendo seus blobs existentes.

## 10. Subagentes nativos do Pi no Linux

- Manter a distinção entre a integração nativa `pi-subagents` e a ferramenta de filhos pertencentes ao engine `subagent_run`.
- Rastrear o caminho efetivamente usado: hoje a integração nativa é preparada apenas quando a delegação gerenciada não está ativa. Não ativar ambas para resolver o bloqueio.
- Adaptar a factory/runner nativos ao supervisor Linux e à versão de pacote compatível, preservando execução leaf, diretório, ferramentas permitidas e vínculo de resultados com o pai.
- Só retirar a recusa Linux após integrar início, conclusão, cancelamento e encerramento dos descendentes. Falta do pacote continua não impedindo um chat Pi que não precise dele.
- Não mudar modelo/harness/permissões de filhos por efeito colateral dessa portabilidade. Preservar as regras próprias de cada mecanismo.

### Aceite

No caminho nativo pertinente, executar um filho e cancelar outro sem órfãos ou acesso às ferramentas administrativas do pai. Confirmar que o caminho gerenciado `subagent_run` continua independente e que um Pi sem o pacote instalado continua iniciando normalmente.

## 11. Seletor nativo de pastas no Linux

- Tornar a capacidade do picker dinâmica no Linux: requer contexto gráfico utilizável e backend disponível, em vez de uma constante baseada somente no sistema operacional.
- Implementar adaptadores pequenos para Zenity e KDialog, executados diretamente com argumentos separados e operação exclusiva de seleção de diretório.
- Preservar o seletor Windows e o navegador de arquivos remoto como fluxo padrão. A janela nativa pertence ao computador do engine, não ao navegador remoto.
- Em VPS headless, não anunciar picker nativo, não tentar abrir janelas e não bloquear cadastro de projeto. O navegador de pastas da aplicação continua sendo o caminho disponível.
- Tratar cancelamento sem erro de produto; distinguir falha do backend de caminho inválido. Validar diretório absoluto existente na resposta e encerrar o processo em cancelamento/timeout da requisição.
- Reutilizar o endpoint/capacidade existente e componentes atuais. Não instalar dependências gráficas automaticamente.

### Aceite

Com ambiente gráfico, selecionar e cancelar diretórios com espaços/Unicode. Sem display/backend, a opção nativa não aparece e o navegador remoto funciona. A ausência de interface gráfica não é um bloqueio para a versão headless.

## 12. Sequência de entrega

| Etapa | Entrega verificável | Dependências |
| --- | --- | --- |
| A | Diagnóstico OpenCode e diretório canônico do gh | Nenhuma |
| B | Composer/fila comuns e finalização correta das automações | Nenhuma; validar junto com A |
| C | Autorização persistida pela configuração da automação | Integrar com a propriedade de run/turno definida em B |
| D | Supervisão e admissão de graphs Linux | Base Linux existente; verificar Pi/OpenCode/Codex |
| E | Shell persistido e executores Bash/sh | D para execução real; editor/contrato podem ser preparados antes |
| F | Backend Linux de credenciais MCP | Independente; validar também em graphs após D |
| G | Integração nativa Pi e picker gráfico Linux | Supervisor existente; D reaproveita a mesma base |
| H | Documentação, compatibilidade e preparação de versão | A–G concluídas e evidências registradas |

Entregar cada etapa em mudanças focadas. Essa sequência não solicita criação de sessões, subagentes ou execução paralela por agentes.

## 13. Persistência, documentação e versão

- Atualizar `product.md` na implementação para refletir as decisões aprovadas, especialmente composer de automações e autorização por configuração.
- Alterar schemas/validações SQLite conforme os novos campos exigirem, preservando registros anteriores. Não renumerar o contrato de storage/protocolo apenas porque a release muda para 0.6.2; avaliar a negociação real necessária para clientes antigos.
- Evitar que um cliente antigo remova `shell` ou autorizações que não conhece: preservar metadados no round trip ou rejeitar de forma explícita a edição incompatível. Não aceitar downgrade silencioso de semântica.
- Atualizar `engine/README.md`, `distribution/linux/README.md`, README principal e orientações pertinentes do projeto, removendo afirmações obsoletas apenas para os caminhos efetivamente entregues.
- Na etapa de release, alinhar os marcadores de versão dos componentes distribuídos: frontend/lockfile, Tauri/Cargo/lockfile, mobile quando distribuído e scripts/instaladores. Verificar os valores existentes no momento da implementação; o checkout investigado ainda declara 0.6.0 em vários arquivos.
- Registrar evidência por frente em `V0_6_2_IMPLEMENTATION_PROGRESS.md`, criado durante a implementação: implementado, checagem executada, validação humana pendente e limitações restantes.
- Não transformar um build bem-sucedido em alegação de suporte Linux validado. Não apagar bases, credenciais ou artefatos anteriores para fazer verificações passarem.

## 14. Verificação e critérios de conclusão

Seguir a política de validação do projeto: menor checagem relevante por mudança, sem suíte completa, testes de integração complexos ou revisão adversarial automática. Os cenários abaixo orientam validação humana/focada durante a implementação; não foram executados ao escrever este plano.

| Cenário | Evidência esperada |
| --- | --- |
| Inspeção OpenCode falha | Categoria real, Node efetivo e diagnóstico sanitizado; nenhuma mensagem genérica enganosa |
| Automação ativa/ociosa/falha | Input disponível; envio e interrupção comuns; fila progride sem corrida com continuação |
| Conclusão sem intervenção | Resultado classificado e run finalizado; servidor retido não deixa automação presa |
| gh em manual/automação | Mesmo diretório/configuração efetiva e conta esperada, sem revelar token |
| Graph autorizado na automação | Execução sem confirmação por mensagem; origem persistida; duplicata idempotente |
| Graph fora da autorização | Recusa sem herdar grant de outra automação/projeto/turno |
| Graph Linux com cada harness | Choice só aceita após confirmações; dependências, Fork/Join e workspaces funcionam |
| Bash/sh/PowerShell | Shell escolhido preservado; payload é dado; exit code/cancelamento corretos |
| MCP manual/OAuth no Linux | Segredos persistem protegidos e funcionam após restart; refresh não perde credenciais |
| Pi nativo | Filho termina/cancela corretamente; não interfere no subagent_run gerenciado |
| Picker Linux | Diálogo em host gráfico; navegação remota disponível em host headless |
| Windows e dados anteriores | PowerShell, DPAPI, composer e autorização humana comum continuam funcionais |

Para concluir a versão: compilar os componentes alterados para os alvos pertinentes, registrar uma rodada focada de validação real em Linux e os cenários Windows afetados, e explicitar o que ainda não foi validado. Acesso à VPS e execução com provedores/credenciais reais devem ser combinados no momento da validação; esta tarefa de planejamento não os executa.
