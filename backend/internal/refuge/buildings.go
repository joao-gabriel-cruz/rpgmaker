package refuge

import (
	"crypto/rand"
	"encoding/hex"
	"slices"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Building states.
const (
	StatePlanned  = "planned"
	StateBuilding = "building"
	StateDone     = "done"
	maxProposals  = 20
	maxBuildings  = 40
)

// Building is one construction: Level is the target level, Built the levels already finished.
// Workers are the População allocated: they build while the state is building and operate once done.
type Building struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Level    int    `json:"level"`
	Built    int    `json:"built"`
	State    string `json:"state"`
	Progress int    `json:"progress"`
	Paid     bool   `json:"paid"`
	Workers  int    `json:"workers"`
}

// Proposal is a building suggested by a player and waiting for the Mestre.
type Proposal struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	UserID string `json:"userId"`
	By     string `json:"by"`
}

var (
	errBadKind   = realtime.UserError("Construção desconhecida")
	errHasKind   = realtime.UserError("Essa construção já existe ou já foi proposta")
	errNoBuild   = realtime.UserError("Construção não encontrada")
	errNoProp    = realtime.UserError("Proposta não encontrada")
	errTooMany   = realtime.UserError("Há construções ou propostas demais")
	errNotLevel  = realtime.UserError("Essa construção não tem níveis")
	errNotDone   = realtime.UserError("Conclua o nível atual antes de evoluir")
	errCancel    = realtime.UserError("Só dá para cancelar obras que ainda não foram concluídas")
	errAllocNeg  = realtime.UserError("A alocação não pode ser negativa")
	errAllocOver = realtime.UserError("Não há trabalhadores suficientes para essa alocação")
)

func newID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (r *Refuge) findBuilding(id string) int {
	return slices.IndexFunc(r.Buildings, func(b Building) bool { return b.ID == id })
}

func (r *Refuge) hasKind(kind string) bool {
	return slices.ContainsFunc(r.Buildings, func(b Building) bool { return b.Kind == kind }) ||
		slices.ContainsFunc(r.Proposals, func(p Proposal) bool { return p.Kind == kind })
}

// plan adds a planned building of kind at level 1.
func (r *Refuge) plan(kind string) error {
	if _, ok := DefOf(kind); !ok {
		return errBadKind
	}
	if r.hasKind(kind) {
		return errHasKind
	}
	if len(r.Buildings) >= maxBuildings {
		return errTooMany
	}
	r.Buildings = append(r.Buildings, Building{ID: newID(), Kind: kind, Level: 1, State: StatePlanned})
	return nil
}

func (r *Refuge) propose(kind, userID, by string) error {
	if _, ok := DefOf(kind); !ok {
		return errBadKind
	}
	if r.hasKind(kind) {
		return errHasKind
	}
	if len(r.Proposals) >= maxProposals {
		return errTooMany
	}
	r.Proposals = append(r.Proposals, Proposal{ID: newID(), Kind: kind, UserID: userID, By: by})
	return nil
}

// resolve removes a proposal and, when approved, turns it into a planned building.
func (r *Refuge) resolve(id string, approve bool) error {
	i := slices.IndexFunc(r.Proposals, func(p Proposal) bool { return p.ID == id })
	if i < 0 {
		return errNoProp
	}
	kind := r.Proposals[i].Kind
	r.Proposals = slices.Delete(r.Proposals, i, i+1)
	if approve {
		return r.plan(kind)
	}
	return nil
}

// upgrade starts the next level of a finished leveled building.
func (r *Refuge) upgrade(id string) error {
	i := r.findBuilding(id)
	if i < 0 {
		return errNoBuild
	}
	b := &r.Buildings[i]
	if d, _ := DefOf(b.Kind); !d.Leveled {
		return errNotLevel
	}
	if b.State != StateDone {
		return errNotDone
	}
	b.Level++
	b.State, b.Progress, b.Paid = StatePlanned, 0, false
	return nil
}

// cancel drops a building that never produced anything (nothing was built yet).
func (r *Refuge) cancel(id string) error {
	i := r.findBuilding(id)
	if i < 0 {
		return errNoBuild
	}
	if r.Buildings[i].Built > 0 {
		return errCancel
	}
	r.Buildings = slices.Delete(r.Buildings, i, i+1)
	return nil
}

// allocate replaces every building's workers; the total cannot pass the available workers.
// Workers sent to a planned building start its construction.
func (r *Refuge) allocate(alloc map[string]int) error {
	total := 0
	for id, n := range alloc {
		if n < 0 {
			return errAllocNeg
		}
		if r.findBuilding(id) < 0 {
			return errNoBuild
		}
		total += n
	}
	if total > Workers(r.Population.Level) {
		return errAllocOver
	}
	for i := range r.Buildings {
		b := &r.Buildings[i]
		b.Workers = alloc[b.ID]
		if b.Workers > 0 && b.State == StatePlanned {
			b.State = StateBuilding
		}
	}
	return nil
}
