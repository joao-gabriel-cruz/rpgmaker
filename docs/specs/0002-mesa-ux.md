# Spec 0002 — Mesa com o mapa em destaque (UX/UI)

- **Status:** READY
- **Data:** 2026-11-21
- **Origem:** pedido do usuário — "interface o mais fácil e dinâmica para ser jogada, deixando o mapa em destaque"

## Problema

A página da campanha empilha tudo verticalmente (sessão, rodada, efeitos, presença, partitura,
mapa, personagens, refúgio, membros, configurações) e o mapa fica no meio da rolagem, com cinco
barras de ferramentas empilhadas por cima. Para jogar é preciso rolar a página, trocar de página
para rolar dados (ficha) e caçar botões. Uma mesa virtual boa deixa o mapa sempre visível e traz as
ações até ele.

## Princípios

1. **O mapa é a tela.** A mesa ocupa a janela inteira (`100dvh`), sem rolagem de página. Tudo o
   mais flutua sobre o mapa ou vive em painéis recolhíveis.
2. **Uma ação, um gesto.** Rolar dados, mover o token, pingar e medir sem sair do mapa.
3. **Mostrar o que importa agora.** Jogador vê seu personagem e a cena; ferramentas do Mestre só
   aparecem para o Mestre e só as opções da ferramenta ativa ficam abertas.
4. **Estado sempre visível, mas discreto.** Sessão, rodada/fase e quem está online ficam na barra
   superior em formato compacto.
5. **Teclado em primeiro lugar** para quem joga muito: atalhos de uma tecla, `Esc` volta para
   "Mover", `?` mostra a lista de atalhos.
6. **Consistência visual:** tema escuro (zinc/emerald), ícones `lucide-react` (contorno, stroke
   1.5 ao lado de texto normal, preenchido/realçado só no estado ativo), superfícies com sombra
   para elevação, raio concêntrico, `active:scale-[0.96]` nos botões, transições só de
   propriedades específicas (nunca `transition: all`), sem animação em interações de alta
   frequência; respeitar `prefers-reduced-motion`.

## Layout da mesa (desktop)

```
┌───────────────────────────────────────────────────────────────────────────────┐
│ TopBar: ◀ Campanha · Cena ▾(Mestre) │ Sessão ● │ Rodada 2 · [Prep|Inf|Assim|Fim] │ 👥 avatares │ ⚙ │
├──┬───────────────────────────────────────────────────────────────────┬────────┤
│T │                                                                   │ Side-  │
│o │                     MAPA 3D (preenche tudo)                       │ bar    │
│o │                                                                   │ (abas  │
│l │   [painel de opções da ferramenta ativa, flutuando junto à rail]  │ ícone) │
│  │                                                                   │        │
│R │                    bandeja de dados (overlay central)             │        │
│a │                                                                   │        │
│i ├───────────────────────────────────────────────────────────────────┤        │
│l │ Hotbar do jogador: retrato · Saúde · Determinação/Assimilação ·   │        │
│  │ [🎲 Rolar] · atalhos de instinto · abrir ficha                    │        │
└──┴───────────────────────────────────────────────────────────────────┴────────┘
```

### TopBar (altura ~48px)
- Voltar para "Minhas campanhas", nome da campanha.
- **Mestre:** seletor de cena (dropdown com criar/renomear/ativar/apagar; a cena ativa marcada).
  Jogador vê só o nome da cena ativa.
- Sessão: indicador (ponto verde "Em sessão"/cinza) — Mestre clica para iniciar/encerrar.
- Rodada: "Rodada N" + 4 fases compactas com a atual realçada; Mestre tem ◀ ▶ e iniciar/encerrar.
- Presença: avatares/iniciais empilhados dos online (tooltip com nome e papel).
- Menu ⚙ (Mestre): Membros e convites, Configurações da campanha, Regras de efeitos — abrem em
  diálogo/drawer, fora do fluxo do mapa.

### ToolRail (esquerda, vertical, ícones 40px com tooltip + atalho)
Ferramentas para todos: **Mover/selecionar (V)**, **Ping (P)**, **Régua (R)**, **Desenhar (D)**.
Separador e, só para o Mestre: **Terreno (T)**, **Objetos (O)**, **Muro (M)**, **Névoa (F)**,
**Tokens (K)**, e o toggle **Rascunho** (com faixa de aviso quando ativo).
- Uma ferramenta ativa por vez, estado único compartilhado (store/contexto) que o mapa consulta para
  decidir o que o clique faz — substitui as checagens espalhadas em cada barra hoje.
- **Painel de opções** flutuante ao lado da rail, só da ferramenta ativa (ex.: Terreno → tipo,
  subir/descer/definir altura, preencher, apagar, tamanho; Objetos → paleta; Desenhar → cores,
  espessura, apagar meus/limpar; Névoa → modo, raio, revelar/esconder, tudo; Tokens → colocar
  personagem, focar todos). Recolhível.
- Câmera: botões pequenos no canto inferior direito do mapa (centralizar, zoom +/−, girar).

