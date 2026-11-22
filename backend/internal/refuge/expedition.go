package refuge

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// Expedition statuses.
const (
	ExpProposed = "proposed"
	ExpReady    = "ready"
	ExpRolling  = "rolling"
	ExpDone     = "done"
)

// EventExpeditionResult is announced to everyone when an expedition's roll settles.
const EventExpeditionResult = "refuge.expedition.result"

const (
	maxExpeditions = 30
	maxParty       = 6
	maxRegionNote  = 300
)

// abundanceFactor is the units per Success by regional abundance [casa]; non-natural resources count as scarce.
var abundanceFactor = map[string]int{"none": 0, "scarce": 1, "medium": 2, "abundant": 3}

// Expedition is a group supply run: who goes, what for, and what came back.
type Expedition struct {
	ID            string   `json:"id"`
	Status        string   `json:"status"`
	CharacterIDs  []string `json:"characterIds"`
	LeaderName    string   `json:"leaderName"`
	Resource      string   `json:"resource"`
	RegionNote    string   `json:"regionNote"`
	ProposedBy    string   `json:"proposedBy"`
	Week          int      `json:"week"`
	RollID        string   `json:"rollId,omitempty"`
	Factor        int      `json:"factor"`
	Successes     int      `json:"successes"`
	Adaptations   int      `json:"adaptations"`
	Pressures     int      `json:"pressures"`
	Found         int      `json:"found"`
	Discarded     int      `json:"discarded"`
	Extras        int      `json:"extras"`
	Complications int      `json:"complications"`
	Log           []string `json:"log"`
}

// ExpeditionInput proposes or starts an expedition; the first character leads.
type ExpeditionInput struct {
	CharacterIDs []string `json:"characterIds"`
	Resource     string   `json:"resource"`
	RegionNote   string   `json:"regionNote"`
}

func findExpedition(r *Refuge, id string) (*Expedition, error) {
	for i := range r.Expeditions {
		if r.Expeditions[i].ID == id {
			return &r.Expeditions[i], nil
		}
	}
	return nil, bad("Expedição não encontrada")
}

func (s *Service) character(ctx context.Context, campaignID, id string) (characters.View, error) {
	if s.links.Chars == nil {
		return characters.View{}, bad("Personagens indisponíveis")
	}
	v, err := s.links.Chars.Get(ctx, campaigns.Member{CampaignID: campaignID, Role: campaigns.RoleGM}, id)
	if err != nil {
		return v, bad("Personagem não encontrado")
	}
	return v, nil
}

// StartExpedition starts (Mestre) or proposes (player) a supply run.
func (s *Service) StartExpedition(ctx context.Context, campaignID string, sender realtime.Sender, in ExpeditionInput) error {
	n := len(in.CharacterIDs)
	if n < 1 || n > maxParty || !IsResource(in.Resource) || len(in.RegionNote) > maxRegionNote {
		return bad("Escolha de 1 a 6 personagens e um recurso válido")
	}
	for i, id := range in.CharacterIDs {
		if slices.Index(in.CharacterIDs, id) != i {
			return bad("Personagem repetido na expedição")
		}
	}
	leader, err := s.character(ctx, campaignID, in.CharacterIDs[0])
	if err != nil {
		return err
	}
	isGM := sender.Role == campaigns.RoleGM
	if !isGM && (leader.OwnerID == nil || *leader.OwnerID != sender.UserID) {
		return bad("Você só pode propor expedições lideradas por um personagem seu")
	}
	for _, id := range in.CharacterIDs[1:] {
		if _, err := s.character(ctx, campaignID, id); err != nil {
			return err
		}
	}
	_, err = s.Mutate(ctx, campaignID, func(r *Refuge) error { return addExpedition(r, in, leader.Name, sender.UserID, isGM) })
	return err
}

func addExpedition(r *Refuge, in ExpeditionInput, leaderName, userID string, isGM bool) error {
	if len(r.Expeditions) >= maxExpeditions {
		i := slices.IndexFunc(r.Expeditions, func(e Expedition) bool { return e.Status == ExpDone })
		if i < 0 {
			return bad("Há expedições demais em andamento")
		}
		r.Expeditions = slices.Delete(r.Expeditions, i, i+1)
	}
	status := ExpProposed
	if isGM {
		status = ExpReady
	}
	r.Expeditions = append(r.Expeditions, Expedition{ID: newID(), Status: status, CharacterIDs: in.CharacterIDs, LeaderName: leaderName,
		Resource: in.Resource, RegionNote: in.RegionNote, ProposedBy: userID, Week: r.Week, Log: []string{}})
	return nil
}

// ApproveExpedition accepts a player's proposal.
func (s *Service) ApproveExpedition(ctx context.Context, campaignID, id string) error {
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		e, err := findExpedition(r, id)
		if err == nil && e.Status != ExpProposed {
			return bad("Esta expedição não está esperando aprovação")
		}
		if err == nil {
			e.Status = ExpReady
		}
		return err
	})
	return err
}

