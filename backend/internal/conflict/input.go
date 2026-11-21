package conflict

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// Input is the editable part sent by the Mestre; progress and counters are server-owned.
type Input struct {
	SceneID     *string      `json:"sceneId"`
	Title       string       `json:"title"`
	Conditions  []string     `json:"conditions"`
	Objectives  []Objective  `json:"objectives"`
	Threats     []Threat     `json:"threats"`
	Activations []Activation `json:"activations"`
}

var idRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)

func bad(msg string) error { return realtime.UserError(msg) }

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func fixID(id string) (string, error) {
	if id == "" {
		return newID(), nil
	}
	if !idRe.MatchString(id) {
		return "", bad("Identificador inválido.")
	}
	return id, nil
}

// normalize validates the input against the previous doc and keeps progress/counters by id.
func (s *Service) normalize(ctx context.Context, m campaigns.Member, in Input, prev Doc) (Doc, string, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 120 {
		return Doc{}, "", bad("Informe um título de até 120 caracteres.")
	}
	var d Doc
	for _, c := range in.Conditions {
		if c = strings.TrimSpace(c); c != "" && len(d.Conditions) < 50 {
			d.Conditions = append(d.Conditions, c)
		}
	}
	var err error
	if d.Objectives, err = objectives(in.Objectives, prev.Objectives); err != nil {
		return Doc{}, "", err
	}
	if d.Threats, err = s.threats(ctx, m, in.Threats); err != nil {
		return Doc{}, "", err
	}
	if d.Activations, err = activations(in.Activations, prev.Activations, d.Threats); err != nil {
		return Doc{}, "", err
	}
	return d, title, nil
}

func objectives(in, prev []Objective) ([]Objective, error) {
	if len(in) > 30 {
		return nil, bad("Objetivos demais (máximo 30).")
	}
	old := map[string]Objective{}
	for _, o := range prev {
		old[o.ID] = o
	}
	out := make([]Objective, 0, len(in))
	for _, o := range in {
		id, err := checkObjective(&o)
		if err != nil {
			return nil, err
		}
		o.ID, o.Progress = id, old[id].Progress
		out = append(out, o)
	}
	return out, nil
}

func checkObjective(o *Objective) (string, error) {
	id, err := fixID(o.ID)
	o.Title = strings.TrimSpace(o.Title)
	switch {
	case err != nil:
		return "", err
	case o.Title == "" || len(o.Title) > 120:
		return "", bad("Todo objetivo precisa de um título.")
	case o.Kind != KindMain && o.Kind != KindSecondary, o.Scope != ScopeCollective && o.Scope != ScopeIndividual:
		return "", bad("Tipo ou alcance de objetivo inválido.")
	case o.Target < 1 || o.Target > 999:
		return "", bad("A meta de Sucessos deve estar entre 1 e 999.")
	}
	return id, nil
}

func checkPool(t *Threat) error {
	if t.Pool == nil {
		t.Pool = &rolls.ThreatPool{}
	}
	p := *t.Pool
	if p.D6 < 0 || p.D10 < 0 || p.D12 < 0 || p.D6 > 20 || p.D10 > 20 || p.D12 > 20 || p.D6+p.D10+p.D12 < 1 {
		return bad("A parada da ameaça deve ter entre 1 e 20 dados de cada tipo.")
	}
	return nil
}

func (s *Service) threats(ctx context.Context, m campaigns.Member, in []Threat) ([]Threat, error) {
	if len(in) > 30 {
		return nil, bad("Ameaças demais (máximo 30).")
	}
	out := make([]Threat, 0, len(in))
	for _, t := range in {
		id, err := fixID(t.ID)
		t.Name = strings.TrimSpace(t.Name)
		if err != nil {
			return nil, err
		}
		if t.Name == "" || len(t.Name) > 80 {
			return nil, bad("Toda ameaça precisa de um nome.")
		}
		if err := checkPool(&t); err != nil {
			return nil, err
		}
		if t.NpcCharacterID != nil && *t.NpcCharacterID == "" {
			t.NpcCharacterID = nil
		}
		if err := s.checkNPC(ctx, m, t.NpcCharacterID); err != nil {
			return nil, err
		}
		t.ID = id
		out = append(out, t)
	}
	return out, nil
}

func checkActivation(a Activation, known map[string]bool) (string, error) {
	id, err := fixID(a.ID)
	switch {
	case err != nil:
		return "", err
	case a.Symbol != "success" && a.Symbol != "adaptation" && a.Symbol != "pressure":
		return "", bad("Símbolo de ativação inválido.")
	case a.Threshold < 1 || a.Threshold > 99:
		return "", bad("O limiar da ativação deve estar entre 1 e 99.")
	case !known[a.Source]:
		return "", bad("Origem da ativação inválida.")
	}
	return id, nil
}

func activations(in, prev []Activation, threats []Threat) ([]Activation, error) {
	if len(in) > 30 {
		return nil, bad("Ativações demais (máximo 30).")
	}
	old := map[string]Activation{}
	for _, a := range prev {
		old[a.ID] = a
	}
	known := map[string]bool{SourcePlayers: true}
	for _, t := range threats {
		known[SourceThreat+t.ID] = true
	}
	out := make([]Activation, 0, len(in))
	for _, a := range in {
		id, err := checkActivation(a, known)
		if err != nil {
			return nil, err
		}
		prevAct := old[id]
		a.ID, a.Text = id, strings.TrimSpace(a.Text)
		a.Count = 0
		if prevAct.Symbol == a.Symbol && prevAct.Source == a.Source {
			a.Count = prevAct.Count
		}
		out = append(out, a)
	}
	return out, nil
}

// tally adds symbols to the activations fed by source and reports the ones that fired.
func tally(acts []Activation, source string, sym assimilacao.Symbols) []Triggered {
	var fired []Triggered
	for i := range acts {
		a := &acts[i]
		if a.Source != source {
			continue
		}
		a.Count += map[string]int{"success": sym.Success, "adaptation": sym.Adaptation, "pressure": sym.Pressure}[a.Symbol]
		if times := a.Count / a.Threshold; times > 0 {
			a.Count %= a.Threshold
			fired = append(fired, Triggered{ActivationID: a.ID, Symbol: a.Symbol, Threshold: a.Threshold, Text: a.Text, Source: source, Times: times})
		}
	}
	return fired
}

// Triggered is the payload of conflict.activation.triggered.
type Triggered struct {
	ActivationID string `json:"activationId"`
	Symbol       string `json:"symbol"`
	Threshold    int    `json:"threshold"`
	Text         string `json:"text"`
	Source       string `json:"source"`
	Times        int    `json:"times"`
}
