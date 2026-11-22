package refuge

// ResourceKeys lists the 12 resources in display order; ResourceGroups splits them in 3 groups.
var ResourceKeys = []string{"water", "food", "medicine", "clothing", "plants", "animals", "wood", "biomass", "minerals", "fuel", "ammo", "buildingMaterial"}

// ResourceGroups maps each group (survival, organic, industry) to its resources.
var ResourceGroups = map[string][]string{
	"survival": {"water", "food", "medicine", "clothing"},
	"organic":  {"plants", "animals", "wood", "biomass"},
	"industry": {"minerals", "fuel", "ammo", "buildingMaterial"},
}

// NaturalKeys are the resources that have a regional abundance.
var NaturalKeys = []string{"water", "plants", "animals", "wood", "minerals", "biomass"}

// Abundances lists the abundance levels of a natural resource in the region.
var Abundances = []string{"none", "scarce", "medium", "abundant"}

// MaxLevel is the top of the 0-6 attribute scales.
const MaxLevel = 6

// DefenseLabels, MoraleLabels and BelligerenceLabels are the PT-BR names per level 0-6.
var (
	DefenseLabels      = []string{"Nenhuma", "Linhas e sinos", "Cerca de arame farpado", "Muro de madeira", "Muro de pedra/tijolo", "Complexo prisional", "Base militar"}
	MoraleLabels       = []string{"Amotinada", "Desiludida", "Hesitante", "Resoluta", "Animada", "Empenhada", "Exaltada"}
	BelligerenceLabels = []string{"Pacifista", "Mínima", "Razoável", "Eficiente", "Ameaçadora", "Terrível", "Arrasadora"}
)

// Derived values are computed on the server and never stored.
type Derived struct {
	DefenseLabel      string `json:"defenseLabel"`
	MoraleLabel       string `json:"moraleLabel"`
	BelligerenceLabel string `json:"belligerenceLabel"`
	Inhabitants       int    `json:"inhabitants"`
	Workers           int    `json:"workers"`
	ReservesTotal     int    `json:"reservesTotal"`
	ReservesFree      int    `json:"reservesFree"`
	ReservesOverflow  int    `json:"reservesOverflow"`
}

// IsResource reports whether key is one of the 12 resources.
func IsResource(key string) bool {
	for _, k := range ResourceKeys {
		if k == key {
			return true
		}
	}
	return false
}

func label(labels []string, level int) string {
	if level < 0 || level >= len(labels) {
		return ""
	}
	return labels[level]
}

// Workers is the working part of a population level: one in every 6 levels does not work.
func Workers(level int) int { return level - level/6 }

// ReservesTotal sums every stocked resource.
func (s Sheet) ReservesTotal() int {
	t := 0
	for _, v := range s.Resources {
		t += v
	}
	return t
}

// Derive computes labels, inhabitants, workers and the reserves figures.
func (s Sheet) Derive() Derived {
	total := s.ReservesTotal()
	d := Derived{
		DefenseLabel:      label(DefenseLabels, s.Defense),
		MoraleLabel:       label(MoraleLabels, s.Morale),
		BelligerenceLabel: label(BelligerenceLabels, s.Belligerence),
		Inhabitants:       s.Population.Level * 10,
		Workers:           Workers(s.Population.Level),
		ReservesTotal:     total,
	}
	if total > s.ReservesMax {
		d.ReservesOverflow = total - s.ReservesMax
	} else {
		d.ReservesFree = s.ReservesMax - total
	}
	return d
}

// AddResource stocks amount of key within the reserves cap and returns the discarded excess.
func (s *Sheet) AddResource(key string, amount int) (discarded int) {
	free := s.Derive().ReservesFree
	added := min(amount, free)
	s.Resources[key] += added
	return amount - added
}

// CapReserves cuts the stock down to ReservesMax, always taking from the largest stock
// (first in display order on ties), and returns how many units were lost per resource.
func (s *Sheet) CapReserves() map[string]int {
	lost := map[string]int{}
	for s.ReservesTotal() > s.ReservesMax {
		big := ""
		for _, k := range ResourceKeys {
			if big == "" || s.Resources[k] > s.Resources[big] {
				big = k
			}
		}
		s.Resources[big]--
		lost[big]++
	}
	return lost
}
