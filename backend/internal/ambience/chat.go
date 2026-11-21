package ambience

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// MaxSayLen is the longest speech bubble, in characters.
const MaxSayLen = 140

var emojis = []string{"👍", "😂", "😱", "❤️", "💀", "❓"}

type said struct {
	CharacterID string `json:"characterId"`
	Name        string `json:"name"`
	Text        string `json:"text,omitempty"`
	Emoji       string `json:"emoji,omitempty"`
	At          int64  `json:"at"`
}

// speaker loads the character and checks that the sender may speak for it (owner or Mestre).
func (s *Service) speaker(ctx context.Context, snd realtime.Sender, campaignID, charID string) (said, error) {
	m := campaigns.Member{CampaignID: campaignID, UserID: snd.UserID, Role: snd.Role}
	v, err := s.chars.Get(ctx, m, charID)
	if err != nil {
		return said{}, realtime.UserError("Personagem não encontrado")
	}
	if snd.Role != campaigns.RoleGM && (v.OwnerID == nil || *v.OwnerID != snd.UserID) {
		return said{}, realtime.UserError("Este personagem não é seu")
	}
	return said{CharacterID: charID, Name: v.Name, At: time.Now().UnixMilli()}, nil
}

// Say broadcasts a speech bubble for a character to the whole table.
func (s *Service) Say(ctx context.Context, snd realtime.Sender, campaignID, charID, text string) error {
	text = strings.TrimSpace(text)
	if text == "" || utf8.RuneCountInString(text) > MaxSayLen {
		return realtime.UserError("A fala deve ter de 1 a 140 caracteres")
	}
	ev, err := s.speaker(ctx, snd, campaignID, charID)
	if err != nil {
		return err
	}
	ev.Text = text
	s.events.Broadcast(campaignID, realtime.ToAll, EventSaid, ev)
	return nil
}

// React broadcasts a quick emoji reaction above a character's token.
func (s *Service) React(ctx context.Context, snd realtime.Sender, campaignID, charID, emoji string) error {
	if !oneOf(emojis, emoji) {
		return realtime.UserError("Reação inválida")
	}
	ev, err := s.speaker(ctx, snd, campaignID, charID)
	if err != nil {
		return err
	}
	ev.Emoji = emoji
	s.events.Broadcast(campaignID, realtime.ToAll, EventReacted, ev)
	return nil
}
