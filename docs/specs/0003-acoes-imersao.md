# Spec 0003 — Ações, itens e imersão na mesa

- **Status:** READY (pré-aprovada pelo usuário em 2026-11-21; ele revisa depois)
- **Depende de:** 0002 (mesa em tela cheia, hotbar e compositor de rolagem)

## Pedido do usuário (resumo)
1. Itens personalizados criados pelo Mestre.
2. Animações de ataque.
3. Animação de rolagem de dados e **tirar a caixa de texto da rolagem**: no lugar, **padrões a partir
   do modelo** (sistema/ficha) e o jogador escolhe qual usar.
4. Interações com animação: percepção, achar um item, descansar em um lugar, dançar em um festival da
   comunidade, ataque etc.
5. Analisar e implementar outras mecânicas que aumentem a imersão.

## Conceito central: Ações
Uma **Ação** é o "verbo" que o personagem faz na mesa. Ela junta, num gesto só: alvo no mapa →
rolagem pré-montada → animação para todos → consequência sugerida ao Mestre.

- **Catálogo de ações por sistema** (módulo Assimilação no backend define os padrões; o Mestre pode
  criar ações próprias da campanha). Cada ação: `id, nome, ícone (lucide), categoria (combate,
  exploração, social, sobrevivência), alvo (nenhum | célula | token | objeto), rolagens possíveis
  (lista de {instinto, aptidão} — a primeira é o padrão, o jogador pode trocar), animação (emote da
  miniatura + efeito no mapa), consequência (nenhuma | sugerir dano | revelar loot | descanso | aviso
  narrativo)`.
- **Padrões do Assimilação** (pares instinto+aptidão sugeridos, ajustáveis):
  | Ação | Categoria | Alvo | Rolagens sugeridas | Animação | Consequência |
  |---|---|---|---|---|---|
  | Atacar corpo a corpo | combate | token | Potência+Armas; Reação+Armas | investida + golpe + impacto | sugerir dano ao alvo (Sucessos) |
  | Atacar à distância | combate | token | Percepção+Armas | mira + projétil + impacto | sugerir dano |
  | Perceber | exploração | célula/nenhum | Percepção+Sobrevivência; Percepção+Segurança | olhar em volta + onda de "radar" | revela loot/segredos ocultos no raio conforme Sucessos |
  | Vasculhar / procurar item | exploração | objeto/célula | Sagacidade+Sobrevivência; Percepção+Manufaturas | agachar/mexer + brilho | encontra item escondido ali |
  | Esgueirar-se | exploração | nenhum | Reação+Furtividade | agachar + leve transparência | aviso ao Mestre |
  | Consertar / improvisar | sobrevivência | objeto | Sagacidade+Manufaturas; Sagacidade+Engenharia | martelar + faíscas | aviso |
  | Primeiros socorros | sobrevivência | token | Sagacidade+Medicina | ajoelhar + cruz brilhando | sugerir cura (reduz dano) |
  | Descansar | sobrevivência | nenhum (célula atual) | Resolução+Sobrevivência | sentar + fogueira/“zzz” | sugere ao Mestre recuperar Determinação/dano; em local seguro (Refúgio/fogueira) dá bônus |
  | Dançar / celebrar | social | nenhum | Influência+Expressão | dança + notas/confete | eleva Moral do Refúgio (sugestão) em festival |
  | Convencer / intimidar | social | token | Influência+Expressão; Potência+Expressão | gesto + balão | aviso |
- **Barra de ações** (na hotbar da 0002): botões por categoria; escolher ação → se pede alvo, o cursor
  vira mira e o jogador clica no token/célula/objeto (Esc cancela) → abre o **compositor** já preenchido
  (a rolagem padrão marcada e as alternativas como chips; modificadores disponíveis) → Rolar.
- **Fluxo no servidor:** `action.perform {characterId, actionId, target, rollChoice, modifiers}` valida
  alcance simples (alvo visível/na cena), cria a rolagem pelo `internal/rolls` (mesmo fluxo de manter
  dados) e emite `action.started` (animação começa para todos já na rolagem) e, ao finalizar a
  rolagem, `action.resolved` com o resultado e a consequência sugerida ao Mestre (reutilizar o fluxo de
  sugestões do `internal/effects` — confirmar/descartar).
- **Rolagem sem caixa de texto:** a aba Rolagens perde o campo `/r`. No lugar: (a) as ações; (b) para o
  Mestre, um **montador de parada** por botões (+d6 +d10 +d12, limpar, rolar) e atalhos genéricos
  (d6, 2d6, d10, d12, d20 — para outros sistemas); o comando `/r` continua aceito no backend, só sai da UI.
