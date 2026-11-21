package effects

var defaults = map[string][]Rule{
	"terrain:lava": {
		{TriggerEnter, EffectDamage, 2, "A lava queima!"},
		{TriggerTurnStart, EffectDamage, 2, "O calor da lava continua queimando."},
	},
	"terrain:water": {{TriggerEnter, EffectNotice, 0, "Terreno difícil"}},
	"terrain:snow":  {{TriggerTurnStart, EffectDifficulty, 1, "O frio dificulta a próxima ação."}},
	"prefab:campfire": {
		{TriggerEnter, EffectDamage, 1, "As chamas da fogueira queimam."},
	},
}

var labels = map[string]string{
	"terrain:floor": "Piso", "terrain:grass": "Grama", "terrain:dirt": "Terra", "terrain:sand": "Areia",
	"terrain:water": "Água", "terrain:stone": "Pedra", "terrain:snow": "Neve", "terrain:lava": "Lava",
	"terrain:ruin": "Ruína",
	"prefab:wall":  "Parede", "prefab:barricade": "Barricada", "prefab:tree": "Árvore", "prefab:bush": "Arbusto",
	"prefab:rock": "Rocha", "prefab:car": "Carro", "prefab:crate": "Caixote", "prefab:campfire": "Fogueira",
	"prefab:volcano": "Vulcão", "prefab:watchtower": "Torre de vigia", "prefab:post": "Poste",
	"prefab:fence": "Cerca", "prefab:door": "Porta",
}

func id(kind, key string) string { return kind + ":" + key }

func labelOf(kind, key string) string {
	if l, ok := labels[id(kind, key)]; ok {
		return l
	}
	return key
}

func cloneRules(r []Rule) []Rule { return append([]Rule{}, r...) }
