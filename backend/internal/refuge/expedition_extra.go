package refuge

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// ExtraInput spends one Adaptation: +1 of another resource or a compendium item.
type ExtraInput struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
	ItemID   string `json:"itemId"`
}

// ComplicationInput spends one Pressure: injury (damage suggestion), encounter (note) or loss (of a resource).
type ComplicationInput struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	CharacterID string `json:"characterId"`
	Resource    string `json:"resource"`
	Amount      int    `json:"amount"`
	Note        string `json:"note"`
}

const maxHarm = 5

func rollErr(err error) error {
	var a assimilacao.RuleError
	var g rolls.RuleError
	if errors.As(err, &a) || errors.As(err, &g) {
		return bad(err.Error())
	}
	return err
}

// compendiumItem reads the chosen item (nil when the extra is a resource).
func (s *Service) compendiumItem(ctx context.Context, campaignID string, in ExtraInput) ([]byte, error) {
	if in.ItemID == "" {
		if !IsResource(in.Resource) {
			return nil, bad("Escolha um recurso ou um item")
		}
		return nil, nil
	}
	if s.links.Compendium == nil {
		return nil, bad("Compêndio indisponível")
	}
	raw, err := s.links.Compendium(ctx, campaignID, in.ItemID)
	if err != nil {
		return nil, bad("Item não encontrado no compêndio")
	}
	return raw, nil
}

// ExpeditionExtra delivers the bonus of one Adaptation chosen by the Mestre.
func (s *Service) ExpeditionExtra(ctx context.Context, campaignID string, in ExtraInput) error {
	item, err := s.compendiumItem(ctx, campaignID, in)
	if err != nil {
		return err
	}
	_, err = s.Mutate(ctx, campaignID, func(r *Refuge) error {
		e, err := findExpedition(r, in.ID)
		if err != nil {
			return err
		}
		if e.Status != ExpDone || e.Extras < 1 {
			return bad("Não há Adaptações sobrando nesta expedição")
		}
		line, err := deliverExtra(r, in.Resource, item)
		if err != nil {
			return err
		}
		e.Log = append(e.Log, line)
		e.Extras--
		return nil
	})
	return err
}

func deliverExtra(r *Refuge, resource string, item []byte) (string, error) {
	if item != nil {
		if len(r.Items) >= maxItems {
			return "", bad("O inventário do Refúgio está cheio")
		}
		r.Items = append(r.Items, item)
		return "Adaptação: um item do compêndio foi achado", nil
	}
	if r.AddResource(resource, 1) > 0 {
		return fmt.Sprintf("Adaptação: +1 de %s (Reservas cheias, descartado)", ResourceLabels[resource]), nil
	}
	return fmt.Sprintf("Adaptação: +1 de %s", ResourceLabels[resource]), nil
}

// ExpeditionComplication applies the consequence of one Pressure chosen by the Mestre.
func (s *Service) ExpeditionComplication(ctx context.Context, campaignID string, in ComplicationInput) error {
	amount := min(max(in.Amount, 1), maxHarm)
	var hurt *effects.Suggestion
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		e, err := findExpedition(r, in.ID)
		if err != nil {
			return err
		}
		if e.Status != ExpDone || e.Complications < 1 || len(in.Note) > maxRegionNote {
			return bad("Não há Pressões sobrando nesta expedição")
		}
		line, sg, err := complicate(r, e, in, amount)
		if err != nil {
			return err
		}
		hurt = sg
		e.Complications--
		e.Log = append(e.Log, line)
		return nil
	})
	if err != nil || hurt == nil || s.links.Effects == nil {
		return err
	}
	hurt.CampaignID = campaignID
	if v, cerr := s.character(ctx, campaignID, hurt.CharacterID); cerr == nil {
		hurt.CharacterName = v.Name
	}
	_, err = s.links.Effects.Suggest(ctx, *hurt)
	return err
}

func complicate(r *Refuge, e *Expedition, in ComplicationInput, amount int) (string, *effects.Suggestion, error) {
	switch in.Kind {
	case "encounter":
		return "Pressão: encontro. " + in.Note, nil, nil
	case "loss":
		if !IsResource(in.Resource) {
			return "", nil, bad("Escolha o recurso perdido")
		}
		lost := min(amount, r.Resources[in.Resource])
		r.Resources[in.Resource] -= lost
		return fmt.Sprintf("Pressão: perda de %d de %s", lost, ResourceLabels[in.Resource]), nil, nil
	case "injury":
		if !slices.Contains(e.CharacterIDs, in.CharacterID) {
			return "", nil, bad("Escolha um personagem da expedição")
		}
		msg := fmt.Sprintf("Ferimento na expedição (%d de dano)", amount)
		sg := &effects.Suggestion{CharacterID: in.CharacterID, Source: "Expedição", Trigger: "action", Effect: effects.EffectDamage, Amount: amount, Message: msg}
		return "Pressão: ferimento (dano sugerido ao Mestre)", sg, nil
	}
	return "", nil, bad("Complicação inválida")
}
