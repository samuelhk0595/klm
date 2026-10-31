# Plano: engines salvas com uma conexão ativa

Status: plano concluído; execução autorizada em worktree separada, por checkpoints.
Data: 2026-10-30.

## 1. Objetivo e decisões

Permitir salvar um número de engines sem limite de produto e alternar qual delas
o cliente utiliza. Apenas uma engine fica ativa por vez. Selecionar outra engine
recarrega a interface inteira e passa a mostrar seus projetos, sessões, topics,
KLM Agent, feed, Happening now, Analytics, Tasks e configurações.

Decisões confirmadas pelo usuário:

- Primeira entrega no desktop nativo Windows/Tauri.
- Um único seletor, com aparência de avatar, entre Theme e Settings no rodapé do
  Project Rail.
- Cada engine recebe um nome durante o pareamento.
- Nome e ícone são preferências locais deste cliente.
- Ícone opcional, escolhido em um catálogo predefinido; sem ícone, usar as iniciais.
- A troca efetua uma recarga completa e recebe atualizações apenas da nova engine.
- Preservar todos os rascunhos de mensagens e formulários por engine, restaurando-os
  quando o usuário retornar.

Esta proposta substitui o desenho anterior de conexões simultâneas. Não haverá
rail/feed agregado, vínculo de projetos entre engines, badges de origem nas
sessões, abas de engine em Settings ou seletores independentes em Analytics,
Tasks, KLM Agent, criação de sessões ou topics.

Navegador/Focus e Flutter/WebView ficam para uma entrega posterior. Continuam com
seu comportamento atual de uma conexão; a seleção do desktop não sincroniza com
eles. O fato de o desktop servir o frontend do Focus não compartilha seu storage
nem suas credenciais nativas com o navegador.

## 2. Base existente e abordagem

A implementação atual já corresponde a uma engine por carregamento:

| Área | Situação atual | Evolução planejada |
| --- | --- | --- |
| `src/platform.ts` | URL única em `klm.engine-url.v1`; troca com reload | Resolver a URL do registro ativo no desktop |
| `src/engine.ts` | `ENGINE_URL` definido no carregamento do módulo | Continuar fixo durante toda a vida da página |
| `src/clientAuth.ts` | Credencial, geração e bloqueio globais da engine atual | Manter isolamento por carregamento e separar o pareamento de um novo destino |
| `src/engineIdentity.ts` | Identidade do banco por endpoint e namespace de preferências | Reutilizar proteção contra troca do banco no mesmo endereço |
| `src/features/settings/ClientConnection.tsx` | Gate de autenticação para toda a aplicação | Permitir selecionar outra engine ou adicionar uma mesmo sem conexão |
| `src/features/chat/scopedConnection.ts` | Um stream compartilhado pelos escopos montados | Encerrar na troca; abrir apenas os escopos da nova página |
| `src/App.tsx` | Estado de uma engine; rascunhos de chat em memória | Preservar estado editável antes da recarga |
| `src-tauri/src/client_auth.rs` | Credenciais no Windows Credential Manager, por URL e app | Reutilizar sem copiar tokens entre endpoints |

Os caminhos `src/` e `src-tauri/` desta tabela são relativos a `clients/desktop/`.

Manter o modelo de uma engine ativa evita transformar cada request, componente e
cache em uma implementação multi-engine. Não é necessário duplicar o App nem
manter árvores React ocultas para engines inativas. A recarga elimina naturalmente
os estados e caches em memória da engine anterior.

Nenhuma alteração nas APIs de projetos, sessões, topics ou execução da engine é
prevista. O trabalho concentra-se no frontend e na persistência local do desktop.

## 3. Registro local de engines

Criar um registro versionado, por instalação/perfil do cliente, contendo:

- `id`: identificador local estável, independente do nome.
- `name`: nome escolhido pelo usuário.
- `iconId`: identificador opcional de um ícone permitido.
- `url`: endpoint canônico, validado pelo normalizador existente.
- `activeEngineId`: referência à única entrada selecionada.
- Versão do formato para migração.

O registro contém metadados, nunca códigos de pareamento ou credenciais. Pode
residir em uma única chave de localStorage para atualizar entradas e seleção em
uma gravação. Se a persistência falhar, não apresentar a troca como concluída.

Regras propostas para o cadastro:

- Nome obrigatório, após trim, com limite de tamanho coerente com os campos atuais.
- Ícone do catálogo Lucide já utilizado pelo projeto, como Monitor, Server, Cloud,
  House e Database; armazenar o ID, não SVG arbitrário.
- Iniciais: primeira letra das duas primeiras palavras; nome de uma palavra usa
  até duas letras, respeitando caracteres Unicode.
