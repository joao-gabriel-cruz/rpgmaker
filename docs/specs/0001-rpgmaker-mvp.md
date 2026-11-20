# Spec 0001 — rpgmaker: mesa virtual para Assimilação RPG (e outros sistemas)

- **Status:** READY
- **Data:** 2026-11-20
- **Origem:** sessão de grill (rodadas P1–P34)

---

## 1. Visão geral

Mesa virtual (VTT) web, privada, para um grupo jogar RPG online em tempo real. O Mestre
conduz campanhas; jogadores mantêm fichas, rolam dados 3D e movem tokens num **mapa 3D**
que o Mestre edita **ao vivo** (terreno, objetos, efeitos). O primeiro sistema suportado é
**Assimilação RPG**; a arquitetura aceita outros sistemas (ex.: D&D) como módulos em código.

**Motivação:** custo zero de operação e personalização total (ferramentas existentes de fãs
não atendem ao mapa vivo/3D e à personalização desejada).

### Uso e licença
- Uso **privado** do grupo do autor. Sem fins comerciais, sem cadastro público.
- Assimilação RPG é propriedade de terceiros (Editora Assimilar). O app **não reproduz texto
  nem arte do livro**: ícones, símbolos dos dados e textos são próprios. Rodapé "ferramenta
  não-oficial de fã".
- Nota: "RPG Maker" é marca registrada de terceiros — ok como nome interno/privado; trocar se
  o projeto um dia for aberto.

## 2. Decisões consolidadas

| # | Tema | Decisão |
|---|------|---------|
| P1 | Público | Eu e meu grupo (privado), arquitetura multi-campanha |
| P2 | Modo | Mesa virtual online, tempo real |
| P3/P32 | Plataforma | Web, **desktop para todos**; não pode quebrar em tablet, sem otimização mobile |
| P4 | Papéis | Mestre (poderes especiais) + Jogadores; papel **por campanha** |
| P6/P11 | Sistemas | Vários sistemas; cada um é **módulo em código** (Go + TS) atrás de interface comum; **1 sistema por campanha**; primeiro: Assimilação |
| P7/P18 | Contas | E-mail + senha; cadastro **só via convite**; reset de senha manual pelo admin no MVP |
| P8 | Stack | Backend **Go**; frontend **React + TypeScript + Tailwind**; 3D com **React Three Fiber** |
| P19/P27 | Infra | VPS própria + Cloudflare; **Postgres**; detalhes da VPS levantados na hora do deploy |
| P14 | Ficha | Ficha Assimilação completa; limites de criação **configuráveis pelo Mestre** (bloquear/avisar); Assimilações em texto livre (baralho = futuro, opção do Mestre); NPCs do Mestre |
| P15/P23 | Rolador | Jogador escolhe dados mantidos; Empenho desconta Determinação; Ocupação/Evento Marcante controlados 1×/sessão; penalidade de Saúde automática; rolagem assimilada; dados de dificuldade; rolagem secreta; **dados 3D**; `/r XdY+Z` genérico; **servidor sorteia** e a animação cai no resultado |
| P16/P25 | Comunicação | Voz no Discord; **sem chat** no app por enquanto; log de rolagens |
| P5/P12/P20/P28 | Mapa | **Editor próprio 3D**, grade de **blocos low-poly** (célula = tipo + altura) + prefabs |
| P13 | Recursos de mapa | Tokens, névoa de guerra, ping, medição, desenho livre, várias cenas, paleta de terreno (muro, vale, pico, vulcão…) |
| P21/P31 | Efeitos de terreno | Terreno tem efeitos: bloqueia movimento, **bloqueia visão (visão dinâmica)**, efeitos ao entrar/começar turno **sugeridos ao Mestre, que confirma**; sem limite de movimento automático |
| P22 | Mudanças no mapa | Ao vivo **e** modo rascunho/revelar **e** estados pré-programados |
| P29 | Câmera | Livre por jogador |
| P30 | Modelos 3D | Pacotes CC0 low-poly embutidos (upload de .glb = futuro) |
| P17/P26 | Partitura/Refúgio | Fase 4, Ameaças ligadas a tokens |
| P33 | Fases | 0 Fundação → 1 Jogável sem mapa → 2 Mapa 3D → 3 Mapa vivo → 4 Partitura/Refúgio |
| P34 | Projeto | `rpgmaker`, monorepo |

