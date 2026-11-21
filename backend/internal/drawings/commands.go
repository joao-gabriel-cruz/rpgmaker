package drawings

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Registrar is the part of the realtime hub used to register commands.
type Registrar interface {
	Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard)
}

type addIn struct {
	SceneID string  `json:"sceneId"`
	Color   string  `json:"color"`
	Width   int     `json:"width"`
	Points  []Point `json:"points"`
}

type rulerIn struct {
	SceneID string `json:"sceneId"`
	From    *Point `json:"from"`
	To      *Point `json:"to"`
}

// RegisterCommands wires drawing.add, drawing.remove, drawing.clear and ruler.move.
func RegisterCommands(r Registrar, svc *Service) {
	handle(r, "drawing.add", func(ctx context.Context, c realtime.Command, in addIn) error {
		return svc.Add(ctx, c.Sender, c.CampaignID, in.SceneID, in.Color, in.Width, in.Points)
	})
	handle(r, "drawing.remove", func(ctx context.Context, c realtime.Command, in struct {
		DrawingID string `json:"drawingId"`
	}) error {
		return svc.Remove(ctx, c.Sender, c.CampaignID, in.DrawingID)
	})
	handle(r, "drawing.clear", func(ctx context.Context, c realtime.Command, in struct {
		SceneID string `json:"sceneId"`
	}) error {
		return svc.Clear(ctx, c.Sender, c.CampaignID, in.SceneID)
	}, realtime.GMOnly)
	handle(r, "ruler.move", func(_ context.Context, c realtime.Command, in rulerIn) error {
		return svc.RulerMove(c.CampaignID, c.Sender, in.SceneID, in.From, in.To)
	})
}

func handle[T any](r Registrar, typ string, fn func(context.Context, realtime.Command, T) error, guards ...realtime.Guard) {
	r.Handle(typ, func(ctx context.Context, cmd realtime.Command) error {
		var in T
		if err := json.Unmarshal(cmd.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		return fn(ctx, cmd, in)
	}, guards...)
}

// Register mounts GET /api/campaigns/{id}/scenes/{sceneId}/drawings.
func Register(mux *http.ServeMux, svc *Service, cmp *campaigns.Handlers) {
	mux.Handle("GET /api/campaigns/{id}/scenes/{sceneId}/drawings", cmp.RequireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		out, err := svc.List(r.Context(), m, r.PathValue("sceneId"))
		switch {
		case errors.Is(err, ErrNotFound):
			write(w, http.StatusNotFound, map[string]string{"error": "Cena não encontrada"})
		case err != nil:
			slog.Error("list drawings failed", "err", err)
			write(w, http.StatusInternalServerError, map[string]string{"error": "Erro interno. Tente novamente"})
		default:
			write(w, http.StatusOK, out)
		}
	})))
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