- Nomes iguais não fundem engines. Exibir a URL como informação secundária no menu
  para desambiguar.
- Rejeitar cadastro duplicado da mesma URL canônica e oferecer a entrada existente.
- Não inferir equivalência entre endpoints por nome ou diretórios dos projetos.
- Manter HTTPS para engines remotas e HTTP autenticado apenas em loopback,
  preservando portas explícitas e prefixos de caminho.
- Reabrir o aplicativo com a última engine selecionada.

O limite ilimitado significa ausência de cota artificial de cadastros. O menu terá
rolagem e busca para muitas entradas, sem fazer probes ou abrir conexões de fundo
com cada engine salva.

### Migração

1. Se não existe registro novo, importar a URL atualmente resolvida como primeira
   entrada, sem exigir novo pareamento.
2. Gerar um nome inicial local baseado no hostname, editável pelo usuário.
3. Preservar as entradas existentes do Credential Manager e os namespaces de
   navegação/preferências associados ao endpoint e à identidade do banco.
4. Só considerar a migração concluída depois da gravação válida do registro.
5. Manter o caminho legado de resolução para navegador e mobile.
6. No desktop em desenvolvimento, tratar `VITE_ENGINE_URL` como endereço inicial
   quando ainda não existe seleção salva, em vez de impedir toda troca. Documentar
   essa mudança de semântica e atualizar a cobertura existente de resolução.
7. Preservar a separação dos perfis dev/release e seus identificadores de aplicativo.

## 4. Seletor e gerenciamento

### Rodapé do rail

Ordem visual: **Theme → Engine avatar → Settings**.

O avatar permanece disponível com uma única engine. Tooltip e nome acessível
identificam a engine ativa, por exemplo `Switch engine: VPS`.

Ao clicar, abrir um menu/popover com:

- Lista de engines salvas: ícone/iniciais, nome e URL.
- Marcação da engine atual.
- Ação `Add engine`.
- Ações locais `Edit` e `Remove` para gerenciar os cadastros.

Usar os componentes existentes de menu, botões, inputs e diálogo. A seleção deve
funcionar por teclado e o menu deve manter foco visível e devolver foco ao avatar
ao fechar. Não usar indicadores que aleguem que engines inativas estão online:
elas não estarão sendo monitoradas.

Selecionar a entrada já ativa apenas fecha o menu. Selecionar outra executa o
fluxo de troca descrito abaixo. Nome/ícone podem ser editados sem reconectar.

`Remove` deve confirmar a remoção do cadastro local. Não apaga projetos ou histórico
da engine e não equivale a `Stop`. A limpeza de credenciais locais deve ser
específica ao endpoint; revogação remota permanece em Connected devices.
Ao remover a engine ativa, abrir o seletor de conexão, sem escolher silenciosamente
outra. Com zero entradas, apresentar o fluxo inicial de adicionar engine.

### Pareamento

O formulário inclui `Engine name`, ícone opcional, `Engine URL`, `Client name` e
`Pairing code`. Engine name é o apelido local; Client name continua sendo o nome
do dispositivo mostrado pela engine em Connected devices.

Fluxo proposto:

1. Validar nome e URL, inclusive duplicidade, antes de consumir o código.
2. Preservar os rascunhos atuais antes de iniciar a transição.
3. Parear o destino explicitamente informado usando o contrato atual.
4. Gravar a credencial somente na entrada nativa desse endpoint.
5. Confirmar acesso autenticado ao destino, registrar a entrada e ativá-la.
6. Recarregar o cliente no destino recém-adicionado.

Falha ou cancelamento mantém a engine anterior selecionada. Parear uma candidata
não pode invalidar a geração de autenticação, apagar a credencial ou bloquear o
cliente ativo: hoje `pairClient` altera esses globais, portanto essa parte precisa
ser separada. O handshake de pareamento é temporário, sem iniciar um segundo
inventário ou stream de atualizações.

Se o pareamento tiver sido aceito e uma gravação local posterior falhar, permitir
recuperar usando a credencial já salva, sem consumir novamente o mesmo código ou
criar outro dispositivo automaticamente.

O campo atual `Settings → Chat → Engine URL → Connect` deve usar o mesmo fluxo de
cadastro/seleção, evitando uma segunda fonte de URL ativa. Settings continua
configurando somente a engine atual.

## 5. Troca de engine

Implementar uma única operação de troca, utilizada pelo menu e pela tela de
recuperação de conexão:

1. Entrar em estado `Switching`, desabilitando novas ações e cliques repetidos.
2. Capturar e gravar os rascunhos, formulários, navegação e comprovantes de operações
   em andamento no namespace da engine atual.