// CancelExpedition drops an expedition that has not rolled yet (the Mestre, or the proposer).
func (s *Service) CancelExpedition(ctx context.Context, campaignID string, sender realtime.Sender, id string) error {
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		e, err := findExpedition(r, id)
		if err != nil {
			return err
		}
		if e.Status == ExpRolling || (sender.Role != campaigns.RoleGM && e.ProposedBy != sender.UserID) {
			return bad("Não é possível cancelar esta expedição")
		}
		r.Expeditions = slices.DeleteFunc(r.Expeditions, func(x Expedition) bool { return x.ID == id })
		return nil
	})
	return err
}

// authorizeRoll finds the ready expedition and checks that sender may roll it.
func (s *Service) authorizeRoll(ctx context.Context, campaignID string, sender realtime.Sender, id, instinct string) (Expedition, characters.View, error) {
	r, err := s.Get(ctx, campaignID)
	if err != nil {
		return Expedition{}, characters.View{}, err
	}
	e, err := findExpedition(&r, id)
	if err != nil {
		return Expedition{}, characters.View{}, err
	}
	leader, err := s.character(ctx, campaignID, e.CharacterIDs[0])
	if err != nil {
		return *e, leader, err
	}
	if e.Status != ExpReady || (instinct != "perception" && instinct != "sagacity") ||
		(sender.Role != campaigns.RoleGM && (leader.OwnerID == nil || *leader.OwnerID != sender.UserID)) {
		return *e, leader, bad("Apenas o líder (ou o Mestre) rola, com Percepção ou Sagacidade, quando a expedição está pronta")
	}
	return *e, leader, nil
}

// RollExpedition rolls the leader's Percepção/Sagacidade + Sobrevivência; every extra member adds a kept die (up to 2).
func (s *Service) RollExpedition(ctx context.Context, campaignID string, sender realtime.Sender, id, instinct string) error {
	e, leader, err := s.authorizeRoll(ctx, campaignID, sender, id, instinct)
	if err != nil {
		return err
	}
	entry, err := s.links.Rolls.RequestEntry(ctx, campaignID, sender, rolls.RequestInput{CharacterID: leader.ID, Instinct: instinct,
		Aptitude: "survival", BonusKept: min(len(e.CharacterIDs)-1, 2)})
	if err != nil {
		return rollErr(err)
	}
	_, err = s.Mutate(ctx, campaignID, func(r *Refuge) error {
		x, err := findExpedition(r, id)
		if err == nil {
			x.Status, x.RollID = ExpRolling, entry.RollID
		}
		return err
	})
	return err
}

// ExpeditionFinalized is the rolls.OnFinalized hook that settles the expedition tied to the roll.
func (s *Service) ExpeditionFinalized(ctx context.Context, campaignID, rollID string) {
	r, err := s.Get(ctx, campaignID)
	if err != nil || !slices.ContainsFunc(r.Expeditions, func(e Expedition) bool { return e.Status == ExpRolling && e.RollID == rollID }) {
		return
	}
	fin, err := s.links.Rolls.Final(ctx, campaignID, rollID)
	if err != nil {
		slog.Error("refuge: read expedition roll failed", "err", err)
		return
	}
	var done Expedition
	_, err = s.Mutate(ctx, campaignID, func(r *Refuge) error {
		i := slices.IndexFunc(r.Expeditions, func(e Expedition) bool { return e.Status == ExpRolling && e.RollID == rollID })
		if i < 0 {
			return bad("Expedição já resolvida")
		}
		settle(r, &r.Expeditions[i], fin.Symbols)
		done = r.Expeditions[i]
		return nil
	})
	if err != nil {
		slog.Error("refuge: settle expedition failed", "err", err)
		return
	}
	s.events.Broadcast(campaignID, realtime.ToAll, EventExpeditionResult, map[string]Expedition{"expedition": done})
}

func settle(r *Refuge, e *Expedition, sym assimilacao.Symbols) {
	factor, ok := abundanceFactor[r.Region.Abundance[e.Resource]]
	if !ok || !slices.Contains(NaturalKeys, e.Resource) {
		factor = 1
	}
	e.Factor, e.Successes, e.Adaptations, e.Pressures = factor, sym.Success, sym.Adaptation, sym.Pressure
	e.Found = sym.Success * factor
	e.Discarded = r.AddResource(e.Resource, e.Found)
	e.Extras, e.Complications, e.Status = sym.Adaptation, sym.Pressure, ExpDone
	e.Log = append(e.Log, summary(e))
}

func summary(e *Expedition) string {
	msg := fmt.Sprintf("%d Sucesso(s) x %d = %d de %s nas Reservas", e.Successes, e.Factor, e.Found-e.Discarded, ResourceLabels[e.Resource])
	if e.Discarded > 0 {
		msg += fmt.Sprintf(" (%d descartado(s): Reservas cheias)", e.Discarded)
	}
	return msg
}
