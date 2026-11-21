# Spec 0004 — Modelos de mapa para o Mestre

- **Status:** READY (pedido do usuário em 2026-11-21: "vários modelos de mapas onde o mestre pode usar
  para sua história, todos com no mínimo 64x64")

## Comportamento
- Ao criar uma cena, o Mestre escolhe **"Em branco"** (tamanho livre 16–128) ou um **modelo** numa
  galeria: cartão com miniatura vista de cima, nome, descrição curta (gancho de história), tamanho,
  etiquetas (urbano, natureza, refúgio, combate, social, perigo) e ambiente sugerido.
- Modelos são **gerados por código no servidor** (determinísticos por `seed`): o mesmo modelo com outra
  semente dá uma **variação** ("Gerar outra variação" na galeria, com prévia atualizada antes de criar).
- Ao criar a partir de um modelo: a cena nasce com o terreno, os objetos e o **ambiente sugerido**
  (hora do dia/clima, aplicado via `internal/ambience`). Tudo continua editável normalmente.
- **Tamanho mínimo dos modelos: 64×64.** Limite de cena sobe para **128×128** (migration ajustando o
  CHECK, validação no service, seletor de tamanho no front, conferindo desempenho de renderização,
  visão e névoa em 128×128).

## Contrato do backend (pacote `internal/maptemplates`)
```go
type Ambience struct{ TimeOfDay, Weather string }      // valores do internal/ambience
type Template struct {
    ID, Name, Description string                       // PT-BR na Name/Description
    Tags     []string
    Width, Depth int                                   // >= 64, <= 128
    Ambience Ambience
    Generate func(seed int64) (maps.Grid, []maps.Object)
}
func Register(t Template)                              // cada modelo num arquivo tpl_<id>.go com init()
func All() []Template; func Get(id string) (Template, bool)
```
- Helpers compartilhados em `builder.go` (ruído value/simplex simples com seed, pintar retângulo/círculo/
  linha grossa, estrada, rio sinuoso com margens, elevação suave/colina, construção retangular com
  muros + porta + interior, espalhar árvores/arbustos/pedras por densidade evitando estradas e
  construções, cercas, garantir caminhos livres). Todos os objetos devem passar na validação de
  `maps` (prefab conhecido, pegada no grid, sem sobreposição bloqueante absurda).
- REST: `GET /api/campaigns/{id}/map-templates` (Mestre) lista metadados; `GET .../map-templates/{tid}/preview?seed=` devolve
  `{grid (serializado compacto), objects}` para a miniatura; criação: `POST .../scenes` aceita
  `{name, templateId, seed}` além do modo em branco.

## Qualidade dos modelos
- Legíveis e jogáveis: caminhos claros, pontos de interesse (onde esconder loot, onde emboscar, onde
  descansar), variedade de altura, sem "ruído" sem sentido. Estética pós-apocalíptica do Assimilação.
- Cada gerador tem teste: dimensões ≥64, objetos válidos para `maps`, determinismo (mesma seed → mesmo
  resultado; seeds diferentes → resultados diferentes), sem objetos bloqueando a área de entrada.

## Modelos
| # | ID | Nome | Tamanho | Ambiente | Ideia |
|---|---|---|---|---|---|
| 1 | vila-abandonada | Vila Abandonada | 64×64 | entardecer, limpo | casas de muros com portas, praça de terra, rio, morro, carros largados |
| 2 | rodovia-engarrafada | Rodovia Engarrafada | 96×64 | dia, neblina | rodovia longa com carros/barricadas, posto, mato dos dois lados |
| 3 | cidade-em-ruinas | Cidade em Ruínas | 96×96 | dia, cinzas | quarteirões de prédios (ruína), ruas, entulho, praça central |
| 4 | floresta-clareira | Floresta Densa | 64×64 | dia, limpo | mata fechada, riacho, clareira com acampamento escondido |
| 5 | fazenda | Sítio Isolado | 80×80 | amanhecer, limpo | casa, celeiro, plantações (fileiras de terra), cercas, poço |
| 6 | refugio-escola | Refúgio Fortificado | 64×64 | noite, limpo | escola murada, torres, sacos de areia, horta, fogueira |
| 7 | festival | Festival da Comunidade | 64×64 | noite, limpo | praça com barracas de feira, fogueira central, postes de luz, bandeirolas |
| 8 | vale-vulcanico | Vale Vulcânico | 96×96 | entardecer, cinzas | vulcão, rios de lava, rochas, terreno alto e perigoso |
| 9 | passo-nevado | Passo Nevado | 80×80 | dia, neve | montanhas, desfiladeiro, cabana, pinheiros |
| 10 | litoral-pier | Litoral e Píer | 64×96 | dia, limpo | praia, mar, píer (ponte), barracos de pescador |
| 11 | pantano | Pântano | 64×64 | noite, neblina | poças, lama, árvores mortas, passarelas |
| 12 | desfiladeiro-ponte | Desfiladeiro da Ponte | 64×96 | dia, limpo | cânion com rio no fundo, ponte única, emboscada nas bordas |

## Novos prefabs (objetos procedurais low-poly, mesmas regras do catálogo)
`tent` (barraca de acampamento, 2×2, bloqueia movimento), `barrel` (1×1), `bench` (2×1, bloqueia movimento),
`market_stall` (barraca de feira 2×2, com toldo colorido via prop `color`), `lamp_post` (poste com luz — à
noite ilumina como fogueira), `sandbags` (1×1 baixo, bloqueia movimento, não visão), `ruined_wall` (muro
quebrado, altura 1–2 via prop), `well` (poço 1×1), `sign` (placa), `bridge` (tábua 1×1 apoiada na
altura do terreno vizinho — permite atravessar água/vão; não bloqueia), `bunting` (bandeirolas entre
postes, decorativo). Backend (`maps.objects`) e frontend (`map3d/objects`) com ids iguais; paleta do
editor atualizada.
