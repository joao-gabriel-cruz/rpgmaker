package refuge

// Building kinds; each is built at most once and can only be raised in level when Leveled.
const (
	KindApothecary   = "apothecary"
	KindTailor       = "tailor"
	KindCommunity    = "community"
	KindBiomass      = "biomassGathering"
	KindDormitory    = "dormitory"
	KindAmmoFactory  = "ammoFactory"
	KindWorksFactory = "worksFactory"
	KindFountain     = "waterSource"
	KindFortify      = "fortification"
	KindMill         = "mill"
	KindEquipment    = "equipmentWorkshop"
	KindMechanic     = "mechanicWorkshop"
	KindQuarry       = "quarry"
	KindPort         = "port"
	KindCanteen      = "canteen"
	KindRefineryAnim = "refineryAnimal"
	KindRefineryBio  = "refineryBiomass"
	KindRefineryVeg  = "refineryVegetal"
	KindReserves     = "expandedReserves"
	KindSawmill      = "sawmill"
)

// Effect sizes that the book leaves open [casa].
const (
	DormitoryPopulation = 5  // [casa] +5 to the population cap
	ExpandedReserves    = 10 // [casa] +10 to the reserves cap
	PortMobility        = 1  // [casa] +1 Mobility
	PortUpkeepWeeks     = 26 // semiannual upkeep
	PortUpkeepWood      = 5  // [casa] wood paid at each upkeep
	RefineryWeekly      = 10 // refineries convert at most 10 units per week
	MaterialPerTen      = 1  // [casa] 1 Material de Construção per 10 Pontos de Obra
	WoodPerMaterial     = 2  // [casa] 2 Madeira stand in for each missing Material
)

// Def is a catalog entry. Cost is in Pontos de Obra (per level when Leveled).
// Extract names the natural resource gathered by abundance; Input/Output a 1:1 conversion
// per allocated worker (capped by Cap when > 0); Doubles a resource whose weekly production is doubled.
type Def struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Effect  string `json:"effect"`
	Cost    int    `json:"cost"`
	Leveled bool   `json:"leveled"`
	Extract string `json:"extract,omitempty"`
	Input   string `json:"input,omitempty"`
	Output  string `json:"output,omitempty"`
	Cap     int    `json:"cap,omitempty"`
	Doubles string `json:"doubles,omitempty"`
}

// Catalog lists the 20 buildings in display order (fan sheet, with [casa] fillers noted).
var Catalog = []Def{
	{Kind: KindApothecary, Name: "Boticário", Cost: 10, Input: "plants", Output: "medicine", Effect: "Converte Plantas em Remédios (1 por trabalhador)"},
	{Kind: KindTailor, Name: "Casa de Costura", Cost: 10, Input: "plants", Output: "clothing", Doubles: "clothing", Effect: "[casa] Converte Plantas em Vestuário; +1 Vestuário a cada Vestuário produzido"},
	{Kind: KindCommunity, Name: "Centro Comunitário", Cost: 10, Leveled: true, Effect: "+1 Moral por nível"},
	{Kind: KindBiomass, Name: "Coleta de Biomassa", Cost: 20, Extract: "biomass", Effect: "Produz Biomassa conforme a abundância da região"},
	{Kind: KindDormitory, Name: "Dormitório", Cost: 20, Effect: "[casa] +5 no teto de População"},
	{Kind: KindAmmoFactory, Name: "Fábrica de Munição", Cost: 20, Input: "minerals", Output: "ammo", Effect: "Converte Minerais em Munição"},
	{Kind: KindWorksFactory, Name: "Fábrica para Obras", Cost: 20, Input: "minerals", Output: "buildingMaterial", Effect: "Converte Minerais em Material de Construção"},
	{Kind: KindFountain, Name: "Fonte de Água", Cost: 10, Extract: "water", Effect: "A água não é mais consumida do estoque; com trabalhadores, produz Água"},
	{Kind: KindFortify, Name: "Fortificação", Cost: 10, Leveled: true, Effect: "+1 Defesa por nível; exige 1 População alocada como guarda"},
	{Kind: KindMill, Name: "Moinho", Cost: 10, Input: "plants", Output: "food", Effect: "[casa] Converte Plantas em Alimento"},
	{Kind: KindEquipment, Name: "Oficina de Equipamentos", Cost: 10, Leveled: true, Effect: "[casa] Permite fabricar itens com Escassez até o nível (o Mestre usa o compêndio)"},
	{Kind: KindMechanic, Name: "Oficina Mecânica", Cost: 10, Effect: "Restaura veículos (efeito narrativo)"},
	{Kind: KindQuarry, Name: "Pedreira", Cost: 20, Extract: "minerals", Effect: "Produz Minerais conforme a abundância da região"},
	{Kind: KindPort, Name: "Porto", Cost: 20, Effect: "[casa] +1 Mobilidade; manutenção a cada 26 semanas: 5 Madeira"},
	{Kind: KindCanteen, Name: "Refeitório", Cost: 10, Doubles: "food", Effect: "Dobra o Alimento produzido"},
	{Kind: KindRefineryAnim, Name: "Refinaria Animal", Cost: 10, Input: "animals", Output: "fuel", Cap: RefineryWeekly, Effect: "Converte Animais em Combustível (até 10 por semana)"},
	{Kind: KindRefineryBio, Name: "Refinaria de Biomassa", Cost: 10, Input: "biomass", Output: "fuel", Cap: RefineryWeekly, Effect: "Converte Biomassa em Combustível (até 10 por semana)"},
	{Kind: KindRefineryVeg, Name: "Refinaria Vegetal", Cost: 10, Input: "plants", Output: "fuel", Cap: RefineryWeekly, Effect: "Converte Plantas em Combustível (até 10 por semana)"},
	{Kind: KindReserves, Name: "Reservas Expandidas", Cost: 10, Effect: "[casa] +10 no teto de Reservas"},
	{Kind: KindSawmill, Name: "Serraria", Cost: 10, Input: "wood", Output: "buildingMaterial", Effect: "Converte Madeira em Material de Construção"},
}

// DefOf finds a catalog entry.
func DefOf(kind string) (Def, bool) {
	for _, d := range Catalog {
		if d.Kind == kind {
			return d, true
		}
	}
	return Def{}, false
}

// CostAt is the Pontos de Obra needed to reach the given level.
func (d Def) CostAt(level int) int {
	if d.Leveled {
		return d.Cost * level
	}
	return d.Cost
}

// MaterialCost is the Material de Construção paid when work on a level starts [casa].
func (d Def) MaterialCost(level int) int { return d.CostAt(level) / 10 * MaterialPerTen }

// abundanceHalves is the extraction factor in halves: none 0, scarce 0.5, medium 1, abundant 2 [casa].
var abundanceHalves = map[string]int{"none": 0, "scarce": 1, "medium": 2, "abundant": 4}

// applyEffect adds delta levels of the building's lasting effect to the sheet (once, on completion).
func applyEffect(s *Sheet, kind string, delta int) {
	switch kind {
	case KindCommunity:
		s.Morale = min(MaxLevel, s.Morale+delta)
	case KindFortify:
		s.Defense = min(MaxLevel, s.Defense+delta)
	case KindDormitory:
		s.Population.Max = min(maxPopMax, s.Population.Max+DormitoryPopulation*delta)
	case KindReserves:
		s.ReservesMax += ExpandedReserves * delta
	case KindPort:
		s.Mobility = min(maxMobility, s.Mobility+PortMobility*delta)
	case KindFountain:
		s.HasWaterSource = true
	}
}
