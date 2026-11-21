package rolls

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// EventDifficulty tells the Mestre and the character's owner the pending difficulty.
const EventDifficulty = "roll.difficulty.updated"

// Usage tells which once-per-session modifiers a character already used in a session.
type Usage struct {
	OccupationUsed   bool `json:"occupationUsed"`
	MarkingEventUsed bool `json:"markingEventUsed"`
}

// State is what the roll panel needs about a character.
type State struct {
	Usage
	SessionActive bool `json:"sessionActive"`
	DifficultyD12 int  `json:"difficultyD12"`
}

func requestOf(in RequestInput) assimilacao.RollRequest {
	return assimilacao.RollRequest{
		Instinct: in.Instinct, Aptitude: in.Aptitude, SecondInstinct: in.SecondInstinct, Effort: in.Effort,
		UseOccupation: in.UseOccupation, UseMarkingEvent: in.UseMarkingEvent, DifficultyD12: in.DifficultyD12,
	}
}

func nullID(id string) *string {
	if id == "" {
		return nil
	}
	return &id
}

// prepare checks session-bound modifiers and resolves the difficulty; it returns the active session id.
func (s *Service) prepare(ctx context.Context, campaignID string, sender realtime.Sender, characterID string, req *assimilacao.RollRequest) (string, error) {
	if req.DifficultyD12 != 0 && sender.Role != campaigns.RoleGM {
		return "", realtime.UserError("Apenas o Mestre pode definir a dificuldade.")
	}
	sessionID, err := s.sess.ActiveID(ctx, campaignID)
	if err != nil {
		return "", err
	}
	if err := s.checkUsage(ctx, sessionID, characterID, *req); err != nil {
		return "", err
	}
	if req.DifficultyD12 == 0 {
		req.DifficultyD12, err = s.repo.Difficulty(ctx, campaignID, characterID)
	}
	return sessionID, err
}

func (s *Service) checkUsage(ctx context.Context, sessionID, characterID string, req assimilacao.RollRequest) error {
	if !req.UseOccupation && !req.UseMarkingEvent {
		return nil
	}
	if sessionID == "" {
		return realtime.UserError("Inicie uma sessão para usar Ocupação ou Evento Marcante.")
	}
	used, err := s.repo.Usage(ctx, sessionID, characterID)
	if err != nil {
		return err
	}
	if req.UseOccupation && used.OccupationUsed {
		return realtime.UserError("A Ocupação já foi usada nesta sessão.")
	}
	if req.UseMarkingEvent && used.MarkingEventUsed {
		return realtime.UserError("O Evento Marcante já foi usado nesta sessão.")
	}
	return nil
}

// commitCosts consumes the pending difficulty and charges Empenho on the sheet.
func (s *Service) commitCosts(ctx context.Context, campaignID string, sender realtime.Sender, view characters.View, sheet assimilacao.Sheet, req assimilacao.RollRequest) error {
	if err := s.repo.SetDifficulty(ctx, campaignID, view.ID, 0); err != nil {
		return err
	}
	if !req.Effort {
		return nil
	}
	points := assimilacao.ApplyCosts(sheet, req).TugOfWar.DeterminationPoints
	raw, err := withDetermination(view.Sheet, points)
	if err != nil {
		return err
	}
	_, err = s.chars.Save(ctx, member(campaignID, sender), view.ID, raw)
	return err
}

// withDetermination patches only tugOfWar.determinationPoints so other sheet fields stay untouched.
func withDetermination(sheet json.RawMessage, points int) (json.RawMessage, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(sheet, &doc); err != nil {
		return nil, err
	}
	tug := map[string]json.RawMessage{}
	if raw, ok := doc["tugOfWar"]; ok {
		if err := json.Unmarshal(raw, &tug); err != nil {
			return nil, err
		}
	}
	tug["determinationPoints"], _ = json.Marshal(points)
	doc["tugOfWar"], _ = json.Marshal(tug)
	return json.Marshal(doc)
}

// SetDifficulty stores the d12 (0-3) added to the character's next roll and tells the Mestre and the owner.
func (s *Service) SetDifficulty(ctx context.Context, campaignID string, sender realtime.Sender, characterID string, d12 int) error {
	if d12 < 0 || d12 > 3 {
		return realtime.UserError("A dificuldade deve ficar entre 0 e 3 d12.")
	}
	view, err := s.chars.Get(ctx, member(campaignID, sender), characterID)
	if err != nil {
		return err
	}
	if err := s.repo.SetDifficulty(ctx, campaignID, view.ID, d12); err != nil {
		return err
	}
	to := func(r realtime.Sender) bool {
		return r.Role == campaigns.RoleGM || (view.OwnerID != nil && r.UserID == *view.OwnerID)
	}
	s.events.Broadcast(campaignID, to, EventDifficulty, map[string]any{"characterId": view.ID, "d12": d12})
	return nil
}

// State reports the modifiers a character can still use; the pending difficulty is only revealed to the Mestre and the owner.
func (s *Service) State(ctx context.Context, m campaigns.Member, characterID string) (State, error) {
	view, err := s.chars.Get(ctx, m, characterID)
	if err != nil {
		return State{}, err
	}
	sessionID, err := s.sess.ActiveID(ctx, m.CampaignID)
	if err != nil {
		return State{}, err
	}
	st := State{SessionActive: sessionID != ""}
	if sessionID != "" {
		if st.Usage, err = s.repo.Usage(ctx, sessionID, view.ID); err != nil {
			return State{}, err
		}
	}
	if m.Role != campaigns.RoleGM && (view.OwnerID == nil || *view.OwnerID != m.UserID) {
		return st, nil
	}
	st.DifficultyD12, err = s.repo.Difficulty(ctx, m.CampaignID, view.ID)
	return st, err
}
