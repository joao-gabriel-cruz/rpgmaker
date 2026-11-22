package refuge

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

type idArg struct {
	ID string `json:"id"`
}

func gmCommand[T any](hub *realtime.Hub, typ string, run func(context.Context, realtime.Command, T) error) {
	hub.Handle(typ, func(ctx context.Context, c realtime.Command) error {
		var in T
		if err := json.Unmarshal(c.Payload, &in); err != nil {
			return realtime.UserError("Requisição inválida")
		}
		return run(ctx, c, in)
	}, realtime.GMOnly)
}

func anyCommand[T any](hub *realtime.Hub, typ string, run func(context.Context, realtime.Command, T) error) {
	hub.Handle(typ, func(ctx context.Context, c realtime.Command) error {
		var in T
		if err := json.Unmarshal(c.Payload, &in); err != nil {
			return realtime.UserError("Requisição inválida")
		}
		return run(ctx, c, in)
	})
}

func registerCrisisCommands(hub *realtime.Hub, svc *Service) {
	gmCommand(hub, "refuge.crisis.create", func(ctx context.Context, c realtime.Command, in CrisisInput) error {
		_, err := svc.AddCrisis(ctx, c.CampaignID, c.Sender.UserID, in)
		return err
	})
	gmCommand(hub, "refuge.crisis.update", func(ctx context.Context, c realtime.Command, in CrisisPatch) error {
		return svc.UpdateCrisis(ctx, c.CampaignID, in)
	})
	gmCommand(hub, "refuge.crisis.delete", func(ctx context.Context, c realtime.Command, in idArg) error {
		return svc.DeleteCrisis(ctx, c.CampaignID, in.ID)
	})
	gmCommand(hub, "refuge.crisis.confirm", func(ctx context.Context, c realtime.Command, in idArg) error {
		return svc.ConfirmCrisis(ctx, c.CampaignID, in.ID)
	})
	gmCommand(hub, "refuge.crisis.settings", func(ctx context.Context, c realtime.Command, in CrisisSettings) error {
		return svc.SetCrisisSettings(ctx, c.CampaignID, in)
	})
	gmCommand(hub, "refuge.crisis.roll", func(ctx context.Context, c realtime.Command, in struct {
		CrisisID string `json:"crisisId"`
	}) error {
		return svc.RollSeverity(ctx, c.CampaignID, c.Sender, in.CrisisID)
	})
	gmCommand(hub, "refuge.crisis.discard", func(ctx context.Context, c realtime.Command, in struct {
		CrisisID string `json:"crisisId"`
		Index    int    `json:"index"`
	}) error {
		return svc.DiscardDie(ctx, c.CampaignID, in.CrisisID, in.Index)
	})
	registerSuggestionCommands(hub, svc)
}

type suggestionArg struct {
	CrisisID     string `json:"crisisId"`
	SuggestionID string `json:"suggestionId"`
}

func registerSuggestionCommands(hub *realtime.Hub, svc *Service) {
	gmCommand(hub, "refuge.crisis.apply", func(ctx context.Context, c realtime.Command, in suggestionArg) error {
		return svc.ResolveSuggestion(ctx, c.CampaignID, in.CrisisID, in.SuggestionID, true)
	})
	gmCommand(hub, "refuge.crisis.dismiss", func(ctx context.Context, c realtime.Command, in suggestionArg) error {
		return svc.ResolveSuggestion(ctx, c.CampaignID, in.CrisisID, in.SuggestionID, false)
	})
}

func registerResidentCommands(hub *realtime.Hub, svc *Service) {
	gmCommand(hub, "refuge.resident.add", func(ctx context.Context, c realtime.Command, in ResidentInput) error {
		return svc.AddResident(ctx, c.CampaignID, c.Sender.UserID, in)
	})
	gmCommand(hub, "refuge.resident.update", func(ctx context.Context, c realtime.Command, in ResidentInput) error {
		return svc.UpdateResident(ctx, c.CampaignID, in)
	})
	gmCommand(hub, "refuge.resident.remove", func(ctx context.Context, c realtime.Command, in idArg) error {
		return svc.RemoveResident(ctx, c.CampaignID, in.ID)
	})
}

func registerExpeditionCommands(hub *realtime.Hub, svc *Service) {
	anyCommand(hub, "refuge.expedition.start", func(ctx context.Context, c realtime.Command, in ExpeditionInput) error {
		return svc.StartExpedition(ctx, c.CampaignID, c.Sender, in)
	})
	anyCommand(hub, "refuge.expedition.cancel", func(ctx context.Context, c realtime.Command, in idArg) error {
		return svc.CancelExpedition(ctx, c.CampaignID, c.Sender, in.ID)
	})
	anyCommand(hub, "refuge.expedition.roll", func(ctx context.Context, c realtime.Command, in struct {
		ID       string `json:"id"`
		Instinct string `json:"instinct"`
	}) error {
		return svc.RollExpedition(ctx, c.CampaignID, c.Sender, in.ID, in.Instinct)
	})
	gmCommand(hub, "refuge.expedition.approve", func(ctx context.Context, c realtime.Command, in idArg) error {
		return svc.ApproveExpedition(ctx, c.CampaignID, in.ID)
	})
	gmCommand(hub, "refuge.expedition.extra", func(ctx context.Context, c realtime.Command, in ExtraInput) error {
		return svc.ExpeditionExtra(ctx, c.CampaignID, in)
	})
	gmCommand(hub, "refuge.expedition.complication", func(ctx context.Context, c realtime.Command, in ComplicationInput) error {
		return svc.ExpeditionComplication(ctx, c.CampaignID, in)
	})
}
