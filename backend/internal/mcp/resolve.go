package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

var (
	errNoActiveScene = errors.New("nenhuma cena ativa")
	errOpenDraft     = errors.New("há um rascunho aberto nesta cena; aplique ou descarte no editor antes")
)

// activeScene loads the campaign's active scene as the Mestre sees it (draft included).
func activeScene(ctx context.Context, t Table, s Session) (scenes.Scene, error) {
	list, err := t.Scenes.List(ctx, s.CampaignID)
	if err != nil {
		return scenes.Scene{}, fmt.Errorf("list scenes: %w", err)
	}
	for _, sum := range list {
		if sum.IsActive {
			return t.Scenes.Get(ctx, s.CampaignID, sum.ID, campaigns.RoleGM)
		}
	}
	return scenes.Scene{}, errNoActiveScene
}

// requireNoDraft refuses live edits while the editor holds an unapplied draft.
func requireNoDraft(sc scenes.Scene) error {
	if sc.Draft != nil {
		return errOpenDraft
	}
	return nil
}

func tableView(ctx context.Context, t Table, s Session, sc scenes.Scene) (TableView, error) {
	amb, err := t.Ambience.Get(ctx, s.CampaignID, sc.ID)
	if err != nil {
		return TableView{}, fmt.Errorf("get ambience: %w", err)
	}
	list, err := sceneTokens(ctx, t, s, sc.ID)
	if err != nil {
		return TableView{}, err
	}
	infos := make([]TokenInfo, 0, len(list))
	for _, tk := range list {
		infos = append(infos, tokenInfo(tk))
	}
	return TableView{Scene: sc, Ambience: amb, Tokens: infos}, nil
}

func tokenInfo(tk tokens.Token) TokenInfo {
	var st struct {
		Conditions []string `json:"conditions"`
	}
	_ = json.Unmarshal(tk.Status, &st)
	return TokenInfo{ID: tk.ID, Name: tk.Name, X: tk.X, Z: tk.Z, Player: tk.OwnerID != nil, Conditions: st.Conditions}
}

// sceneTokens lists the scene's tokens as the Mestre sees them.
func sceneTokens(ctx context.Context, t Table, s Session, sceneID string) ([]tokens.Token, error) {
	m := campaigns.Member{CampaignID: s.CampaignID, UserID: s.Sender.UserID, Role: campaigns.RoleGM}
	list, err := t.Tokens.List(ctx, m, sceneID)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	return list, nil
}
