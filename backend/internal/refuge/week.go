package refuge

import (
	"fmt"
	"maps"
	"slices"
)

// Weekly rules: consumption per population level, and the yearly/biennial milestones.
const (
	YearWeeks    = 52  // 1 Remédio per level
	BiennialWeek = 104 // 1 Vestuário per level
	historyKeep  = 20
	maxAdjust    = 9999
)

// CrisisTrigger is a condition found while closing a week that may become a crisis.
type CrisisTrigger struct {
	Kind     string `json:"kind"` // "overpopulation" | "breach" | "shortage"
	Detail   string `json:"detail"`
	Resource string `json:"resource,omitempty"`
}

// WorkLine reports the construction progress of one building in the week.
type WorkLine struct {
	BuildingID string `json:"buildingId"`
	Name       string `json:"name"`
	Level      int    `json:"level"`
	Points     int    `json:"points"`
	Progress   int    `json:"progress"`
	Cost       int    `json:"cost"`
	Done       bool   `json:"done"`
	Blocked    string `json:"blocked,omitempty"`
}

// WeekReport is what the week brought: Entered/Spent are net flows per resource, Missing what
// was needed and absent, Lost what exceeded the reserves cap, Adjustments the Mestre's manual deltas.
type WeekReport struct {
	Week        int             `json:"week"`
	Entered     map[string]int  `json:"entered"`
	Spent       map[string]int  `json:"spent"`
	Missing     map[string]int  `json:"missing"`
	Lost        map[string]int  `json:"lost"`
	Adjustments map[string]int  `json:"adjustments"`
	Works       []WorkLine      `json:"works"`
	Completed   []string        `json:"completed"`
	Triggers    []CrisisTrigger `json:"triggers"`
	Notes       []string        `json:"notes"`
}

type weekSim struct {
	r   *Refuge
	rep WeekReport
}

func newReport(week int) WeekReport {
	return WeekReport{Week: week, Entered: map[string]int{}, Spent: map[string]int{}, Missing: map[string]int{},
		Lost: map[string]int{}, Adjustments: map[string]int{}, Works: []WorkLine{}, Completed: []string{},
		Triggers: []CrisisTrigger{}, Notes: []string{}}
}

func (w *weekSim) gain(key string, n int) {
	if n > 0 {
		w.r.Resources[key] += n
		w.rep.Entered[key] += n
	}
}

func (w *weekSim) spend(key string, n int) {
	w.r.Resources[key] -= n
	w.rep.Spent[key] += n
}

// Simulate computes the closing of the week on a copy: the next state and the report.
// adjust holds the Mestre's manual resource deltas, applied after production and before consumption.
func Simulate(in Refuge, adjust map[string]int) (Refuge, WeekReport) {
	r := in
	r.Resources = maps.Clone(in.Resources)
	r.Buildings = slices.Clone(in.Buildings)
	w := &weekSim{r: &r, rep: newReport(r.Week)}
	active := slices.DeleteFunc(slices.Clone(r.Buildings), func(b Building) bool { return b.State != StateDone })
	w.runWorks()
	w.produce(active)
	w.adjust(adjust)
	w.consume(active)
	maps.Copy(w.rep.Lost, r.CapReserves())
	w.checkTriggers()
	r.Week++
	return r, w.rep
}

// ValidateAdjust rejects unknown resources and absurd deltas.
func ValidateAdjust(adjust map[string]int) error {
	for k, v := range adjust {
		if !IsResource(k) || v < -maxAdjust || v > maxAdjust {
			return ErrInvalid
		}
	}
	return nil
}

func (w *weekSim) adjust(adjust map[string]int) {
	for _, k := range ResourceKeys {
		d, ok := adjust[k]
		if !ok || d == 0 {
			continue
		}
		d = max(d, -w.r.Resources[k])
		w.r.Resources[k] += d
		w.rep.Adjustments[k] = d
	}
}

// guardless is the Defense bonus lost because a Fortificação has nobody allocated as guard.
func guardless(bs []Building) int {
	lost := 0
	for _, b := range bs {
		if b.Kind == KindFortify && b.Built > 0 && b.Workers < 1 {
			lost += b.Built
		}
	}
	return lost
}

// EffectiveDefense is the Defense that counts against threats.
func EffectiveDefense(r Refuge) int { return max(0, r.Defense-guardless(r.Buildings)) }

func (w *weekSim) checkTriggers() {
	r := w.r
	if r.Population.Level > r.Population.Max {
		w.rep.Triggers = append(w.rep.Triggers, CrisisTrigger{Kind: "overpopulation",
			Detail: fmt.Sprintf("Superlotação: População %d acima do teto %d", r.Population.Level, r.Population.Max)})
	}
	if g := guardless(r.Buildings); g > 0 {
		w.rep.Notes = append(w.rep.Notes, fmt.Sprintf("Fortificação sem guarda: %d de Defesa não conta", g))
	}
	if def := EffectiveDefense(*r); def < r.Region.Danger {
		w.rep.Triggers = append(w.rep.Triggers, CrisisTrigger{Kind: "breach",
			Detail: fmt.Sprintf("Defesa %d abaixo do Perigo %d da região", def, r.Region.Danger)})
	}
}
