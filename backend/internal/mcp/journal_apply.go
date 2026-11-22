package mcp

import (
	"context"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

type inverseStep func(ctx context.Context, t Table, s Session, e entry) error

var inverseSteps = []inverseStep{
	applyCells, applyObjects, applyToken, applyAmbience, applyCondition, applyRemoveTokens, applyPlaceTokens,
}

func applyInverse(ctx context.Context, t Table, s Session, e entry) error {
	for _, step := range inverseSteps {
		if err := step(ctx, t, s, e); err != nil {
			return err
		}
	}
	return nil
}

func applyCells(ctx context.Context, t Table, s Session, e entry) error {
	if len(e.inv.Cells) == 0 {
		return nil
	}
	return t.Scenes.Paint(ctx, s.CampaignID, e.view.SceneID, "", maps.Patch{Cells: e.inv.Cells})
}

func applyObjects(ctx context.Context, t Table, s Session, e entry) error {
	if e.inv.Objects == nil {
		return nil
	}
	return t.Scenes.EditObjects(ctx, s.CampaignID, e.view.SceneID, "", *e.inv.Objects)
}

func applyToken(ctx context.Context, t Table, s Session, e entry) error {
	k := e.inv.Token
	if k == nil {
		return nil
	}
	return t.Tokens.Move(ctx, s.Sender, s.CampaignID, k.TokenID, k.X, k.Z, true)
}

func applyAmbience(ctx context.Context, t Table, s Session, e entry) error {
	if e.inv.Ambience == nil {
		return nil
	}
	return t.Ambience.Set(ctx, s.CampaignID, *e.inv.Ambience)
}

func applyCondition(ctx context.Context, t Table, s Session, e entry) error {
	k := e.inv.Condition
	if k == nil {
		return nil
	}
	return t.Ambience.SetCondition(ctx, s.Sender, s.CampaignID, k.CharacterID, k.Condition, k.Active)
}

func applyRemoveTokens(ctx context.Context, t Table, s Session, e entry) error {
	for _, id := range e.inv.RemoveTokens {
		if err := t.Tokens.Remove(ctx, s.Sender, s.CampaignID, id); err != nil {
			return err
		}
	}
	return nil
}

func applyPlaceTokens(ctx context.Context, t Table, s Session, e entry) error {
	for _, sp := range e.inv.PlaceTokens {
		if err := t.Tokens.Place(ctx, s.Sender, s.CampaignID, e.view.SceneID, sp.CharacterID, sp.X, sp.Z); err != nil {
			return err
		}
	}
	return nil
}
