package refuge

import (
	"context"
	"errors"
	"slices"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
)

const (
	maxResidents = 50
	maxNote      = 500
)

var residentRoles = []string{"guard", "collector", "medic", "artisan", "leader", "other"}

// Resident is a campaign NPC living in the Refúgio with a role; it never changes the Population math.
type Resident struct {
	ID          string `json:"id"`
	CharacterID string `json:"characterId"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Note        string `json:"note"`
}

// ResidentInput adds or edits a resident.
type ResidentInput struct {
	ID          string `json:"id"`
	CharacterID string `json:"characterId"`
	Role        string `json:"role"`
	Note        string `json:"note"`
}

func validResident(role, note string) error {
	if !slices.Contains(residentRoles, role) || len(note) > maxNote {
		return bad("Função inválida ou nota longa demais (até 500)")
	}
	return nil
}

// AddResident marks a campaign NPC as a resident of the Refúgio.
func (s *Service) AddResident(ctx context.Context, campaignID, userID string, in ResidentInput) error {
	if err := validResident(in.Role, in.Note); err != nil {
		return err
	}
	if s.links.Chars == nil {
		return bad("Personagens indisponíveis")
	}
	view, err := s.links.Chars.Get(ctx, campaigns.Member{CampaignID: campaignID, UserID: userID, Role: campaigns.RoleGM}, in.CharacterID)
	if errors.Is(err, characters.ErrNotFound) || (err == nil && view.Kind != characters.KindNPC) {
		return bad("Escolha um NPC da campanha")
	}
	if err != nil {
		return err
	}
	_, err = s.Mutate(ctx, campaignID, func(r *Refuge) error {
		if len(r.Residents) >= maxResidents || slices.ContainsFunc(r.Residents, func(x Resident) bool { return x.CharacterID == view.ID }) {
			return bad("Limite de moradores atingido ou NPC já é morador")
		}
		r.Residents = append(r.Residents, Resident{ID: newID(), CharacterID: view.ID, Name: view.Name, Role: in.Role, Note: in.Note})
		return nil
	})
	return err
}

// UpdateResident changes a resident's role and note.
func (s *Service) UpdateResident(ctx context.Context, campaignID string, in ResidentInput) error {
	if err := validResident(in.Role, in.Note); err != nil {
		return err
	}
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		i := slices.IndexFunc(r.Residents, func(x Resident) bool { return x.ID == in.ID })
		if i < 0 {
			return bad("Morador não encontrado")
		}
		r.Residents[i].Role, r.Residents[i].Note = in.Role, in.Note
		return nil
	})
	return err
}

// RemoveResident unmarks a resident.
func (s *Service) RemoveResident(ctx context.Context, campaignID, id string) error {
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		n := len(r.Residents)
		r.Residents = slices.DeleteFunc(r.Residents, func(x Resident) bool { return x.ID == id })
		if len(r.Residents) == n {
			return bad("Morador não encontrado")
		}
		return nil
	})
	return err
}
