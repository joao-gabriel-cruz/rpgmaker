package conflict

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Registrar is the part of the realtime hub used to register commands.
type Registrar interface {
	Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard)
}

func handler[T any](fn func(context.Context, realtime.Command, T) error) realtime.HandlerFunc {
	return func(ctx context.Context, cmd realtime.Command) error {
		var in T
		if err := json.Unmarshal(cmd.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		return commandError(fn(ctx, cmd, in))
	}
}

// RegisterCommands wires conflict.save|close|threat.roll (Mestre) and conflict.assign (owner or Mestre).
func RegisterCommands(r Registrar, svc *Service) {
	r.Handle("conflict.save", handler(func(ctx context.Context, c realtime.Command, in Input) error {
		return svc.Save(ctx, c.CampaignID, c.Sender, in)
	}), realtime.GMOnly)
	r.Handle("conflict.close", handler(func(ctx context.Context, c realtime.Command, _ struct{}) error {
		return svc.Close(ctx, c.CampaignID)
	}), realtime.GMOnly)
	r.Handle("conflict.threat.roll", handler(func(ctx context.Context, c realtime.Command, in struct {
		ThreatID string `json:"threatId"`
	}) error {
		return svc.RollThreat(ctx, c.CampaignID, c.Sender, in.ThreatID)
	}), realtime.GMOnly)
	r.Handle("conflict.assign", handler(func(ctx context.Context, c realtime.Command, in struct {
		RollID      string `json:"rollId"`
		ObjectiveID string `json:"objectiveId"`
	}) error {
		return svc.Assign(ctx, c.CampaignID, c.Sender, in.RollID, in.ObjectiveID)
	}))
}

// RegisterHTTP mounts GET /api/campaigns/{id}/conflict (role-filtered; null when none).
func RegisterHTTP(mux *http.ServeMux, svc *Service, requireMember func(http.Handler) http.Handler) {
	mux.Handle("GET /api/campaigns/{id}/conflict", requireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		c, err := svc.View(r.Context(), m)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			slog.Error("conflict view failed", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Erro interno. Tente novamente"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"conflict": c})
	})))
}
