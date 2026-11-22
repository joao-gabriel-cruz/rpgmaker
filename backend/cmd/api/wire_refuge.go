package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/items"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/refuge"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

// compendiumSource reads campaign compendium items in the sheet item format for expedition finds.
func compendiumSource(d deps, chars *characters.Service, rf *refuge.Service) refuge.ItemSource {
	sceneSvc := scenes.NewService(scenes.NewStore(d.pool), d.hub)
	tokenSvc := tokens.NewService(tokens.NewStore(d.pool), sceneSvc, d.vision.Tap())
	svc := items.NewService(items.NewStore(d.pool), chars, rf, d.effects, tokenSvc, sceneSvc, d.hub)
	return func(ctx context.Context, campaignID, itemID string) (json.RawMessage, error) {
		list, err := svc.List(ctx, campaignID)
		if err != nil {
			return nil, err
		}
		for _, it := range list {
			if it.ID == itemID {
				return items.SheetItem(it), nil
			}
		}
		return nil, errors.New("item not found")
	}
}
