# KLM — engine independente, desktop Tauri e modo foco

Data: 2026-09-14.
Status: P1–P5 implementadas em 2026-09-14; builds e instaladores gerados. Validação humana pendente.
Progresso/evidências: `ENGINE_DESKTOP_IMPLEMENTATION_PROGRESS.md`.
Memória: `plans/engine-desktop-tauri.md`.

## 1. Objetivo e precedência

Entregar uma solução pessoal simples para Windows: dois instaladores, uma engine
Go independente e um cliente Tauri que também serve a interface para navegadores.
Não criar uma plataforma de administração. Comunicação com o usuário deve ser concisa.

Este documento registra as decisões mais recentes e substitui as propostas anteriores
de engine servindo o frontend, engine empacotada como sidecar do cliente e acesso
exclusivo por loopback na API pública. As regras de execução de grafos continuam
em `GRAPH_AUTHORING_REFINEMENT.md`, inclusive workspace original por omissão e
garantia de MCP limitada ao nó/chamadas, não a tarefas remotas destacadas.

Ler `AGENTS.md`, `product.md` e `CONTEXT.md` antes de implementar. Atualizar esses
documentos durante a implementação para refletir as decisões abaixo.

## 2. Contrato fechado

| Componente | Comportamento |
| --- | --- |
| Engine | Instalada separadamente; CLI pública somente `klm start` e `klm stop`. |
| Autostart | Engine inicia por padrão no login do usuário no Windows, sem comando/tela de configuração nesta entrega. Não é serviço antes do login. |
| API | Porta fixa `7331`, escutando em `0.0.0.0`, acessível pela rede. |
| Cliente | Instalador Tauri separado; manter `clients/desktop`. Não instala nem inicia a engine automaticamente. |
| Desktop | Apenas a janela central da aplicação, sem gradiente ou margem externa de página. |
| Servidor web | HTTP na porta fixa `7332`, escutando em `0.0.0.0`, dentro do processo Tauri. |
| Fechar janela | Oculta na bandeja; Tauri e servidor web continuam ativos. |
| Bandeja | Ações mínimas `Open` e `Exit`. Open restaura; Exit encerra Tauri e servidor web, sem parar a engine. |
| Focus | Botão com ícone de foco no header: `Session log → Focus → Side agent`. Abre o navegador padrão. |
| Navegador | Mesma aplicação com gradiente e janela central arrastável; preservar visual atual. |

URLs:

- Próprio computador: `http://localhost:7332`; API `http://localhost:7331`.
- Outro dispositivo: `http://IP-DO-COMPUTADOR:7332`; API no mesmo host, porta `7331`.
- `0.0.0.0` é endereço de escuta, não URL para abrir.
- O aplicativo Tauri precisa ter sido aberto e continuar ativo, mesmo na bandeja,
  para servir o navegador. Autostart do cliente não faz parte desta entrega.

Autenticação de rede, TLS e controles avançados ficam para o futuro, por decisão
expressa do usuário. Preservar grants, sandbox e autenticação do bridge privado.

## 3. Base existente

- `engine/main.go`: flags, startup, data-dir, lock e shutdown; hoje restringe bind a loopback.
- `engine/api.go`: API HTTP/SSE e validação de host/origem/cliente; hoje restringe a loopback.
- `engine/platform_windows.go`: lock de dados e controle dos processos de trabalho.
- `engine/graph_prompts.go`: prompts relidos; pacote instalado precisa conter os recursos.
- `engine/harness.go`: descoberta dos harnesses no ambiente do usuário.
- `clients/desktop`: React/Vite, sem configuração Tauri atualmente.
- `src/engine.ts`: URL hoje fixa em loopback; `App.tsx` também usa essa URL no SSE.
- `src/App.tsx` e `src/styles.css`: header, gradiente, janela central e drag já existentes.

Preservar todas as mudanças locais e os dados. Não reutilizar implementações antigas
de outro repositório. Os binários de teste `engine.exe`, `engine-next.exe` e
`engine-mcp.exe` não são a estrutura de distribuição: gerar `klm.exe` a partir do código atual.

## 4. Engine e CLI mínima