### 2.1 Complementares (confirmadas na revisão em 2026-11-20)
- **A1** Tokens de personagem como **standee** (plaquinha 3D com o retrato da ficha).
- **A2** Botão do Mestre **"focar todos aqui"** + ping centraliza câmera (além da câmera livre).
- **A3** **Visão dinâmica** ("entra" na P31.2) fica na **Fase 2** junto com o mapa 3D base;
  névoa manual do Mestre continua existindo como override.
- **A4** Repositório Git com remoto no **GitHub** (privado).
- **A5** Idioma da UI: apenas PT-BR.

## 3. Arquitetura

```
[Browser: React + TS + Tailwind + R3F]
   │  HTTPS (REST)        │  WSS (tempo real)
   ▼                      ▼
[Cloudflare (DNS, proxy, Tunnel)] ──► [VPS: docker compose]
                                         ├─ api (Go): REST + WebSocket hub + motor de dados
                                         ├─ postgres
                                         └─ volume de arquivos (retratos, assets)
```

- **Frontend:** SPA estática (Vite). Hospedagem: Cloudflare Pages ou servida pela própria VPS
  (decidir no deploy).
- **Backend Go:** monólito modular.
  - `auth` (sessão em cookie httpOnly, senha com argon2id/bcrypt)
  - `campaigns` (campanhas, membros, convites, papéis)
  - `realtime` (hub WebSocket por campanha; servidor autoritativo)
  - `dice` (RNG criptográfico `crypto/rand`)
  - `systems/<sistema>` (módulos de sistema)
  - `maps` (cenas, células, objetos, névoa, visão, efeitos)
  - `assets` (upload/serviço de imagens)
- **Servidor autoritativo:** todo comando (rolar, mover token, editar mapa, alterar ficha) vai
  ao servidor, que valida permissão/regra, persiste e faz broadcast. Clientes nunca decidem
  resultado de dado nem validam bloqueio de movimento sozinhos.
- **Persistência:** Postgres; fichas e dados específicos de sistema em **JSONB** validado pelo
  módulo do sistema.

### 3.1 Módulo de sistema (interface)
Cada sistema implementa, no Go:
- `ID`, `Name`, `Version`
- `ValidateSheet(sheet, campaignSettings) -> []Issue` (issues com severidade; o Mestre define se
  bloqueia ou avisa)
- `DeriveSheet(sheet) -> derived` (ex.: caixas de Saúde)
- `BuildRoll(sheet, rollRequest) -> RollSpec` e `ResolveRoll(spec, rng) -> RollResult`
- `ApplyRollCosts(sheet, roll) -> sheet` (ex.: Empenho consome Determinação)
- `DefaultNPCSheet()`

E no TS: componente de ficha, painel de rolagem, renderização do resultado e definição das
faces dos dados 3D.

### 3.2 Modelo de dados (rascunho)
- `users(id, email, password_hash, display_name, is_admin, created_at)`
- `sessions(id, user_id, expires_at)`
- `campaigns(id, name, system_id, settings jsonb, owner_id, created_at)`
- `campaign_members(campaign_id, user_id, role[gm|player])`
- `invites(token, campaign_id, role, expires_at, used_by)`
- `characters(id, campaign_id, owner_id null, kind[pc|npc], name, portrait_asset_id, sheet jsonb, updated_at)`
- `play_sessions(id, campaign_id, started_at, ended_at)` — controla usos 1×/sessão
- `rolls(id, campaign_id, play_session_id, user_id, character_id, visibility[public|gm|whisper], spec jsonb, result jsonb, created_at)`
- `scenes(id, campaign_id, name, size_x, size_y, is_active, settings jsonb)`
- `scene_cells(scene_id, x, y, terrain_type, height)` (ou chunk jsonb por região — decidir na implementação conforme desempenho)
- `scene_objects(id, scene_id, prefab_id, x, y, z, rotation, props jsonb, layer[live|draft])`
- `tokens(id, scene_id, character_id, x, y, z, props jsonb)`
- `fog(scene_id, revealed jsonb)`; `drawings(id, scene_id, user_id, data jsonb)`
- `scene_snapshots(id, scene_id, name, data jsonb)` — estados pré-programados
- `terrain_types(id, campaign_id null, name, look, rules jsonb)` — padrões globais + overrides por campanha
- `assets(id, campaign_id, kind, path, mime, size)`

