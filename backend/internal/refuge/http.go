package refuge

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Register mounts GET/PUT /api/campaigns/{id}/refuge and the item commands on the hub.
func Register(mux *http.ServeMux, hub *realtime.Hub, svc *Service, h *campaigns.Handlers) {
	mux.Handle("GET /api/campaigns/{id}/refuge", h.RequireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		out, err := svc.Get(r.Context(), m.CampaignID)
		respond(w, out, err)
	})))
	mux.Handle("PUT /api/campaigns/{id}/refuge", h.RequireGM(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var d Editable
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&d); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Requisição inválida"})
			return
		}
		m, _ := campaigns.MemberFrom(r)
		out, err := svc.Update(r.Context(), m, d)
		respond(w, out, err)
	})))
	registerCommands(hub, svc)
}

type moveCmd struct {
	CharacterID string `json:"characterId"`
	From        string `json:"from"`
	To          string `json:"to"`
	Index       int    `json:"index"`
	RefugeIndex int    `json:"refugeIndex"`
}

func registerCommands(hub *realtime.Hub, svc *Service) {
	handle := func(typ string, run func(context.Context, campaigns.Member, moveCmd) error) {
		hub.Handle(typ, func(ctx context.Context, c realtime.Command) error {
			var p moveCmd
			if err := json.Unmarshal(c.Payload, &p); err != nil {
				return realtime.UserError("Requisição inválida")
			}
			return run(ctx, memberOf(c), p)
		})
	}
	handle("refuge.item.deposit", func(ctx context.Context, m campaigns.Member, p moveCmd) error {
		_, err := svc.Deposit(ctx, m, p.CharacterID, p.From, p.Index)
		return err
	})
	handle("refuge.item.withdraw", func(ctx context.Context, m campaigns.Member, p moveCmd) error {
		_, err := svc.Withdraw(ctx, m, p.CharacterID, p.To, p.RefugeIndex)
		return err
	})
}

func memberOf(c realtime.Command) campaigns.Member {
	return campaigns.Member{CampaignID: c.CampaignID, UserID: c.Sender.UserID, Role: c.Sender.Role}
}

func respond(w http.ResponseWriter, v Refuge, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, v)
	case errors.Is(err, ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Dados inválidos: níveis de 0 a 6, população dentro do teto e recursos não negativos"})
	default:
		slog.Error("refuge request failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro interno. Tente novamente"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
