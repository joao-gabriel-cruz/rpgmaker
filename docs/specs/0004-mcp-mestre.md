# 0004 — Claude como narrador da mesa (MCP do Mestre)

Status: DONE (v1) · extensão de NPCs READY · 2026-11-22

## Intenção

No meio da sessão, o Mestre narra e o mundo muda sozinho. Com o Claude (claude.ai ou Claude Desktop)
aberto ao lado da mesa, ele escreve coisas como *"e um chão de lava se abre"* ou *"o Junior cria uma
parede para se defender do ataque"*, e a cena ativa muda na hora para todos os jogadores. Ninguém
precisa abrir o editor nem pintar bloco a bloco, então o ritmo da narração não quebra.

A mesa expõe um **servidor MCP** que o Mestre conecta ao próprio Claude. O Claude lê o estado da mesa
(cena, terreno, objetos, tokens com nomes e posições) e chama ferramentas que aplicam as mudanças pelos
mesmos serviços e eventos em tempo real que o editor já usa.

**Prioridade (palavras do Mestre):** o centro é *alterar o terreno que os jogadores estão vendo* e
dar ao Mestre o poder de *criar coisas em tempo real na história*. Editar a cena ativa ao vivo é a
primeira entrega e precisa estar ótima: rápida, no lugar certo e desfazível. Criar e ativar cenas é
secundário e vem depois.

## Objetivos

- O Mestre conecta o Claude à campanha colando um link gerado no site e consegue alterar a mesa só
  narrando.
