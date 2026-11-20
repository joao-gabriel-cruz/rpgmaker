# Tickets: rpgmaker — MVP
Mesa virtual para Assimilação RPG, fatiada a partir da spec `docs/specs/0001-rpgmaker-mvp.md` (issue #1).

Trabalhe o **frontier**: qualquer ticket cujos bloqueadores estejam todos done. O E-2 (mapa) só depende da fundação e pode andar em paralelo com o E-1.

# E-0 · Fundação (#2)

## O projeto sobe ponta a ponta localmente
**GitHub:** #7
**O que construir:** Com um único comando (`docker compose up`) sobem o backend Go, o frontend React e o Postgres; a página inicial chama a API e mostra que está tudo de pé.
**Épico:** `E-0 · Fundação (#2)`

**Bloqueado por:** Nenhum — pode começar já
**Roteiro de QA:** QA1 [manual] Rodar o compose numa máquina limpa → a página inicial mostra "API ok · banco ok" · QA2 [auto] Teste do endpoint de saúde retorna sucesso com banco disponível e erro com banco fora
- [ ] Monorepo com backend Go, frontend React + TypeScript + Tailwind (Vite) e compose com Postgres
- [ ] Endpoint de saúde da API checa a conexão com o banco
- [ ] Migrations versionadas rodam automaticamente na subida
- [ ] Lint, testes e build passam nos dois lados com comandos documentados no README

## Admin entra e gerencia contas
**GitHub:** #8
**O que construir:** O admin é criado por comando, faz login com e-mail e senha, continua logado ao reabrir o navegador, sai quando quiser e consegue resetar a senha de qualquer usuário.
**Épico:** `E-0 · Fundação (#2)`

**Bloqueado por:** O projeto sobe ponta a ponta localmente (#7)
**Roteiro de QA:** QA1 [manual] Criar admin pelo comando e entrar → vê o painel com o próprio nome · QA2 [manual] Fechar e reabrir o navegador → continua logado; clicar Sair → volta ao login · QA3 [auto] Senha errada → "E-mail ou senha inválidos" sem revelar qual campo errou · QA4 [manual] Admin reseta senha de um usuário → o usuário entra com a senha temporária
- [ ] Comando cria a primeira conta admin
- [ ] Senhas com hash forte (argon2id ou bcrypt); sessão em cookie httpOnly persistida no banco
- [ ] Login, logout e "quem sou eu" funcionando no front
- [ ] Tela de admin lista usuários e gera nova senha temporária

## Mestre cria campanha e convida jogadores
**GitHub:** #9
**O que construir:** Um usuário cria uma campanha escolhendo o sistema de jogo e vira Mestre dela; gera um link de convite; quem abre o link se cadastra (e-mail + senha) e entra como jogador. Um mesmo usuário pode ser Mestre numa campanha e jogador em outra.
**Épico:** `E-0 · Fundação (#2)`

**Bloqueado por:** Admin entra e gerencia contas (#8)
**Roteiro de QA:** QA1 [manual] Criar campanha "Teste" com Assimilação → aparece na lista como Mestre · QA2 [manual] Abrir o link de convite em aba anônima, cadastrar → entra na campanha como Jogador · QA3 [auto] Link expirado ou já usado → "Convite inválido ou expirado" · QA4 [auto] Tentar cadastro sem convite → bloqueado
- [ ] Registro de módulos de sistema com a interface comum e o módulo Assimilação registrado (ainda sem regras)
- [ ] Criar campanha escolhendo sistema; criador vira Mestre
- [ ] Link de convite com expiração; cadastro só é possível via convite válido
- [ ] Lista de membros com papel por campanha; Mestre remove jogador

## Mesa em tempo real com presença
**GitHub:** #10
**O que construir:** Ao abrir a campanha, todos veem quem está online; a conexão em tempo real está pronta para carregar comandos e eventos com checagem de permissão no servidor.
**Épico:** `E-0 · Fundação (#2)`

**Bloqueado por:** Mestre cria campanha e convida jogadores (#9)
**Roteiro de QA:** QA1 [manual] Mestre e jogador em navegadores diferentes → cada um vê o outro online; fechar uma aba → some da lista · QA2 [auto] Usuário que não é membro tenta conectar → conexão recusada · QA3 [auto] Evento destinado só ao Mestre não chega ao jogador
- [ ] WebSocket por campanha autenticado pela sessão; só membros conectam
- [ ] Presença: entra/sai aparece para todos em tempo real
- [ ] Protocolo comando → evento com validação de permissão por papel e eventos filtrados por destinatário
- [ ] Reconexão automática do cliente

## App no ar na VPS atrás da Cloudflare
**GitHub:** #11
**O que construir:** O app fica acessível por HTTPS num domínio seu, rodando na VPS em Docker, atrás da Cloudflare, com backup diário do banco.
**Épico:** `E-0 · Fundação (#2)`

**Bloqueado por:** O projeto sobe ponta a ponta localmente (#7)
**Roteiro de QA:** QA1 [manual] Abrir o domínio no navegador → página carrega via HTTPS e a saúde da API está ok · QA2 [manual] Rodar o backup manualmente e restaurar num banco limpo → dados íntegros
- [ ] Levantar specs da VPS (SO, RAM, Docker) e domínio na Cloudflare
- [ ] Compose de produção e Cloudflare Tunnel (sem portas abertas) com HTTPS e WebSocket funcionando
- [ ] Backup diário do Postgres com retenção
- [ ] Passo a passo de deploy documentado

# E-1 · Jogável sem mapa (#3)

## Jogador cria e edita a ficha do Assimilação
**GitHub:** #12
**O que construir:** O jogador cria seu personagem com a ficha completa do Assimilação; a Saúde é calculada sozinha; os limites de criação bloqueiam ou só avisam conforme o Mestre configurou; o Mestre vê e edita todas as fichas, e as alterações aparecem ao vivo.
**Épico:** `E-1 · Jogável sem mapa (#3)`

**Bloqueado por:** Mesa em tempo real com presença (#10)
**Roteiro de QA:** QA1 [manual] Criar personagem com Potência 2 e Resolução 1 → cada nível de Saúde mostra 4 caixas · QA2 [manual] Com limites em "avisar", pôr 4 pontos extras em Instintos → aviso aparece mas salva · QA3 [auto] Com limites em "bloquear", o mesmo caso → não salva e mostra o motivo · QA4 [manual] Mestre altera a ficha do jogador → o jogador vê a mudança sem recarregar · QA5 [auto] Jogador tenta editar ficha de outro → recusado
- [ ] Módulo Assimilação: schema da ficha, validação com severidade e Saúde derivada (6 níveis × (1 + Potência + Resolução))
- [ ] Ficha: identidade, retrato, 6 Instintos, 12 Aptidões, Cabo de Guerra, Saúde, inventário (corpo/mochila), Características, Assimilações (texto livre)
- [ ] Configuração da campanha: limites de criação = bloquear | avisar
- [ ] Permissões: jogador edita só a própria ficha; Mestre edita todas; mudanças em tempo real

## Mestre mantém NPCs secretos
**GitHub:** #13
**O que construir:** O Mestre cria NPCs e criaturas com ficha simplificada que só ele vê.
**Épico:** `E-1 · Jogável sem mapa (#3)`

**Bloqueado por:** Jogador cria e edita a ficha do Assimilação (#12)
**Roteiro de QA:** QA1 [manual] Mestre cria NPC "Saqueador" → aparece na lista dele · QA2 [auto] Jogador consulta personagens da campanha → NPCs não vêm na resposta
- [ ] Ficha simplificada de NPC definida pelo módulo
- [ ] Lista de NPCs visível apenas ao Mestre

## Rolagem do Assimilação aparece no log
**GitHub:** #14
**O que construir:** O jogador clica num Instinto e numa Aptidão, o servidor sorteia a parada (d6 + d10), o jogador escolhe os dados mantidos e todos veem no log a contagem de Sucesso, Adaptação e Pressão, já com a penalidade de Saúde. O Mestre pode rolar em segredo e qualquer um pode usar `/r XdY+Z`.
**Épico:** `E-1 · Jogável sem mapa (#3)`

**Bloqueado por:** Jogador cria e edita a ficha do Assimilação (#12)
**Roteiro de QA:** QA1 [manual] Jogador com Reação 2 e Atletismo 1 rola → vê 2 d6 + 1 d10, escolhe 1 e o log mostra a contagem · QA2 [auto] Personagem em Laceração → o resultado mostra −1 Sucesso aplicado · QA3 [manual] Mestre faz rolagem secreta → só aparece no log dele · QA4 [auto] `/r 2d6+3` → resultado entre 5 e 15 registrado no log
- [ ] Sorteio no servidor com RNG criptográfico; tabela de faces d6/d10/d12 do módulo
- [ ] Escolha dos dados mantidos pelo jogador com sugestão destacada
- [ ] Penalidade de Saúde aplicada automaticamente e exibida
- [ ] Log persistido com visibilidade (pública, só Mestre); rolagem secreta do Mestre
- [ ] Rolador genérico `/r XdY+Z`
- [ ] Exibição 2D provisória dos dados com os símbolos

## Modificadores e sessão de jogo
**GitHub:** #15
**O que construir:** O jogador usa Empenho (desconta Determinação), Ocupação e Evento Marcante (1×/sessão) e faz rolagens assimiladas; o Mestre inicia e encerra a sessão de jogo e adiciona dados de dificuldade.
**Épico:** `E-1 · Jogável sem mapa (#3)`

**Bloqueado por:** Rolagem do Assimilação aparece no log (#14)
**Roteiro de QA:** QA1 [manual] Usar Empenho com Determinação 5 → rola mantendo 2 e a ficha fica com 4 · QA2 [auto] Usar Ocupação duas vezes na mesma sessão → a segunda é recusada · QA3 [manual] Mestre encerra e inicia nova sessão → Ocupação volta a ficar disponível · QA4 [auto] Rolagem assimilada mantém no mínimo 2 dados e não permite descartar Pressão mantida
- [ ] Empenho: +1 dado mantido e −1 Determinação na ficha
- [ ] Ocupação e Evento Marcante: +1 mantido cada, bloqueados após uso até a próxima sessão
- [ ] Mestre inicia/encerra sessão de jogo; encerrar reseta os usos
- [ ] Rolagem assimilada: dois Instintos, d6 viram d12, mínimo 2 mantidos, Pressões mantidas não descartáveis
- [ ] Mestre adiciona 0–3 d12 de dificuldade a uma rolagem

## Dados 3D caem no resultado sorteado
**GitHub:** #16
**O que construir:** As rolagens viram dados 3D (d6, d10, d12) com faces de símbolos próprios; todos os participantes veem a mesma animação terminar exatamente nas faces que o servidor sorteou.
**Épico:** `E-1 · Jogável sem mapa (#3)`

**Bloqueado por:** Rolagem do Assimilação aparece no log (#14)
**Roteiro de QA:** QA1 [manual] Rolar com 3 navegadores abertos → todos veem os dados pararem nas mesmas faces e o log confere · QA2 [manual] Rolar 20 vezes → a face visível sempre bate com o resultado do log
- [ ] Dados 3D com React Three Fiber e faces customizadas (ícones próprios, sem arte do livro)
- [ ] Animação termina na face sorteada pelo servidor (determinística)
- [ ] Mesma animação em todos os clientes; rolagem secreta anima só para o Mestre
- [ ] Fallback 2D se o 3D falhar

# E-2 · Mapa 3D (#4)

## Mestre cria cenas e pinta terreno em blocos 3D
**GitHub:** #17
**O que construir:** O Mestre cria cenas e pinta o terreno numa grade 3D de blocos (tipo + altura) com pincéis; escolhe a cena ativa, que os jogadores veem ao vivo com câmera livre.
**Épico:** `E-2 · Mapa 3D (#4)`

**Bloqueado por:** Mesa em tempo real com presença (#10)
**Roteiro de QA:** QA1 [manual] Mestre pinta um pico de altura 5 → jogador vê o pico aparecer sem recarregar · QA2 [manual] Mestre troca a cena ativa → a tela dos jogadores muda para a nova cena · QA3 [auto] Jogador tenta pintar terreno → recusado
- [ ] Cenas por campanha com tamanho de grade; cena ativa
- [ ] Célula = tipo de terreno + altura; persistência eficiente
- [ ] Pincéis: tipo, subir/descer altura, preencher área, apagar
- [ ] Tipos de terreno padrão (chão, água, vale, pico, lava…) com regras padrão (bloqueia movimento/visão)
- [ ] Câmera livre por pessoa (girar, zoom, arrastar)
- [ ] Mudanças aparecem ao vivo para todos

## Mestre coloca e remove objetos 3D ao vivo
**GitHub:** #18
**O que construir:** O Mestre escolhe objetos numa paleta (muro arrastável, vulcão com lava, árvore, carro, barricada…) e os coloca ou remove no mapa durante o jogo.
**Épico:** `E-2 · Mapa 3D (#4)`

**Bloqueado por:** Mestre cria cenas e pinta terreno em blocos 3D (#17)
**Roteiro de QA:** QA1 [manual] Mestre arrasta um muro de 6 células → aparece para os jogadores · QA2 [manual] Mestre coloca um vulcão → lava animada visível; remove → some para todos
- [ ] Pacote CC0 low-poly embutido e catálogo de prefabs
- [ ] Muro desenhado arrastando ao longo da grade
- [ ] Colocar, girar e remover objetos; objetos com regras (bloqueia movimento/visão)
- [ ] Mudanças ao vivo para todos

## Jogadores movem seus tokens no mapa
**GitHub:** #19
**O que construir:** Cada personagem vira uma plaquinha 3D com o retrato da ficha; o jogador move o seu, o Mestre move todos; terreno que bloqueia impede a passagem; qualquer um pode dar ping e o Mestre pode focar a câmera de todos num ponto.
**Épico:** `E-2 · Mapa 3D (#4)`

**Bloqueado por:** Jogador cria e edita a ficha do Assimilação (#12), Mestre cria cenas e pinta terreno em blocos 3D (#17)
**Roteiro de QA:** QA1 [manual] Jogador arrasta o próprio token → todos veem mover · QA2 [auto] Mover token para célula com muro → recusado e o token volta · QA3 [manual] Mestre usa "focar todos" → câmera de todos vai para o ponto
- [ ] Token standee com retrato; encaixe na grade
- [ ] Permissões de movimento (jogador: o seu; Mestre: todos)
- [ ] Bloqueio de movimento validado no servidor (terreno/objeto/diferença de altura)
- [ ] Ping visível a todos; "focar todos aqui" do Mestre

## Névoa manual e visão dinâmica
**GitHub:** #20
**O que construir:** O Mestre revela e esconde áreas manualmente; além disso, cada jogador só vê o que o seu token enxerga, considerando relevo e objetos que bloqueiam visão, com o cálculo feito no servidor.
**Épico:** `E-2 · Mapa 3D (#4)`

**Bloqueado por:** Mestre coloca e remove objetos 3D ao vivo (#18), Jogadores movem seus tokens no mapa (#19)
**Roteiro de QA:** QA1 [manual] Token atrás de um muro alto → o outro lado fica escondido para aquele jogador · QA2 [manual] Mestre revela uma área manualmente → todos os jogadores passam a vê-la · QA3 [auto] Jogador não recebe dados de células fora da sua visão
- [ ] Névoa manual (revelar/esconder área) pelo Mestre
- [ ] Linha de visão por token com altura, calculada no servidor por jogador
- [ ] Eventos de mapa filtrados pela visão de cada jogador
- [ ] Recalcular só ao mover token ou alterar terreno/objetos

# E-3 · Mapa vivo (#5)

## Mestre prepara mudanças escondidas e revela de uma vez
**GitHub:** #21
**O que construir:** O Mestre edita o mapa num modo rascunho que os jogadores não veem e aplica tudo de uma vez; também salva estados da cena ("rodada 3: erupção") e os dispara quando quiser.
**Épico:** `E-3 · Mapa vivo (#5)`

**Bloqueado por:** Mestre coloca e remove objetos 3D ao vivo (#18)
**Roteiro de QA:** QA1 [manual] Mestre ativa rascunho e coloca um vulcão → jogadores não veem; aplica → vulcão aparece para todos · QA2 [manual] Salvar estado "erupção", voltar o mapa ao normal e disparar → mapa vai para o estado salvo
- [ ] Camada de rascunho visível só ao Mestre; aplicar/descartar
- [ ] Estados de cena nomeados (snapshots) salvos e disparados
- [ ] Transição simples ao aplicar

## Controle de rodada visível para todos
**GitHub:** #22
**O que construir:** O Mestre conduz a rodada pelas fases do Assimilação (preparação → infectados → assimilador → fim) e todos veem em que fase e rodada estão.
**Épico:** `E-3 · Mapa vivo (#5)`

**Bloqueado por:** Mesa em tempo real com presença (#10)
**Roteiro de QA:** QA1 [manual] Mestre avança para "infectados" → todos veem a fase mudar · QA2 [auto] Jogador tenta avançar a fase → recusado
- [ ] Contador de rodada e fase atual na tela de todos
- [ ] Mestre avança/volta fase e reinicia
- [ ] Evento de início de turno disponível para outros módulos (efeitos de terreno)

## Efeitos de terreno sugeridos ao Mestre
**GitHub:** #23
**O que construir:** Terrenos e objetos têm regras "ao entrar" e "ao começar o turno" (dano na Saúde, +d12 de dificuldade, aviso); quando disparam, o Mestre recebe a sugestão e confirma ou descarta.
**Épico:** `E-3 · Mapa vivo (#5)`

**Bloqueado por:** Modificadores e sessão de jogo (#15), Jogadores movem seus tokens no mapa (#19), Controle de rodada visível para todos (#22)
**Roteiro de QA:** QA1 [manual] Token entra na lava (regra: 2 de dano) → Mestre recebe sugestão; confirma → Saúde do personagem cai 2 · QA2 [manual] Mestre descarta a sugestão → nada muda na ficha · QA3 [auto] Regra "+1 d12" confirmada → a próxima rolagem daquele personagem inclui 1 d12 de dificuldade
- [ ] Editor de regras por tipo de terreno/objeto (padrões + override por campanha)
- [ ] Disparo ao entrar na célula e no início do turno
- [ ] Sugestão ao Mestre; confirmar aplica dano na ficha ou d12 na próxima rolagem
- [ ] Sem limite de movimento automático

## Medição de distância e desenho livre
**GitHub:** #24
**O que construir:** Qualquer participante mede distância entre células e desenha à mão livre sobre o mapa; o Mestre apaga desenhos.
**Épico:** `E-3 · Mapa vivo (#5)`

**Bloqueado por:** Mestre cria cenas e pinta terreno em blocos 3D (#17)
**Roteiro de QA:** QA1 [manual] Medir de um token a outro 5 células adiante → mostra 5 · QA2 [manual] Jogador desenha uma seta → todos veem; Mestre apaga → some
- [ ] Régua em células sobre o terreno 3D
- [ ] Desenho livre visível a todos; apagar o próprio, Mestre apaga todos

# E-4 · Partitura e Refúgio (#6)

## Partitura de Conflito na mesa
**GitHub:** #25
**O que construir:** O Mestre monta a Partitura da cena (condicionantes, objetivos, ameaças, ativações); os Sucessos das rolagens avançam os objetivos; ameaças têm parada própria sorteada no servidor e estão ligadas a tokens no mapa.
**Épico:** `E-4 · Partitura e Refúgio (#6)`

**Bloqueado por:** Modificadores e sessão de jogo (#15), Jogadores movem seus tokens no mapa (#19), Controle de rodada visível para todos (#22)
**Roteiro de QA:** QA1 [manual] Objetivo exige 5 Sucessos; jogador rola 2 e atribui → progresso 2/5 visível a todos · QA2 [manual] Mestre rola a parada de uma ameaça → resultado no log e token da ameaça destacado
- [ ] Partitura por cena: condicionantes, objetivos principais/secundários com progresso
- [ ] Ameaças com parada própria e token vinculado
- [ ] Ativações disparadas por Sucesso/Adaptação/Pressão
- [ ] Rolagens dos jogadores podem ser atribuídas a um objetivo

## Refúgio do grupo
**GitHub:** #26
**O que construir:** O grupo tem um Refúgio com Defesa, Moral, Mobilidade, Beligerância, recursos (água, alimento, madeira) e inventário compartilhado.
**Épico:** `E-4 · Partitura e Refúgio (#6)`

**Bloqueado por:** Jogador cria e edita a ficha do Assimilação (#12)
**Roteiro de QA:** QA1 [manual] Jogador move um item da mochila para o Refúgio → aparece no inventário do Refúgio para todos · QA2 [manual] Mestre reduz Água em 2 → todos veem o novo valor
- [ ] Ficha do Refúgio por campanha
- [ ] Inventário do Refúgio integrado ao inventário das fichas
- [ ] Mestre edita atributos; jogadores veem e movem itens