1. Separar despacho CLI da função que executa a engine. Reutilizar o runtime atual.
2. `klm start` cria worker independente em background, aguarda health/readiness e
   devolve o terminal. Se a mesma engine já está ativa, não cria outra instância.
3. `klm stop` solicita shutdown controlado ao worker e aguarda a saída. Se já estiver
   parado, informa isso sem erro operacional. Não depender de PID persistido para matar processo.
4. Usar canal local de controle Windows, por exemplo named pipe por usuário, para
   conectar o stop ao mesmo caminho de shutdown existente. Não expor comando de stop na LAN.
5. Flags/modo de worker podem existir internamente; não adicionar comandos públicos
   `settings`, `status`, `restart`, `logs` ou opções interativas nesta entrega.
6. Manter dados em `%APPDATA%/klm/engine`; executáveis/recursos em pasta de instalação
   separada. Redirecionar logs do worker para arquivo local, sem console permanente.
7. Reutilizar lock do data-dir. Porta ocupada por outro programa deve produzir erro
   claro, sem escolher porta aleatória nem encerrar processo alheio.
8. Parar a engine interrompe o trabalho conforme o contrato atual, sem replay
   automático de runs. Fechar/encerrar Tauri não chama `klm stop`.

Autostart: o instalador registra início por usuário no login, com lançamento em
background. Usar um único mecanismo Windows simples, preferencialmente registro
HKCU Run com launcher oculto se necessário. Remover registro na desinstalação;
parar manualmente não desabilita o início no próximo login. Sem serviço LocalSystem.

## 5. API na rede

- Alterar bind padrão para `0.0.0.0:7331` e remover a restrição de loopback apenas
  da API pública. Adaptar CORS/Host para chamadas do cliente web em `7332`, do
  WebView Tauri e do Vite em desenvolvimento; manter JSON, limites e validações existentes.
- Não mudar globalmente `loopbackHost`: `linked_bridge.go` usa essa proteção para
  o MCP privado autenticado. Bridge e endpoints privados de harness continuam locais.
- Centralizar resolução do endpoint no frontend. Desktop usa `http://localhost:7331`;
  navegador usa hostname da página com porta `7331`. Aplicar igualmente a fetch e SSE.
- Preservar `VITE_ENGINE_URL` como override técnico de desenvolvimento. Não criar seletor
  de servidor, descoberta de rede ou configuração de porta na UI.
- Arquivos, comandos e diálogos nativos continuam pertencendo ao computador da engine;
  acesso remoto não transfere execução nem o seletor de diretório ao outro dispositivo.

## 6. Tauri, servidor web e apresentação

1. Adicionar Tauri 2 em `clients/desktop/src-tauri`, reutilizando o build React/Vite.
   Rust cuida apenas de janela, bandeja, servidor estático e abertura do navegador.
2. A janela desktop carrega os assets empacotados pelo protocolo de assets Tauri.
   WebView não é servidor HTTP: adicionar servidor estático pequeno no processo Rust.
3. Servidor escuta em `0.0.0.0:7332` e serve somente os assets do build instalado,
   com MIME correto e fallback SPA. Nenhum Node/Vite é necessário em produção.
4. Usar o mesmo build para desktop e navegador. O empacotamento pode incluir uma
   cópia de `dist` como recurso para o servidor; não criar outro frontend/build funcional.
5. Servidor segue o lifecycle da aplicação, não da janela. Exit solicita seu
   encerramento e fecha o processo. Uma página já carregada pode continuar na memória
   do navegador; sem Tauri não há garantia de recarga ou novos assets em `7332`.
6. Implementar instância única: abrir outro atalho restaura a instância existente,
   sem tentar abrir um segundo listener. Se `7332` estiver ocupada por outro programa,
   informar erro e manter porta fixa; não abrir Focus apontando para serviço desconhecido.
7. Detectar desktop pelo contexto nativo Tauri, não pelo hostname. Aplicar classe de
   apresentação: desktop ocupa o WebView; web preserva gradiente, dimensões e drag atuais.
   Não mover `.app-shell` dentro do desktop com o drag web; usar comportamento de janela nativa.
