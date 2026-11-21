package mcp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/ambience"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

const journalCap = 20

// TokenPos is a token position to restore.
type TokenPos struct {
	TokenID string
	X, Z    int
}

// ConditionState is a character condition to restore.
type ConditionState struct {
	CharacterID, Condition string
	Active                 bool
}

// Inverse holds what is needed to revert one change; only set fields are applied.
type Inverse struct {
	Cells     []maps.CellPatch
	Objects   *scenes.ObjectPatch
	Token     *TokenPos
	Ambience  *ambience.Ambience
	Condition *ConditionState
}

// ChangeView is the public face of a journal entry.
type ChangeView struct {
	ID      string    `json:"id"`
	SceneID string    `json:"sceneId"`
	Summary string    `json:"summary"`
	At      time.Time `json:"at"`
}

// Notifier is the part of the realtime hub the journal broadcasts through.
type Notifier interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

type entry struct {
	view ChangeView
	inv  Inverse
}

// Journal keeps the newest changes per campaign, with their inverses.
type Journal struct {
	mu     sync.Mutex
	notify Notifier
	now    func() time.Time
	byCamp map[string][]entry // newest first
}

// NewJournal builds an empty journal.
func NewJournal(n Notifier) *Journal {
	return &Journal{notify: n, now: time.Now, byCamp: map[string][]entry{}}
}

func newChangeID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Record stores a change and tells the Mestre.
func (j *Journal) Record(campaignID, sceneID, summary string, inv Inverse) ChangeView {
	j.mu.Lock()
	v := ChangeView{ID: newChangeID(), SceneID: sceneID, Summary: summary, At: j.now()}
	list := append([]entry{{v, inv}}, j.byCamp[campaignID]...)
	if len(list) > journalCap {
		list = list[:journalCap]
	}
	j.byCamp[campaignID] = list
	// Broadcast never blocks; holding the lock keeps snapshots ordered.
	j.notify.Broadcast(campaignID, realtime.ToGM, "mcp.changed",
		map[string]any{"changes": viewsOf(list), "latest": v})
	j.mu.Unlock()
	return v
}

// List returns the campaign's changes, newest first.
func (j *Journal) List(campaignID string) []ChangeView {
	j.mu.Lock()
	defer j.mu.Unlock()
	return viewsOf(j.byCamp[campaignID])
}

func viewsOf(list []entry) []ChangeView {
	out := make([]ChangeView, len(list))
	for i, e := range list {
		out[i] = e.view
	}
	return out
}

// take removes an entry (id "" = newest).
func (j *Journal) take(campaignID, id string) (entry, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	list := j.byCamp[campaignID]
	for i, e := range list {
		if id == "" || e.view.ID == id {
			j.byCamp[campaignID] = append(append([]entry{}, list[:i]...), list[i+1:]...)
			return e, true
		}
	}
	return entry{}, false
}

// putBack reinserts by time, so undos that ran meanwhile cannot misplace it.
func (j *Journal) putBack(campaignID string, e entry) {
	j.mu.Lock()
	defer j.mu.Unlock()
	list := j.byCamp[campaignID]
	at := 0
	for at < len(list) && !list[at].view.At.Before(e.view.At) {
		at++
	}
	list = append(list[:at], append([]entry{e}, list[at:]...)...)
	if len(list) > journalCap {
		list = list[:journalCap]
	}
	j.byCamp[campaignID] = list
}

func (j *Journal) broadcastList(campaignID string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.notify.Broadcast(campaignID, realtime.ToGM, "mcp.changed",
		map[string]any{"changes": viewsOf(j.byCamp[campaignID])})
}

// Undo reverts a change (id "" = newest) and returns its summary.
func (j *Journal) Undo(ctx context.Context, t Table, s Session, id string) (string, error) {
	e, ok := j.take(s.CampaignID, id)
	if !ok {
		return "", realtime.UserError("Essa mudança não existe mais (talvez já tenha sido desfeita)")
	}
	if err := j.apply(ctx, t, s, e); err != nil {
		if errors.Is(err, scenes.ErrNotFound) || errors.Is(err, errSceneGone) {
			j.broadcastList(s.CampaignID)
			return "", errSceneGone
		}
		j.putBack(s.CampaignID, e)
		return "", err
	}
	j.broadcastList(s.CampaignID)
	return e.view.Summary, nil
}

var errSceneGone = realtime.UserError("A cena dessa mudança não existe mais")

func (j *Journal) apply(ctx context.Context, t Table, s Session, e entry) error {
	if e.view.SceneID != "" {
		sc, err := t.Scenes.Get(ctx, s.CampaignID, e.view.SceneID, campaigns.RoleGM)
		if errors.Is(err, scenes.ErrNotFound) {
			return errSceneGone
		}
		if err != nil {
			return err
		}
		if sc.Draft != nil {
			return realtime.UserError("há um rascunho aberto nesta cena; aplique ou descarte no editor antes")
		}
	}
	return applyInverse(ctx, t, s, e)
}

func applyInverse(ctx context.Context, t Table, s Session, e entry) error {
	inv, c, sc := e.inv, s.CampaignID, e.view.SceneID
	if len(inv.Cells) > 0 {
		if err := t.Scenes.Paint(ctx, c, sc, "", maps.Patch{Cells: inv.Cells}); err != nil {
			return err
		}
	}
	if inv.Objects != nil {
		if err := t.Scenes.EditObjects(ctx, c, sc, "", *inv.Objects); err != nil {
			return err
		}
	}
	if inv.Token != nil {
		if err := t.Tokens.Move(ctx, s.Sender, c, inv.Token.TokenID, inv.Token.X, inv.Token.Z, true); err != nil {
			return err
		}
	}
	if inv.Ambience != nil {
		if err := t.Ambience.Set(ctx, c, *inv.Ambience); err != nil {
			return err
		}
	}
	if inv.Condition != nil {
		k := inv.Condition
		return t.Ambience.SetCondition(ctx, s.Sender, c, k.CharacterID, k.Condition, k.Active)
	}
	return nil
}