- O Claude entende referências espaciais da narração ("perto do Junior", "entre o Junior e o
  atacante", "ao norte") porque enxerga a mesa de forma legível.
- Escopo do que o Claude pode fazer (decidido: opção **B**):
  - terreno: tipo e altura das células;
  - objetos: colocar, remover, abrir/fechar porta;
  - mover tokens;
  - ambientação: hora do dia e clima;
  - condições nos personagens;
  - criar e ativar cenas (para quando os jogadores vão a um lugar que não estava preparado).
- As mudanças **aplicam direto na cena ao vivo**, sem rascunho. Cada mudança feita pelo Claude pode ser
  **desfeita** com um clique na mesa (ou pedindo ao próprio Claude).

## Não-objetivos (v1)

- Rolagens, dano e testes. O dado fica com a mesa, onde os jogadores veem e confiam.
- Login OAuth (fica como evolução, ver Abordagens).
- Efeitos visuais novos para as mudanças do Claude (ex.: animação da lava "se abrindo"). A mudança
  aparece como uma edição normal do mapa.
- Fichas: editar atributos, inventário e itens.
- Jogadores usando o Claude. Só o Mestre conecta.

## Anti-objetivos (falharia mesmo funcionando)

- O Mestre precisar parar a cena para revisar ou aprovar o que o Claude fez.
- Qualquer pessoa além do Mestre conseguir alterar a mesa pelo MCP.
- Uma mudança errada do Claude não poder ser revertida rápido.
- O MCP furar regras do servidor que o editor respeita (limites de tamanho, terrenos e prefabs
  válidos, células dentro do grid).

## Restrições

- Backend Go existente; módulo novo em `internal/mcp`, ligado em `cmd/api/wire.go` com `wireMCP(d deps)`.
- Limites do repo: arquivo ≤ 300 linhas, função ≤ 30 linhas; código em inglês, UI em PT-BR.
- Conectores personalizados do claude.ai aceitam só OAuth ou endereço sem login, por isso o segredo vai
  no caminho do link.
- Produção: o nginx da imagem web encaminha `/api/` para a API. O MCP fica sob `/api/mcp/...` para não
  mexer no roteamento. O transporte Streamable HTTP pode usar SSE, e o nginx bufferiza por padrão,
  então é preciso desligar o buffer para essa rota ou fazer o servidor responder JSON simples (o
  builder escolhe).
- Se a cena tiver um rascunho aberto, `scenes.Service.Paint/EditObjects` gravam no rascunho, que fica
  oculto para os jogadores. Isso contradiz "aplica direto" (ver Decisões).

## Abordagens consideradas

1. **MCP dentro do backend Go, com link secreto.** *(escolhida)* Endpoint na própria API com o SDK
   oficial de MCP para Go (`github.com/modelcontextprotocol/go-sdk`), transporte Streamable HTTP. As
   ferramentas chamam os serviços em processo e reaproveitam o broadcast. Prós: um deploy só, nada para
   instalar, funciona no claude.ai e no Desktop. Contra: quem tiver o link age como Mestre naquela
   campanha, então o link é tratado como senha (revogável).
2. **Igual ao 1, com OAuth.** Sem segredo no link, mas exige um servidor de autorização completo
   (registro dinâmico de clientes, PKCE, tokens), várias vezes o trabalho do MCP. Pode vir depois em
   cima do 1.
3. **MCP local em Node na máquina do Mestre.** Só funciona no Desktop, exige instalação, e a pintura do
   mapa passa pelo WebSocket. Descartada.

O Mestre escolheu a 1.

## Desenho

### Acesso: link do Mestre

- Tabela `mcp_links(id uuid, campaign_id, user_id, secret_hash, created_at, last_used_at, revoked_at)`.
  Guarda só o hash (sha256) do segredo, e o segredo tem 32 bytes aleatórios em base64url.
- **Um link ativo por campanha e Mestre.** Gerar um novo revoga o anterior.
- URL: `https://<domínio>/api/mcp/<segredo>`. O segredo identifica campanha e Mestre, então o Claude não
  precisa escolher campanha.
- A cada requisição, o servidor resolve o link, confere se não está revogado e se o usuário **ainda é
  Mestre** da campanha, e atualiza `last_used_at`. Link inválido → 404, sem revelar se existiu.
- As ações rodam como `realtime.Sender{UserID: <Mestre>, Role: GM, Name: "Claude"}`.
- REST (só Mestre): `GET /api/campaigns/{id}/mcp-link` (estado: existe, criado em, último uso),
  `POST .../mcp-link` (gera e devolve a URL **uma única vez**), `DELETE .../mcp-link` (revoga).

### Site

- **Tela "Conectar Claude"** nas configurações da campanha (só o Mestre vê): botão para gerar o link,
  mostrado uma vez com botão de copiar, último uso, botão de revogar e um passo a passo curto para
  claude.ai (Configurações → Conectores → adicionar conector personalizado) e para o Claude Desktop.
- **Na mesa (só o Mestre):** cada mudança do Claude vira um aviso discreto, como "Claude: lava em 14
  células · Desfazer". Um histórico curto (últimas mudanças) também fica acessível. Jogadores não veem
  esses avisos, só o mapa mudando.

### Ferramentas MCP

Nomes em inglês (identificadores); descrições escritas para o Claude, explicando coordenadas e
convenções. Todas operam na **cena ativa**, salvo as de cena.

Leitura:
- `get_table`: cena ativa (nome, tamanho), ambientação, tokens (nome, personagem, x/z, condições,
  jogador ou NPC), objetos (id, prefab, x/z, rotação, estado) e um **mapa em texto** (uma linha por z,
  um caractere por terreno, com legenda e alturas não-zero) para o Claude raciocinar sobre o espaço.
  Também explica a convenção de direções (qual eixo é "norte").
- `list_scenes`: cenas da campanha (id, nome, tamanho, ativa).
- `get_catalog`: terrenos e prefabs válidos, com o que bloqueia movimento e visão, tamanhos e
  condições disponíveis. Pode ir embutido nas descrições das ferramentas, se for mais simples.

Escrita (cada chamada vira **uma** mudança desfazível):
- `paint_terrain`: forma (`rect`, `circle`, `line` com espessura, ou lista de `cells`), `terrain`
  opcional e altura absoluta ou relativa (`height` / `heightDelta`, limitada a 0–10). O servidor
  rasteriza a forma em um `maps.Patch`.
- `place_objects`: lista de `{prefab, x, z, rotation, props}` e/ou `line {prefab, from, to}` (para
  paredes, cercas e barricadas em linha).
- `remove_objects`: por ids ou por área.
- `set_door`: abre ou fecha uma porta.
- `move_token`: por id de token ou nome do personagem, para x/z. Usa `force` como o Mestre (ignora
  alcance), mas respeita células bloqueadas.
- `set_ambience`: hora do dia (`dawn|day|dusk|night`) e/ou clima (`clear|rain|storm|fog|snow|ash`).
- `set_condition`: personagem + condição + ativa/inativa.
- `create_scene` (nome, largura, profundidade, terreno base opcional) e `activate_scene`.
- **Ativar leva o grupo** (decidido pelo Mestre): ao ativar uma cena pelo MCP, todo personagem de
  jogador (com dono) que ainda não tem token nela ganha um token perto de um ponto de chegada. O ponto é
  o parâmetro opcional `arrival {x,z}` (padrão: centro da cena), e os tokens se espalham pelas células
  livres mais próximas, sem cair em célula bloqueada. Quem já tem token na cena fica onde estava. Os
  tokens na cena anterior continuam lá. NPCs não são levados. Ativar pelo editor continua como hoje.
- `undo_last`: desfaz a última mudança do Claude nesta campanha.

Cada ferramenta de escrita devolve um resumo curto do que mudou ("14 células → lava em torno de
(12,8)"), que é o mesmo texto do aviso na mesa.

### Desfazer

- O servidor guarda, por campanha, as **últimas 20 mudanças** do Claude em memória. Cada uma tem o
  inverso dela: células com terreno e altura anteriores, objetos removidos ou adicionados, posição
  anterior do token, ambientação anterior e condição anterior.
- Desfazer aplica o inverso pelos mesmos serviços (com broadcast normal) e tira a mudança da lista.
- Se alguém editou as mesmas células depois, o desfazer sobrescreve essa edição. Isso é aceito,
  porque é o comportamento mais simples e previsível. Vale também para mudanças do próprio Claude que
  se sobrepõem: qualquer mudança da lista pode ser desfeita, em qualquer ordem, sem regra de conflito.
- O desfazer age na cena da mudança, mesmo que ela não seja mais a ativa.
- `create_scene` e `activate_scene` **não entram no desfazer**. Para voltar, o Claude ativa a cena
  anterior. Apagar cena não é ferramenta do MCP.
- Comando GMOnly no tempo real `mcp.undo {changeId}` para o botão da mesa, e evento ToGM
  `mcp.changed {changes:[{id, summary, at}]}` para os avisos e o histórico.

## Decisões tomadas por mim (corrigíveis)

- **Rascunho aberto:** as ferramentas de escrita recusam e respondem ao Claude "há um rascunho aberto
  nesta cena; aplique ou descarte no editor". Assim nada some num rascunho oculto.
- Desfazer guardado em memória: um reinício da API perde o histórico, o que é aceitável para o uso no
  meio da sessão.
- Limites por chamada: até 4096 células em `paint_terrain` e 200 objetos em `place_objects`, além dos
  limites que o serviço de cenas já aplica.
- O mapa em texto do `get_table` vai inteiro (até 128×128 = 16384 caracteres, desde que a main passou a aceitar cenas de 128).
- **Segredo nos logs:** a API não registra o caminho de `/api/mcp/...`, e o nginx da imagem web desliga
  o access log nessa rota. O log do Caddy da VPS fica fora do repo. É um risco aceito, mitigado pela
  revogação do link.

## Pontos do código que o plano precisa absorver

- `wire.go` cria instâncias separadas de `scenes.Service` por módulo, cada uma com seu mapa de locks e
  seu notifier (a de vision para cenas, o tap de effects para tokens). O `wireMCP` precisa reaproveitar
  as **mesmas instâncias** que o editor usa. Senão, edições do editor e do Claude se sobrescrevem, e os
  efeitos de névoa e terreno não disparam.
- `tokens.Service.Move` com `force` hoje pula todas as checagens. "Ignora alcance, mas respeita células
  bloqueadas" exige uma opção nova de movimento.

## Fica com o builder

- Formato exato do mapa em texto e da legenda, e o texto das descrições das ferramentas.
- SSE com buffer desligado no nginx ou respostas em JSON.
- Número da migration (segue a convenção do ticket).
- Onde exatamente a tela "Conectar Claude" e os avisos entram na UI atual.
- Como as ferramentas reaproveitam os serviços: chamar direto `scenes.Service`, `tokens.Service` e
  `ambience.Service` com o Sender do Mestre, expondo o mínimo necessário (ex.: um método que aplica um
  patch sempre na cena ao vivo).

## Extensão: NPCs vilões (2026-11-22, aprovada)

O Mestre narra *"surge um infectado brutamontes"* ou *"três infectados saem da mata ao norte"* e o Claude cria
os NPCs já prontos para combate e os põe na cena ativa.

- **`create_npc`**: `{name, count? (1–10), instincts?, knowledge?, practices?, boxesPerLevel?, threats?, notes?, at? {x,z}}`.
  Cria NPC(s) pela rota de personagens já existente (só o Mestre cria NPC). Os campos da ficha não informados ficam no
  padrão da ficha de NPC (Instintos 1, aptidões 0, 3 caixas por nível). Com `count > 1`, os nomes ficam "Nome 1…N",
  porque cada token na cena é de um personagem diferente. Com `at`, coloca os tokens nas células livres mais próximas
  (mesma busca em anéis do `activate_scene`), e a colocação entra no desfazer.
- **`place_token`**: `{character, x, z}`: põe um personagem existente (NPC ou jogador) na cena ativa, numa célula
  livre e transitável. Entra no desfazer (inverso: remover o token).
- **`remove_token`**: `{token}`: tira um token da cena ativa. Entra no desfazer (inverso: recolocar na mesma célula).
- **Desfazer uma criação** remove os tokens que ela colocou e **mantém o NPC na lista de personagens**: apagar a ficha
  automaticamente poderia levar junto ajustes do Mestre.
- Regras de sempre: só a cena ativa, recusa com rascunho aberto, célula bloqueada ou ocupada recusada, mensagens em PT-BR,
  mudança ao vivo para os jogadores e aviso com "Desfazer" só para o Mestre.
- Fora: retrato/miniatura do NPC e rolagens do NPC.