### 3.3 Protocolo tempo real (rascunho)
Mensagens JSON `{type, campaignId, payload, reqId}` sobre WSS. Cliente → servidor: comandos
(`roll.request`, `token.move`, `map.paint`, `map.object.add/remove`, `map.draft.commit`,
`fog.reveal`, `ping`, `camera.focusAll`, `sheet.patch`, …). Servidor → clientes: eventos
(`roll.result`, `token.moved`, `map.changed`, `effect.suggested`, …) filtrados por
permissão (ex.: rolagem secreta só para o Mestre; rascunho do mapa só para o Mestre; visão
dinâmica calculada por jogador).

## 4. Módulo Assimilação RPG

> Fonte das regras: ferramentas de fãs e wiki (o livro não foi consultado). **Validar a tabela
> de faces e os números contra o livro antes de fechar a Fase 1.**

### 4.1 Ficha
- **Identidade:** nome, retrato, Geração (Pré-Colapso / Durante o Colapso / Pós-Colapso),
  Ocupação, Evento Marcante, 2 Propósitos Pessoais, 2 Propósitos Coletivos.
- **Instintos (6):** Influência, Percepção, Potência, Reação, Resolução, Sagacidade.
  Criação: começam em 1, +3 pontos, máx. 3.
- **Aptidões (12):** Conhecimentos — Biologia, Erudição, Engenharia, Geografia, Medicina,
  Segurança; Práticas — Armas, Atletismo, Expressão, Furtividade, Manufaturas, Sobrevivência.
  Criação: 7 pontos, máx. 2.
- **Cabo de Guerra:** Determinação + Assimilação = 10 (início 9/1); nível e pontos atuais;
  Determinação 0 → Suscetível.
- **Saúde (derivada):** 6 níveis × (1 + Potência + Resolução) caixas. Penalidades: Laceração e
  Ferimentos −1 Sucesso; Debilitação −2 Sucessos e custo de Determinação; Incapacitação.
- **Inventário:** 3 espaços corpo, 6 mochila (+ Refúgio na Fase 4); item: nome, Qualidade,
  Escassez, Pressão, características.
- **Características:** XP (7 iniciais), custo 2× nível.
- **Assimilações:** lista livre (nome, naipe ♥♦♠♣, descrição).
- **Configuração da campanha:** limites de criação = `bloquear` | `avisar`.

### 4.2 Dados
Faces (d6 usa 1–6; d10 usa 1–10; d12 usa 1–12):

| Face | Símbolos |
|------|----------|
| 1–2 | — |
| 3–4 | Pressão |
| 5 | Adaptação + Pressão |
| 6 | Sucesso |
| 7 | Sucesso ×2 |
| 8 | Sucesso + Adaptação |
| 9 | Sucesso + Adaptação + Pressão |
| 10 | Sucesso ×2 + Pressão |
| 11 | Sucesso + Adaptação ×2 + Pressão |
| 12 | Pressão ×2 |

### 4.3 Rolagem
1. Jogador clica em Instinto + Aptidão → parada = Instinto × d6 + Aptidão × d10.
2. Modificadores opcionais: **Empenho** (+1 mantido, −1 Determinação automático),
   **Ocupação** e **Evento Marcante** (+1 mantido cada, 1×/sessão, bloqueados após uso),
   bônus manual do Mestre.
3. **Rolagem assimilada:** dois Instintos, d6 viram d12, mínimo 2 mantidos, Pressões mantidas
   não descartáveis.
4. **Dificuldade:** Mestre adiciona 0–3 d12 (ou sugeridos por efeito de terreno, após
   confirmação).
5. Servidor sorteia → todos veem a animação 3D cair nas faces sorteadas.
6. Jogador **escolhe os dados mantidos** (app destaca sugestão); quantidade = 1 + bônus.
7. Penalidade de Saúde aplicada automaticamente ao total de Sucessos, exibida no resultado.
8. Resultado final no log: contagem de Sucesso / Adaptação / Pressão + detalhes.
- **Rolagem secreta** do Mestre (só ele vê). **`/r XdY+Z`** genérico disponível para todos.
- Mestre "inicia/encerra sessão de jogo" → reseta usos 1×/sessão.

## 5. Mapa 3D

- **Grade de blocos:** cada célula tem `terrain_type` + `height`. Pincéis: tipo de terreno,
  subir/descer altura, preencher área, apagar.