3. Confirmar a persistência; se falhar, manter a engine atual e permitir tentar
   novamente. Não recarregar perdendo dados que deveriam ser preservados.
4. Fechar streams, cancelar leituras e impedir novas tentativas automáticas da
   página antiga. Tratar efeitos de escrita já enviados conforme sua confirmação
   real, sem assumir que um abort desfez a ação no servidor.
5. Persistir `activeEngineId` e chamar `window.location.reload()`.
6. Na inicialização, resolver a URL selecionada antes de construir autenticação,
   `ENGINE_URL` e namespaces; autenticar e carregar seu bootstrap.
7. Validar a identidade do banco, restaurar a navegação/rascunhos correspondentes e
   abrir o stream normal da nova engine.

Todo carregamento opera com um endpoint imutável. Não alterar `ENGINE_URL` em
memória enquanto requests antigos ainda utilizam credenciais e estado anteriores.

Trocar de engine não envia Stop, não interrompe agentes, não pausa automações e
não cancela graphs remotos. Ao voltar, buscar o estado atual da engine. Engines
inativas não recebem heartbeats do cliente; a expiração normal de pareamento por
inatividade continua válida e pode exigir parear novamente.

### Destino offline ou sem autorização

Mostrar a conexão da engine escolhida com seu nome/ícone e as ações `Retry`,
`Pair again` quando aplicável, `Switch engine` e `Add engine`.

O seletor precisa existir fora do gate que hoje impede a montagem de todo o App.
Uma engine offline não pode prender o usuário sem acesso às outras entradas.
Não retornar automaticamente à engine anterior, nem exibir seu conteúdo sob o
nome da engine escolhida.

## 6. Preservar todos os rascunhos e formulários

Esta é uma parte obrigatória da entrega, escolhida explicitamente pelo usuário.
Não basta salvar o texto do chat: vários editores possuem estado React próprio.

### Persistência

Criar um armazenamento local versionado de estado editável e adaptadores pequenos
por formulário, sem serializar indiscriminadamente a árvore React ou o DOM.

- Namespace: cadastro/endpoint + identidade do banco + tipo de superfície + ID da
  entidade, quando houver.
- Preservar revisão-base/versão de edição quando a API utiliza controle de conflito.
- Salvar durante a edição e efetuar um flush confirmado antes da troca.
- Guardar os dados necessários para reabrir o mesmo editor, aba e entidade.
- Hidratar o estado antes que valores padrão ou respostas tardias sobrescrevam o
  rascunho. Mudanças no servidor continuam sujeitas à reconciliação normal.
- Limpar o rascunho após sucesso confirmado de Save/Send ou descarte explícito.
- Restaurar um formulário não salva, envia, instala, autoriza ou executa nada.

Como formulários incluem API keys, PATs e outros segredos ainda não enviados,
persistir esses rascunhos em armazenamento nativo protegido pelo usuário do Windows.
Uma implementação direta é um arquivo de drafts por namespace, com proteção DPAPI
e gravação atômica no diretório de dados do desktop. Comandos Tauri devem aceitar
chaves controladas e valores, não caminhos arbitrários do renderer. O armazenamento
de drafts é separado do Credential Manager já usado para autenticação: a função
atual de credenciais valida tokens de tamanho fixo e não é um cofre genérico.

Metadados públicos de engines continuam no registro simples. Não colocar segredos
de formulários em localStorage, URLs, logs ou no diretório estático servido ao Focus.
Erros de armazenamento devem impedir a troca destrutiva, com recuperação explícita.

### Superfícies a cobrir

| Superfície | Estado editável a preservar |
| --- | --- |
| Chat normal e KLM Agent | Texto, menções de arquivos/pastas e destino do composer |
| Side chat | Texto e referências/trechos selecionados do chat principal |
| Criação de sessão/topic/projeto | Campos, seleções e destino do cadastro |
| Edição de projeto/topic | Valores locais e identidade da entidade |
| Agents e Graphs | Draft de configuração, topologia e demais alterações ainda não confirmadas |
| Project settings | Webhooks, Tasks, fases/mapeamentos e automações, incluindo revisões-base |
| Settings | Transcription, Engine, Library e outros formulários editáveis |
| Perguntas pendentes | Resposta em edição, vinculada à identidade da pergunta ainda pendente |
| Pareamento/edição de engines | Nome, ícone, URL e demais campos, com validade real do código |