- **Rolagens sempre animadas:** toda rolagem (ação, genérica, ameaça) passa pela bandeja 3D; rolagens
  genéricas com dados que o renderizador não tem (d4, d8, d20) ganham geometria simples.

## Itens personalizados
- **Compêndio de itens da campanha** (Mestre): nome, descrição, ícone (lucide ou emoji) e cor, tipo
  (arma corpo a corpo, arma à distância, ferramenta, consumível, vestuário, recurso, chave/missão,
  outro), Qualidade, Escassez, Pressão, características (tags), e efeitos opcionais: bônus de dado
  mantido em ações de um tipo (ex.: arma +1 mantido em ataques), cura ao consumir, recurso do Refúgio.
  Armas definem a animação de ataque (corte, impacto, tiro, flecha) e aparecem na mão da miniatura.
- **Entregar item:** Mestre arrasta/atribui um item do compêndio para o inventário de um personagem
  (corpo/mochila, respeitando limites) ou para o Refúgio.
- **Loot no mapa:** Mestre esconde itens numa célula ou dentro de um objeto (caixote, carro, ruína) com
  dificuldade (nº de Sucessos para achar). Jogadores não veem até encontrar (Perceber/Vasculhar);
  ao achar, animação de brilho e o item vai para quem achou (ou fica disponível para pegar).
- **Usar item consumível** pela ficha/hotbar (ex.: bandagem → sugere cura ao Mestre).

## Animações (miniatura + mapa)
- **Emotes da miniatura** (procedurais, curtos, ~1–2 s): investida/golpe, mira/disparo, agachar,
  olhar em volta, ajoelhar, sentar/dormir, dançar, gesticular, comemorar, ser atingido (recuo +
  piscar), cair (Incapacitado).
- **Efeitos no mapa:** trilha de projétil, arco de golpe, faíscas de impacto, número de dano flutuante
  (quando aplicado), onda de percepção, brilho de item encontrado, “zzz”, notas musicais/confete,
  balão de fala.
- Sincronizados por evento para todos (mesma semente quando houver aleatoriedade), respeitando
  `prefers-reduced-motion` (versão estática: ícone/flash).

## Outras mecânicas de imersão (análise → o que entra agora)
| Ideia | Valor para a mesa | Custo | Decisão |
|---|---|---|---|
| Condições/status nos tokens (Ferido, Debilitado, Suscetível, Assustado, Escondido, Infectado) com ícones | alto: leitura rápida do estado | baixo | **entra** |
| Balões de fala e reações rápidas (emoji) nos tokens | alto: interpretação sem sair do mapa | baixo | **entra** |
| Dia/noite e clima (chuva, neblina, neve) controlados pelo Mestre; à noite tokens com lanterna/tocha iluminam um raio | alto: clima pós-apocalíptico | médio | **entra** |
| Efeitos sonoros procedurais (WebAudio, sem arquivos): dados, golpe, tiro, passos, chuva, fogueira; volume e mudo por jogador | médio-alto | médio | **entra** |
| Teste de Descanso do Assimilação com baralho de Assimilações (naipes por símbolo) | alto: mecânica central do sistema | médio | **entra** (consequência da ação Descansar, com o Mestre confirmando) |
| Holofote de turno: na fase da rodada, câmera de todos pode seguir o token ativo | médio | baixo | **entra** (opcional do Mestre) |
| Handouts: Mestre revela imagem/nota para todos ou para um jogador | médio | baixo | backlog |
| Trilha sonora/ambiente com arquivos | médio | depende de assets | backlog |
| Diário da sessão gerado a partir do log | médio | médio | backlog |

## Entregas
1. **Compêndio de itens + loot no mapa** (backend `internal/items`, front aba "Itens" do Mestre e loot).
2. **Ambiente e presença**: dia/noite, clima, luz de tocha à noite, condições nos tokens, balões de fala
   e reações, sons procedurais com controle de volume.
3. **Ações com animação** (depende da 0002 entrega 2 e de 1 para loot): catálogo de ações do
   Assimilação, barra de ações, mira de alvo, `action.perform`, emotes e efeitos, consequências via
   sugestões ao Mestre, montador de parada sem caixa de texto, rolagens sempre animadas, teste de
   Descanso com baralho de Assimilações, holofote de turno.