8. Adicionar Focus entre Session log e Side agent, reutilizando IconButton/Lucide,
   com label/tooltip em inglês. No desktop abre `http://localhost:7332` via integração
   de abertura do navegador. No navegador já estamos em Focus; ocultar o botão redundante.
9. Engine indisponível mantém erro recuperável/Retry existente. Não parar/iniciar
   engine a partir do cliente nesta primeira versão.

Não prometer transporte de rascunhos ou localStorage entre WebView e navegador.
Dados persistidos vêm da mesma engine; estado local continua por cliente. Abrir
Focus apenas abre a URL, sem copiar mensagens ou iniciar execução.

## 7. Instaladores e recursos

- Produzir dois instaladores Windows independentes, preferencialmente NSIS para ambos:
  script pequeno da engine e bundle NSIS do Tauri. Sem pipeline de distribuição complexo.
- Engine instala `klm.exe`, prompts e demais recursos necessários; configura PATH
  do usuário e autostart. Após instalação, iniciar a engine. Não instalar harnesses
  nem substituir credenciais/configurações existentes.
- Cliente instala aplicativo, assets e atalhos, usando o suporte WebView2 do bundle.
  Não registrar autostart do cliente, copiar engine ou compartilhar pasta de binários.
- Conferir descoberta de Git/harnesses e recursos sem depender do cwd do repositório.
- Validar acesso LAN com a autorização padrão do Firewall Windows quando necessária;
  não construir uma tela própria de configuração de rede.
- Desinstalar cliente encerra somente Tauri/servidor web; desinstalar engine encerra
  seu worker e remove autostart/PATH. Preservar dados do usuário.

## 8. Ordem de implementação e critérios de saída

| Etapa | Entrega verificável |
| --- | --- |
| P1 — CLI | Start desacoplado do terminal, stop gracioso, lock/readiness e dados existentes preservados. |
| P2 — Rede | API em 7331 acessível via LAN; frontend resolve host correto em HTTP e SSE; bridge privado continua local. |
| P3 — Tauri | Janela React sem moldura web externa, bandeja Open/Exit, close oculta e instância única. |
| P4 — Web/Focus | Mesmo build servido em 7332 pelo Tauri; botão abre navegador; gradiente/drag e acesso de outro dispositivo. |
| P5 — Instalação | Dois instaladores, PATH/autostart engine, recursos resolvidos e funcionamento fora do checkout. |

Antes de cada etapa, conferir alterações locais. Manter progresso implementado,
pendências e evidências em `ENGINE_DESKTOP_IMPLEMENTATION_PROGRESS.md` durante a execução.

## 9. Validação proporcional

Seguir AGENTS.md: sem suítes completas, revisões adversariais ou modelos/grafos
automáticos. Build Go nas mudanças Go; build frontend/Tauri quando necessário para
checar integração/empacotamento. Evitar verificações sobrepostas sem falha ou mudança.

Validação humana mínima:

1. Instalar engine; `klm start` duas vezes não duplica processo; fechar terminal não para.
2. Instalar/abrir cliente; engine conecta; botão Focus abre navegador com visual web.
3. Fechar janela mantém bandeja e web; Open restaura; Exit para web sem parar engine.
4. Outro dispositivo abre `http://IP:7332` e recebe estado/SSE da engine no mesmo IP.
5. `klm stop` encerra controladamente; novo start permite reconexão dos clientes.
6. Novo login inicia somente engine; dados preservados e recursos funcionam fora do repo.

## 10. Fora do escopo

Serviço Windows antes do login, painel de administração, comandos adicionais,
autostart do cliente, autenticação/TLS de rede, descoberta de servidores, portas
dinâmicas, atualizador automático, sincronização de rascunhos, suporte multiplataforma
e redesign da interface. Nenhum commit/push ou instalação nesta fase de planejamento.

## Registro da sessão de planejamento (histórico)

Somente leitura dirigida e documentação. Nenhum código, build, instalação, regra de
firewall, configuração de autostart ou processo em execução foi alterado. Próximo
passo: implementar P1–P5 quando solicitado, usando este plano como fonte de retomada.

Implementação posterior: solicitada e entregue em 2026-09-14. Consultar o documento
de progresso para artefatos, verificações e validação humana; o registro acima
descreve apenas a sessão original de planejamento.
