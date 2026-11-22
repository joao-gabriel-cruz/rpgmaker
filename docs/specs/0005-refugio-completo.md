# Spec 0005 — Refúgio completo (construção, suprimentos, população, crises)

- **Status:** READY (aprovada pelo usuário em 2026-11-21)
- **Fonte das regras:** pesquisa de 2026-11-21. Livro e site oficial não acessíveis; a fonte principal é o
  código da ficha de fã LPfontes/ficha-assimilacao (cita pp. ~103–110 do livro) e o app de Roan Lisboa.
  Marcado como **[fã]** o que vem dessas ferramentas e **[casa]** o que é regra nossa para preencher lacuna.
  Validar contra o livro quando possível.

## Hoje no app × regras pesquisadas
| Item | Hoje | Regras pesquisadas |
|---|---|---|
| Defesa, Moral, Beligerância | 0–10 sem rótulo | 0–6 com rótulo por nível [fã] |
| Mobilidade | 0–10 | nível ≈ quanto de População dá para transportar [fã] |
| População | não existe | níveis (~10 habitantes cada), teto `populacaoMax`; 1 a cada 6 níveis não trabalha [fã] |
| Reservas | não existe | soma dos recursos estocados contra um teto; excedente se perde [fã] |
| Recursos | Água, Alimento, Madeira | 12: Água, Alimento, Remédios, Vestuário, Plantas, Animais, Madeira, Biomassa, Minerais, Combustível, Munição, Material de Construção [fã] |
| Consumo | não existe | por semana × População: 1 Água, 1 Alimento, 1 Madeira; por ano: 1 Remédio; a cada 2 anos: 1 Vestuário [fã] |
| Construções | não existem | ~19 construções com custo em Pontos de Obra (10/20, ou 10×nível) [fã] |
| Crises | não existem | gatilhos (superlotação, defesa < perigo, falta de recurso) + rolagem de gravidade [fã] |
| Região | não existe | Tamanho, Perigo, Habitação, Recursos Naturais, Contaminação, Deslocamento (0–6) + abundância por recurso [fã] |
| Moradores individuais | não existem | sistema trata em níveis de População, não NPCs [fã] |
| Expedições | não existem | sem regra pública encontrada |

## Proposta
1. **Ficha do Refúgio revisada:** Defesa/Moral/Beligerância 0–6 com rótulos (Defesa: Nenhuma → Linhas e
   sinos → Arame farpado → Muro de madeira → Muro de pedra → Complexo prisional → Base militar; Moral:
   Amotinada → … → Exaltada; Beligerância: Pacifista → … → Arrasadora), Mobilidade, População (nível e teto,
   habitantes estimados, trabalhadores = nível − ⌊nível/6⌋), Reservas (total e teto), "tem Fonte de Água".
   Migração dos valores atuais (0–10 → 0–6, arredondando).
2. **12 recursos** em 3 grupos (Sobrevivência, Orgânicos, Indústria), com o teto de Reservas.
3. **Construções:** catálogo com custo, nível, consumo, produção e efeito: Boticário, Casa de Costura,
   Centro Comunitário (+1 Moral por nível), Coleta de Biomassa, Dormitório (+teto de População), Fábrica de
   Munição, Fábrica para Obras, Fonte de Água, Fortificação (+1 Defesa por nível, exige 1 População de
   guarda), Moinho, Oficina de Equipamentos (itens até a Escassez do nível), Oficina Mecânica, Pedreira,
   Porto (+Mobilidade, manutenção semestral), Refeitório (dobra Alimento), Refinarias (Combustível),
   Reservas Expandidas (+teto de Reservas), Serraria [fã]. Construir consome **Pontos de Obra**, que vêm do
   trabalho da População alocada à obra (1 por População por semana) [fã]. Jogadores propõem, Mestre aprova.
4. **Semana do Refúgio (Mestre: "Avançar semana"):** aplica produção das construções (conversões 1 por
   População alocada por semana), soma Pontos de Obra às obras em andamento, consome recursos × População
   (marcos anual/bienal por contador de semanas), corta o excedente acima do teto de Reservas, verifica
   gatilhos de crise e publica um **relatório da semana** para todos (o que entrou, saiu, faltou). Tudo
   revisável antes de confirmar.
5. **Crises:** lista com grau e status (Ativa, Contida, Resolvida). Gatilhos detectados na semana viram
   crise sugerida ao Mestre. **Rolagem de gravidade** na bandeja 3D: 4 dados, descarta 1, conta os símbolos
   dos 3 restantes: 0–1 nada; 2 briga interna (metade da População migra com metade das reservas); 3
   doença (−1 dado em todos os testes até a cura); 4 criatura assimilada grande; 5 rebelião em duas
   facções; 6+ ataque de invasores [fã]. Tipo de dado e quem descarta não confirmados — padrão d6 e o
   Mestre escolhe o descarte [casa]. Cada crise tira 1 de Moral (opção ligada por padrão) [fã].
6. **Região do Refúgio:** os 6 atributos 0–6 + abundância por recurso (Inexistente/Escasso/Médio/Abundante).
   Usada pela Coleta/Pedreira/Fonte e pelo gatilho "Defesa < Perigo" [fã].
7. **Expedição (regra da casa):** ação de grupo "Buscar suprimentos" a partir da mesa: escolhe recurso alvo
   e região; rolagem com Percepção/Sagacidade + Sobrevivência; Sucessos viram unidades do recurso
   conforme a abundância da região, Adaptações trazem um extra ou um item, Pressões geram complicação
   (o Mestre escolhe: ferimento, encontro, perda) [casa]. Resultado entra nas Reservas.
8. **Moradores notáveis (regra da casa, opcional):** a População continua abstrata, mas o Mestre pode
   marcar NPCs como "moradores do Refúgio" com função (guarda, coletor, médico, artesão) para aparecerem
   na aba e reagirem a crises — sem mexer nas contas da População [casa].
9. **Mapa:** marcar uma cena como "Refúgio" (Descansar seguro já usa isso) e opção de criar a cena a partir
   do modelo "Refúgio Fortificado".

## Entregas
1. Ficha revisada + 12 recursos + região (migração dos dados atuais).
2. Construções + Pontos de Obra + semana do Refúgio com relatório.
3. Crises com gatilhos e rolagem de gravidade + expedição + moradores notáveis.