Antes da implementação, enumerar os formulários existentes nessas áreas para que
a promessa de preservar tudo não dependa apenas do estado central em `App.tsx`.
Diálogos fechados para acessar o seletor também precisam conservar a edição, salvo
descarte explícito. Se uma interação nativa estiver em andamento, concluir ou
cancelar sua etapa antes da recarga; gravação de áudio/dispositivos não são
objetos serializáveis. Texto transcrito e dados já produzidos permanecem drafts.

Se entidade ou pergunta deixar de existir, estiver arquivada ou mudar de revisão,
manter o draft recuperável sem reaplicar a ação. A troca de banco no mesmo endpoint
não deve hidratar IDs ou drafts do banco anterior.

### Operações em andamento

Reutilizar o princípio de `PendingSubmission.tsx`: entrega sem confirmação torna-se
incerta e não é repetida automaticamente. Preservar os IDs/recibos de operações
existentes, inclusive saves com reconciliação, para consultar o resultado ao voltar
à mesma engine. Nunca reenviar essa operação para a nova engine.

Operações que não podem atravessar a recarga com segurança devem manter a troca
pendente até terminar ou oferecer cancelamento explícito. Não confundir formulário
restaurado com resultado confirmado de uma mutação.

## 7. Organização da implementação

Arquivos novos sugeridos em `clients/desktop/src/features/engines/`:

- `engineRegistry.ts`: formato, validação e migração dos cadastros.
- `EngineSwitcher.tsx`: avatar e seleção.
- `EngineConnectionDialog.tsx`: cadastro, ícone e pareamento.
- `switchEngine.ts`: sequência única de preservação e recarga.

Persistência de drafts em um módulo específico, com adaptadores por feature e
comandos nativos em um arquivo próprio de `src-tauri/src/`. Nomes finais podem seguir
as convenções encontradas durante a implementação.

Arquivos existentes prioritários:

- `platform.ts`, `main.tsx`, `clientAuth.ts` e `engineIdentity.ts`.
- `features/projects/ProjectRail.tsx`.
- `features/settings/ClientConnection.tsx` e `SettingsDialog.tsx`.
- `App.tsx`, `features/chat/PendingSubmission.tsx` e `scopedConnection.ts`.
- Editores de agents, graphs, automations, webhooks, Tasks e Library para drafts.
- `src-tauri/src/main.rs` para registrar os comandos nativos de persistência.

Reutilizar o design system e a linguagem visual existentes. Todo texto controlado
pela aplicação permanece em inglês.

## 8. Sequência de entrega

### Etapa 1 — Registro e compatibilidade

Adicionar registro versionado, importar a conexão atual, preservar credenciais e
fazer o desktop resolver um único endpoint ativo. Continuar abrindo o cliente atual
sem exigir pareamento novamente.

### Etapa 2 — Preservação de estado

Implementar armazenamento nativo de drafts e integrar os formulários inventariados.
Tratar identidade do banco, revisões, recuperação e operações de entrega incerta.
Esta etapa precisa estar funcional antes de expor a troca que recarrega a página.

### Etapa 3 — Pareamento e seletor

Adicionar avatar no rail, lista, nomes, ícones, edição/remoção local e novo pareamento.
Separar autenticação candidata da conexão ativa e consolidar o antigo Connect.

### Etapa 4 — Troca e recuperação

Integrar flush, encerramento da conexão antiga, escolha persistida e reload.
Disponibilizar seleção mesmo quando a engine ativa estiver indisponível. Restaurar
os editores somente após validar o contexto da engine de destino.

### Etapa 5 — Validação humana e documentação

Atualizar `product.md` e documentação de conexão quando o comportamento estiver
implementado, registrando separadamente evidência de compilação e validação manual.

## 9. Critérios de aceitação

1. Parear Local e VPS com nomes distintos e ícones opcionais; ambas sobrevivem ao
   fechamento/reabertura do desktop.
2. Trocar Local → VPS → Local atualiza toda a interface e restaura o contexto de
   cada engine, sem misturar IDs, projetos, mensagens ou configurações.
3. Apenas a engine ativa mantém stream/renovação de autenticação. Adicionar muitas
   entradas não multiplica conexões em segundo plano.
4. O avatar fica entre Theme e Settings; nome acessível, iniciais, menu, teclado,
   busca e rolagem funcionam com nomes longos e muitas engines.
5. Todos os formulários inventariados e drafts de main/side/KLM Agent retornam com
   seus valores, incluindo campos sensíveis protegidos, sem submissão automática.
6. Uma falha ao salvar drafts impede a recarga; uma revisão divergente não sobrescreve
   alterações remotas silenciosamente.
7. Uma mensagem/save com resposta perdida continua associado à engine original e
   não é repetido pela troca.
8. Pareamento com erro não invalida a engine ativa; credenciais nunca são enviadas
   a outra URL por causa da seleção.
