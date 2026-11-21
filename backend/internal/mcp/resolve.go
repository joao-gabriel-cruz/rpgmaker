package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
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

// pick resolves ref against items by exact id first, then case-insensitive name.
func pick[T any](list []T, ref string, id, name func(T) string) (T, error) {
	var zero T
	for _, it := range list {
		if id(it) == ref {
			return it, nil
		}
	}
	var found []T
	for _, it := range list {
		if strings.EqualFold(name(it), ref) {
			found = append(found, it)
		}
	}
	switch len(found) {
	case 0:
		return zero, newUserError(fmt.Sprintf("não encontrei %q", ref))
	case 1:
		return found[0], nil
	}
	opts := make([]string, len(found))
	for i, it := range found {
		opts[i] = fmt.Sprintf("%s (%s)", name(it), id(it))
	}
	return zero, newUserError(fmt.Sprintf("%q é ambíguo: %s", ref, strings.Join(opts, ", ")))
}

func findToken(list []tokens.Token, ref string) (tokens.Token, error) {
	return pick(list, ref, func(t tokens.Token) string { return t.ID }, func(t tokens.Token) string { return t.Name })
}

func findCharacter(list []characters.View, ref string) (characters.View, error) {
	return pick(list, ref, func(c characters.View) string { return c.ID }, func(c characters.View) string { return c.Name })
}
