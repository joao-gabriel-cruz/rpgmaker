package refuge

import (
	"fmt"
	"slices"
)

// runWorks adds the week's Pontos de Obra to every construction with workers.
func (w *weekSim) runWorks() {
	for i := range w.r.Buildings {
		b := &w.r.Buildings[i]
		if b.State != StateBuilding || b.Workers == 0 {
			continue
		}
		d, _ := DefOf(b.Kind)
		line := WorkLine{BuildingID: b.ID, Name: d.Name, Level: b.Level, Cost: d.CostAt(b.Level)}
		if !b.Paid && !w.payMaterials(d.MaterialCost(b.Level)) {
			line.Progress, line.Blocked = b.Progress, "Falta Material de Construção (ou Madeira)"
			w.rep.Works = append(w.rep.Works, line)
			continue
		}
		b.Paid = true
		line.Points = min(b.Workers, line.Cost-b.Progress)
		b.Progress += line.Points
		line.Progress = b.Progress
		if b.Progress >= line.Cost {
			w.complete(b, d, &line)
		}
		w.rep.Works = append(w.rep.Works, line)
	}
}

func (w *weekSim) complete(b *Building, d Def, line *WorkLine) {
	applyEffect(&w.r.Sheet, b.Kind, b.Level-b.Built)
	b.Built, b.State, b.Paid = b.Level, StateDone, false
	line.Done = true
	name := d.Name
	if d.Leveled {
		name = fmt.Sprintf("%s (nível %d)", d.Name, b.Level)
	}
	w.rep.Completed = append(w.rep.Completed, name)
}

// payMaterials pays n Material de Construção, using 2 Madeira per missing unit; all or nothing.
func (w *weekSim) payMaterials(n int) bool {
	have := w.r.Resources["buildingMaterial"]
	useMat := min(n, have)
	wood := (n - useMat) * WoodPerMaterial
	if w.r.Resources["wood"] < wood {
		return false
	}
	if useMat > 0 {
		w.spend("buildingMaterial", useMat)
	}
	if wood > 0 {
		w.spend("wood", wood)
	}
	return true
}

// produce runs extraction, then conversions, then the doubling buildings (Refeitório, Casa de Costura).
func (w *weekSim) produce(active []Building) {
	idle := w.r.Derive().Workers
	for _, b := range w.r.Buildings {
		idle -= b.Workers
	}
	for _, b := range active {
		if d, _ := DefOf(b.Kind); d.Extract != "" {
			w.extract(b, d)
		}
	}
	for _, b := range active {
		if d, _ := DefOf(b.Kind); d.Input != "" {
			w.convert(b, d)
		}
	}
	for _, b := range active {
		if d, _ := DefOf(b.Kind); d.Doubles != "" {
			w.gain(d.Doubles, w.rep.Entered[d.Doubles])
		}
	}
	if idle > 0 {
		w.rep.Notes = append(w.rep.Notes, fmt.Sprintf("%d trabalhador(es) sem destino", idle))
	}
}

func (w *weekSim) extract(b Building, d Def) {
	n := b.Workers * abundanceHalves[w.r.Region.Abundance[d.Extract]] / 2
	if b.Workers > 0 && n == 0 {
		w.rep.Notes = append(w.rep.Notes, fmt.Sprintf("%s: a região não rende %s", d.Name, ResourceLabel[d.Extract]))
	}
	w.gain(d.Extract, n)
}

func (w *weekSim) convert(b Building, d Def) {
	n := min(b.Workers, w.r.Resources[d.Input])
	if d.Cap > 0 {
		n = min(n, d.Cap)
	}
	if n <= 0 {
		return
	}
	w.spend(d.Input, n)
	w.gain(d.Output, n)
}

// needs lists what the population must be fed this week (plus yearly/biennial/port upkeep).
func (w *weekSim) needs(active []Building) map[string]int {
	level, week := w.r.Population.Level, w.r.Week
	need := map[string]int{"food": level, "wood": level}
	if !w.r.HasWaterSource {
		need["water"] = level
	}
	if week%YearWeeks == 0 {
		need["medicine"] = level
		w.rep.Notes = append(w.rep.Notes, "Marco de 52 semanas (ano): 1 Remédio por nível de População")
	}
	if week%BiennialWeek == 0 {
		need["clothing"] = level
		w.rep.Notes = append(w.rep.Notes, "Marco de 104 semanas: 1 Vestuário por nível de População")
	}
	if week%PortUpkeepWeeks == 0 && slices.ContainsFunc(active, func(b Building) bool { return b.Kind == KindPort }) {
		need["wood"] += PortUpkeepWood
		w.rep.Notes = append(w.rep.Notes, fmt.Sprintf("Manutenção semestral do Porto: %d Madeira", PortUpkeepWood))
	}
	return need
}

// consume charges the needs; short stock becomes a shortage trigger.
func (w *weekSim) consume(active []Building) {
	need := w.needs(active)
	for _, k := range ResourceKeys {
		n, have := need[k], w.r.Resources[k]
		if n == 0 {
			continue
		}
		w.spend(k, min(n, have))
		if have < n {
			w.rep.Missing[k] = n - have
			w.rep.Triggers = append(w.rep.Triggers, CrisisTrigger{Kind: "shortage", Resource: k,
				Detail: fmt.Sprintf("Faltou %s (precisava %d, tinha %d)", ResourceLabel[k], n, have)})
		}
	}
}
