package rounds

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// EventUpdated is broadcast to everyone on each change; payload is {state}.
const EventUpdated = "round.updated"

// Register wires the GM commands round.start/next/prev/end and GET /api/campaigns/{id}/round.
func Register(mux *http.ServeMux, hub *realtime.Hub, svc *Service, requireMember func(http.Handler) http.Handler) {
	cmds := map[string]func(context.Context, string) (State, error){
		"round.start": svc.Start, "round.next": svc.Next, "round.prev": svc.Prev, "round.end": svc.End,
	}
	for typ, fn := range cmds {
		hub.Handle(typ, command(hub, fn), realtime.GMOnly)
	}
	mux.Handle("GET /api/campaigns/{id}/round", requireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		st, err := svc.Get(r.Context(), m.CampaignID)
		if err != nil {
			http.Error(w, `{"error":"Erro interno. Tente novamente"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(st)
	})))
}

func command(hub *realtime.Hub, fn func(context.Context, string) (State, error)) realtime.HandlerFunc {
	return func(ctx context.Context, cmd realtime.Command) error {
		st, err := fn(ctx, cmd.CampaignID)
		if err != nil {
			return userFacing(err)
		}
		hub.Broadcast(cmd.CampaignID, realtime.ToAll, EventUpdated, map[string]State{"state": st})
		return nil
	}
}

func userFacing(err error) error {
	switch {
	case errors.Is(err, errInactive):
		return realtime.UserError("Nenhum conflito em andamento")
	case errors.Is(err, errAtStart):
		return realtime.UserError("O conflito já está no início")
	}
	return err
}