### Sidebar (direita, recolhível, ~360px, abas por ícone com badge)
Abas: **Grupo** (cartões dos personagens: retrato/miniatura, nome, Saúde em barra com o nível,
Determinação/Assimilação; clicar abre a ficha no drawer; Mestre vê também NPCs), **Rolagens**
(log + campo `/r`), **Conflito** (Partitura), **Refúgio**, **Efeitos** (Mestre, badge de pendentes).
Recolher deixa só a coluna de ícones; badges continuam visíveis (ex.: nova rolagem, efeito pendente).

### Hotbar do jogador (inferior, centralizada sobre o mapa)
- Para o jogador: o seu personagem (se tiver mais de um, alterna). Para o Mestre: personagem/NPC
  selecionado no mapa ou no Grupo.
- Mostra retrato/miniatura, Saúde (barra + nível + penalidade), Determinação e Assimilação.
- **🎲 Rolar** abre o **Compositor de rolagem** (popover acima da hotbar): escolher Instinto e
  Aptidão em grades de botões com os valores, modificadores (Empenho, Ocupação, Evento Marcante,
  Assimilada + 2º instinto; Mestre: dificuldade e secreta), prévia da parada ("2d6 + 1d10, manter 2")
  e botão Rolar. Lembra a última combinação ("Rolar de novo").
- Botão **Ficha** abre a ficha completa num **drawer lateral** sobre o mapa (a rota da ficha
  continua existindo para link direto).

### Bandeja de dados
Overlay centralizado sobre o mapa (não um card fixo no canto), fundo translúcido; quando os dados
param, escolha dos mantidos e resultado grande; fecha sozinho alguns segundos após o resultado
para quem não rolou; quem rolou fecha ao confirmar. Toasts curtos para eventos (efeito aplicado,
ping de alguém fora da tela, objetivo concluído).

### Página "Minhas campanhas" e ficha
Fora da mesa, manter o layout atual, só alinhando o visual (ícones, botões). A página antiga da
campanha deixa de existir como lista vertical: membros/convites/configurações vão para o menu ⚙.

## Fora de escopo
Mobile (desktop-first, só não pode quebrar em tablet), temas claros, novas regras de jogo.

## Entregas (tickets)
1. **Mesa em tela cheia** — layout (TopBar, ToolRail com estado de ferramenta unificado, Sidebar com
   abas, slots de hotbar/overlay), migração de todos os componentes existentes para os novos
   lugares, atalhos e `?`.
2. **Hotbar, compositor de rolagem e ficha em drawer** — depende de 1.
3. **Ferramentas do Mestre no painel de opções** — terreno/objetos/muro/névoa/tokens/rascunho
   saindo das barras empilhadas para a rail + painel; controles de câmera — depende de 1.

## Notas da entrega 3 (ferramentas do Mestre)

- **Atalhos novos** (listados no diálogo `?`): Terreno `1–6` (pintar, subir, descer, altura, preencher, apagar) e Objetos `1–3`
  (colocar, selecionar, remover) — as letras ficam com a rail e a câmera; `Shift+R` gira o objeto e `−`/`+` mudam o pincel
  (como antes). Câmera, para todos: `C` centralizar no mapa, `H` no meu token, `Z`/`X` aproximar/afastar, `Q`/`E` girar 90°,
  `U` visão de cima. `Ctrl/⌘+Z` desfaz e `Shift+Ctrl/⌘+Z` (ou `Ctrl+Y`) refaz.
- **Desfazer/refazer** é do cliente e só da sessão atual: cada traço do pincel (até soltar o mouse) e cada operação de objetos
  vira um passo; o patch inverso sai pelos mesmos comandos `map.paint`/`map.objects`. Edições de outras pessoas não entram na
  pilha (desfazer pode sobrescrever uma edição alheia na mesma célula).
- **Tokens:** "Colocar" arma o personagem e o próximo clique no mapa o posiciona (o servidor recusa células inválidas);
  o botão de varinha mantém o posicionamento automático.

## Notas do polimento (RPG-66)

- **Zonas flutuantes sem sobreposição:** o `TableLayout` organiza o mapa em fluxo — linha superior (painel de opções à esquerda,
  drawer da ficha à direita) e linha inferior em grid `auto | 1fr | auto` (dock Falar/Reagir, hotbar centralizada com largura máxima,
  câmera). A câmera sai do mapa via `CameraPortal`. Com o mapa estreito (< 68rem) a dock vira só ícones.
- **Avatar da aparência:** `CharacterAvatar` (`features/miniature`) desenha um busto 2D em SVG a partir de `sheet.appearance`
  (anel na cor do token), com fallback para `portraitUrl` e depois iniciais. Usado em Grupo, hotbar, lista de Tokens e presença.
- **TopBar compacta** (`compact:` = ≤ 1280 px): rótulos viram `sr-only` com `title`/`aria-label`; nada quebra linha.
- **Vulcão:** geometria procedural low-poly (`map3d/objects/volcanoGeometry.ts`), cratera e veios emissivos animados
  (parados com `prefers-reduced-motion`).
