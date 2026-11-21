package rolls

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// KindThreat is a Mestre roll of a conflict threat pool; every die counts.
const KindThreat = "threat"

// ThreatPool is the dice a threat rolls.
type ThreatPool struct {
	D6  int `json:"d6"`
	D10 int `json:"d10"`
	D12 int `json:"d12"`
}

type threatSpec struct {
	Name string     `json:"name"`
	Pool ThreatPool `json:"pool"`
}

// Final is the settled outcome of a roll, as needed by other domains.
type Final struct {
	UserID  string
	Kind    string
	Symbols assimilacao.Symbols
}

// Final returns the owner and symbols of a finalized roll; ErrNotFound/ErrNotFinal otherwise.
func (s *Service) Final(ctx context.Context, campaignID, rollID string) (Final, error) {
	rec, err := s.repo.Get(ctx, campaignID, rollID)
	if err != nil {
		return Final{}, err
	}
	if rec.Status != StatusFinal || rec.Kind == KindGeneric {
		return Final{}, ErrNotFinal
	}
	var res struct {
		Symbols assimilacao.Symbols `json:"symbols"`
	}
	if err := json.Unmarshal(rec.Result, &res); err != nil {
		return Final{}, err
	}
	return Final{UserID: rec.UserID, Kind: rec.Kind, Symbols: res.Symbols}, nil
}

// Threat rolls a threat pool for the Mestre, logs it publicly and returns the symbols drawn.
func (s *Service) Threat(ctx context.Context, campaignID string, sender realtime.Sender, name string, pool ThreatPool) (assimilacao.Symbols, error) {
	var spec assimilacao.RollSpec
	for _, p := range []struct {
		kind string
		n    int
	}{{"d6", pool.D6}, {"d10", pool.D10}, {"d12", pool.D12}} {
		for range p.n {
			spec.Dice = append(spec.Dice, assimilacao.DieSpec{Kind: p.kind})
		}
	}
	roll := assimilacao.Resolve(spec, s.rng)
	var total assimilacao.Symbols
	for _, d := range roll.Dice {
		total = total.Add(assimilacao.FaceSymbols(d.Kind, d.Face))
	}
	result, _ := json.Marshal(assimilacao.Result{Symbols: total, Kept: []int{}})
	err := s.store(ctx, campaignID, sender, Record{Visibility: VisPublic, Kind: KindThreat, Status: StatusFinal, Result: result},
		threatSpec{Name: name, Pool: pool}, roll.Dice, nil)
	return total, err
}
