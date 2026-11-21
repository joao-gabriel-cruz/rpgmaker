package vision

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Registrar is the part of the realtime hub used to register commands.
type Registrar interface {
	Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard)
}

type brushIn struct {
	SceneID string       `json:"sceneId"`
	Cells   []maps.Coord `json:"cells"`
}

type configIn struct {
	SceneID      string `json:"sceneId"`
	Mode         string `json:"mode"`
	VisionRadius int    `json:"visionRadius"`
}

// RegisterCommands wires the Mestre-only fog.reveal, fog.hide and fog.configure commands.
func RegisterCommands(r Registrar, svc *Service) {
	handle(r, "fog.reveal", func(ctx context.Context, c realtime.Command, in brushIn) error {
		return svc.Brush(ctx, c.Sender, c.CampaignID, in.SceneID, in.Cells, true)
	})
	handle(r, "fog.hide", func(ctx context.Context, c realtime.Command, in brushIn) error {
		return svc.Brush(ctx, c.Sender, c.CampaignID, in.SceneID, in.Cells, false)
	})
	handle(r, "fog.configure", func(ctx context.Context, c realtime.Command, in configIn) error {
		return svc.Configure(ctx, c.Sender, c.CampaignID, in.SceneID, in.Mode, in.VisionRadius)
	})
}

func handle[T any](r Registrar, typ string, fn func(context.Context, realtime.Command, T) error) {
	r.Handle(typ, func(ctx context.Context, cmd realtime.Command) error {
		var in T
		if err := json.Unmarshal(cmd.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		return fn(ctx, cmd, in)
	}, realtime.GMOnly)
}

func (s *Service) loadGM(ctx context.Context, snd realtime.Sender, campaignID, sceneID string) (*state, error) {
	if snd.Role != campaigns.RoleGM {
		return nil, deny("Você não tem permissão para fazer isso")
	}
	st, err := s.load(ctx, campaignID, sceneID, snd.Role)
	if errors.Is(err, ErrNotFound) {
		return nil, deny("Cena não encontrada")
	}
	return st, err
}

// Brush reveals or hides the given cells for the players.
func (s *Service) Brush(ctx context.Context, snd realtime.Sender, campaignID, sceneID string, cells []maps.Coord, reveal bool) error {
	if len(cells) > maxBrushCells {
		return deny("Área grande demais")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadGM(ctx, snd, campaignID, sceneID)
	if err != nil {
		return err
	}
	for _, c := range cells {
		if !st.grid.InBounds(c.X, c.Z) {
			return deny("Célula fora do mapa")
		}
		st.revealed.Put(c.X, c.Z, reveal)
	}
	st.fog.Revealed = st.revealed.Encode()
	if err := s.repo.SaveRevealed(ctx, st.sc.ID, st.fog.Revealed); err != nil {
		return err
	}
	s.changed(st)
	return nil
}

// Configure sets the fog mode and the dynamic vision radius of a scene.
func (s *Service) Configure(ctx context.Context, snd realtime.Sender, campaignID, sceneID, mode string, radius int) error {
	if mode != ModeOff && mode != ModeManual && mode != ModeDynamic {
		return deny("Modo de névoa inválido")
	}
	if radius < 1 || radius > MaxRadius {
		return deny("Raio de visão inválido")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadGM(ctx, snd, campaignID, sceneID)
	if err != nil {
		return err
	}
	if err := s.repo.SaveConfig(ctx, st.sc.ID, mode, radius); err != nil {
		return err
	}
	st.fog.Mode, st.fog.Radius = mode, radius
	s.changed(st)
	return nil
}