9. Engine offline/revogada permite voltar a outra pelo seletor de recuperação.
10. Agentes, graphs e automações continuam executando na engine anterior; a UI busca
    seu estado atualizado ao retornar.
11. Migração conserva o acesso e as preferências existentes; banco substituído no
    mesmo endereço não recebe os drafts de entidades antigas.
12. Navegador/Focus e mobile mantêm o fluxo atual; não herdam cadastros ou tokens do
    desktop implicitamente.

Verificação leve durante a implementação: typecheck TypeScript e, quando os comandos
nativos forem adicionados, `cargo check` no projeto Tauri. Usar verificações pequenas
direcionadas para resolução de URL/migração quando necessário; não executar suites
completas, testes de integração complexos ou reviews sem solicitação explícita.

Este documento foi preparado a partir da leitura da arquitetura e das decisões do
usuário. Nenhum comportamento novo foi implementado ou validado em runtime na
etapa de planejamento.

## 10. Coordenação, worktree, PR e testes autorizados

O usuário autorizou a implementação por um executor em uma nova sessão OpenCode,
usando GPT-6.1 Sol Fast com esforço High. O coordenador é o único interlocutor do
usuário e faz revisão simples de código nos checkpoints, sem executar a aplicação
ou testes nesses checkpoints. Dúvidas e impedimentos do executor são enviados ao
coordenador. Não ampliar o escopo nem criar abstrações especulativas.

### Checkpoints obrigatórios

- **CP0 — Base e isolamento:** inspecionar as PRs abertas e suas relações, preparar
  branch/worktree própria sobre a base correta da pilha e informar SHAs, caminhos,
  dependências e inventário de formulários. Parar antes de implementar. Não fazer
  merge das PRs existentes. Se elas não formarem uma pilha inequívoca, levar a
  decisão ao coordenador antes de combinar qualquer branch.
- **CP1 — Registro e resolução:** migração, URL ativa e isolamento de autenticação;
  apresentar diff e resultados apenas das verificações leves pertinentes.
- **CP2 — Preservação:** armazenamento protegido e todos os adaptadores de drafts,
  revisões e recuperação. Apresentar a matriz de formulários cobertos.
- **CP3 — Fluxo completo:** seletor, pareamento, troca, offline e documentação.
  Apresentar diff, instruções para ambiente dev e criar/atualizar a PR empilhada na
  base aprovada, sem merge. Não incluir mudanças alheias da árvore de trabalho original.
- **CP4 — Playwright independente:** após a aprovação de código do CP3, o coordenador
  cria outro executor com o mesmo harness/modelo/esforço para testar os fluxos reais,
  reportar defeitos reproduzíveis e evidências. Correções retornam ao implementador
  antes de repetir os cenários afetados.

Cada checkpoint exige um relatório ao coordenador e uma parada efetiva. O executor
só continua com uma instrução de liberação; pedidos de correção não liberam o
checkpoint seguinte. Relatórios incluem arquivos/commit, diff desde o checkpoint
anterior, evidência disponível, limitações e a próxima etapa proposta.

### Ambiente de teste

- Usar uma engine **local de desenvolvimento**, com build `-tags dev`, API 17331,
  dados/control separados e cliente dev com identidade própria.
- Não conectar à engine instalada/principal nem usar seus dados, credenciais,
  processos ou listeners como alvo de teste.
- A segunda engine de teste pode ser a VPS existente, explicitamente autorizada.
  O coordenador obtém os detalhes da sessão de instalação de harnesses mencionada
  pelo usuário; não colocar tokens/códigos de pareamento em relatórios ou commits.
- Exercitar Local dev → VPS → Local dev, autenticando de verdade, verificando
  recarga, destino dos requests, streams, dados e restauração dos drafts.
- Priorizar Playwright ligado ao WebView2 do Tauri dev via CDP, se disponível, para
  exercitar também credenciais e armazenamento nativos. Um navegador com bridges
  simuladas não comprova o fluxo desktop real; qualquer limite dessa automação
  deve ser reportado ao coordenador antes de declarar os cenários aprovados.
- Testar erros objetivos: URL inválida/duplicada, pareamento inválido, destino
  indisponível, troca repetida, retomada de formulário, recarga e reabertura.
- Trabalhar em dados/projetos de teste próprios. Não reiniciar, atualizar ou parar
  a VPS compartilhada para simular falhas; usar interrupção de rede do cliente ou
  a engine dev isolada.

O pedido explícito de Playwright autoriza esta validação dirigida dos fluxos. Não
implica executar toda a suite do repositório nem fazer revisão adversarial.