- **Paleta de terreno/objetos:** muro (arrastar ao longo da grade), vale, pico, vulcão (prefab
  com lava animada), água, floresta, ruínas, carro, barricada… a partir de pacotes CC0.
- **Camadas:** terreno → objetos → tokens → desenhos → névoa.
- **Cenas:** várias por campanha; Mestre define a cena ativa.
- **Tokens:** standee com retrato (A1); jogador move o seu, Mestre move todos; snap na grade.
- **Câmera:** livre por pessoa; "focar todos" do Mestre (A2); ping centraliza.
- **Névoa:** manual (Mestre revela/esconde) + **visão dinâmica** por token (linha de visão com
  altura; terrenos/objetos com `blocks_vision`). Calculada no servidor por jogador.
- **Medição** de distância e **desenho livre** (Fase 3).
- **Efeitos (regras por tipo de terreno/objeto, editáveis pelo Mestre):**
  - `blocks_movement` (ou bloqueio por diferença de altura > N)
  - `blocks_vision`
  - `on_enter` / `on_turn_start`: dano na Saúde, +N d12 de dificuldade na próxima rolagem,
    aviso. → gera `effect.suggested` para o Mestre, que **confirma ou descarta**.
  - Sem limite de movimento automático.
- **Mudanças ao vivo:** edição aparece imediatamente para todos.
- **Modo rascunho:** Mestre edita numa camada invisível aos jogadores e aplica tudo de uma vez.
- **Estados pré-programados:** snapshots nomeados da cena ("rodada 3: erupção") disparados
  pelo Mestre.

## 6. Fases e critérios de aceite

### Fase 0 — Fundação
- Monorepo (`/backend`, `/frontend`, `docker-compose.yml`), lint/test/build nos dois lados.
- Admin cria primeira conta; Mestre cria campanha escolhendo sistema; gera link de convite;
  convidado se cadastra (e-mail + senha) e entra como jogador.
- Login/logout com sessão persistente; admin reseta senha.
- WebSocket conectado por campanha com presença (quem está online).
- Interface de módulo de sistema + registro do módulo Assimilação (vazio).
- Deploy inicial na VPS atrás da Cloudflare (levantar specs da VPS nesse momento).

### Fase 1 — Jogável sem mapa → primeira sessão real
- Ficha Assimilação completa com Saúde derivada e validação configurável.
- NPCs do Mestre (visíveis só para ele).
- Rolador completo (§4.3) com dados 3D de faces customizadas, sorteio no servidor.
- Log de rolagens persistido, filtrado por visibilidade; `/r` genérico.
- Iniciar/encerrar sessão de jogo.

### Fase 2 — Mapa 3D
- Cenas; editor de blocos com paleta e prefabs CC0; tokens standee; câmera livre; ping;
  focar todos; névoa manual; **visão dinâmica**; bloqueio de movimento; edição ao vivo.

### Fase 3 — Mapa vivo
- Modo rascunho/revelar; estados pré-programados; efeitos `on_enter`/`on_turn_start` com
  confirmação do Mestre; medição; desenho livre; controle de rodada (preparação → infectados →
  assimilador → fim).

### Fase 4 — Partitura de Conflito + Refúgio
- Condicionantes, Objetivos, Ameaças (com parada própria e ligadas a tokens), Ativações;
  Refúgio (Defesa, Moral, Mobilidade, Beligerância, recursos).

### Backlog futuro
Upload de modelos .glb/texturas; baralho de Assimilações; reset de senha por e-mail
(Resend/Brevo free); integração com bot do Discord; outro sistema (D&D); perspectiva/câmera
avançada.

## 7. Riscos
- **Escopo 3D:** editor + visão dinâmica + dados 3D são a parte mais cara. Mitigação: grade de
  blocos (não heightmap), assets CC0, Fase 1 jogável antes do mapa.
- **Visão dinâmica no servidor:** custo de raycast por jogador em grades grandes. Mitigação:
  limitar tamanho de cena no MVP, recalcular só ao mover token/alterar terreno.
- **Dados 3D determinísticos:** animação precisa "cair" na face sorteada. Mitigação: simular
  física e rotacionar a geometria para a face alvo, ou animação pré-calculada.
- **Regras não verificadas no livro:** tabela de faces e números de criação vêm de fontes de
  fãs.
