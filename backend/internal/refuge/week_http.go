package refuge

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

type weekCmd struct {
	Adjust map[string]int `json:"adjust"`
}

type buildingCmd struct {
	Kind       string         `json:"kind"`
	ID         string         `json:"id"`
	Allocation map[string]int `json:"allocation"`
}

func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.ContentLength == 0 {
		return true
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Requisição inválida"})
		return false
	}
	return true
}

func registerWeekHTTP(mux *http.ServeMux, svc *Service, h *campaigns.Handlers) {
	run := func(do func(context.Context, campaigns.Member, map[string]int) (WeekReport, error)) http.Handler {
		return h.RequireGM(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body weekCmd
			if !decodeBody(w, r, &body) {
				return
			}
			m, _ := campaigns.MemberFrom(r)
			rep, err := do(r.Context(), m, body.Adjust)
			if err != nil {
				respond(w, Refuge{}, err)
				return
			}
			writeJSON(w, http.StatusOK, rep)
		}))
	}
	mux.Handle("POST /api/campaigns/{id}/refuge/week/preview", run(svc.PreviewWeek))
	mux.Handle("POST /api/campaigns/{id}/refuge/week/confirm", run(svc.ConfirmWeek))
}

func registerWeekCommands(hub *realtime.Hub, svc *Service) {
	hub.Handle("refuge.week.preview", func(ctx context.Context, c realtime.Command) error {
		var p weekCmd
		if err := json.Unmarshal(c.Payload, &p); err != nil {
			return realtime.UserError("Requisição inválida")
		}
		rep, err := svc.PreviewWeek(ctx, memberOf(c), p.Adjust)
		if err == nil {
			hub.Broadcast(c.CampaignID, realtime.ToUser(c.Sender.UserID), EventWeekPreview, map[string]WeekReport{"report": rep})
		}
		return err
	}, realtime.GMOnly)
	hub.Handle("refuge.week.confirm", func(ctx context.Context, c realtime.Command) error {
		var p weekCmd
		if err := json.Unmarshal(c.Payload, &p); err != nil {
			return realtime.UserError("Requisição inválida")
		}
		_, err := svc.ConfirmWeek(ctx, memberOf(c), p.Adjust)
		return err
	}, realtime.GMOnly)
}

type buildingRun func(context.Context, campaigns.Member, buildingCmd) (Refuge, error)

func handleBuilding(hub *realtime.Hub, typ string, run func(context.Context, realtime.Command, buildingCmd) error, guards ...realtime.Guard) {
	hub.Handle(typ, func(ctx context.Context, c realtime.Command) error {
		var p buildingCmd
		if err := json.Unmarshal(c.Payload, &p); err != nil {
			return realtime.UserError("Requisição inválida")
		}
		return run(ctx, c, p)
	}, guards...)
}

func registerBuildingCommands(hub *realtime.Hub, svc *Service) {
	handleBuilding(hub, "refuge.building.propose", func(ctx context.Context, c realtime.Command, p buildingCmd) error {
		_, err := svc.ProposeBuilding(ctx, memberOf(c), c.Sender.Name, p.Kind)
		return err
	})
	gm := func(typ string, f buildingRun) {
		handleBuilding(hub, typ, func(ctx context.Context, c realtime.Command, p buildingCmd) error {
			_, err := f(ctx, memberOf(c), p)
			return err
		}, realtime.GMOnly)
	}
	gm("refuge.building.add", func(ctx context.Context, m campaigns.Member, p buildingCmd) (Refuge, error) {
		return svc.AddBuilding(ctx, m, p.Kind)
	})
	gm("refuge.building.approve", func(ctx context.Context, m campaigns.Member, p buildingCmd) (Refuge, error) {
		return svc.ResolveProposal(ctx, m, p.ID, true)
	})
	gm("refuge.building.reject", func(ctx context.Context, m campaigns.Member, p buildingCmd) (Refuge, error) {
		return svc.ResolveProposal(ctx, m, p.ID, false)
	})
	gm("refuge.building.upgrade", func(ctx context.Context, m campaigns.Member, p buildingCmd) (Refuge, error) {
		return svc.UpgradeBuilding(ctx, m, p.ID)
	})
	gm("refuge.building.cancel", func(ctx context.Context, m campaigns.Member, p buildingCmd) (Refuge, error) {
		return svc.CancelBuilding(ctx, m, p.ID)
	})
	gm("refuge.work.allocate", func(ctx context.Context, m campaigns.Member, p buildingCmd) (Refuge, error) {
		return svc.Allocate(ctx, m, p.Allocation)
	})
}
