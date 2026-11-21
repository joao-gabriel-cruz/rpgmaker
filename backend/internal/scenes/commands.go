package scenes

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Registrar is the part of the realtime hub used to register commands.
type Registrar interface {
	Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard)
}

// RegisterCommands wires the Mestre-only scene commands: map.paint, map.objects and scene.activate.
func RegisterCommands(r Registrar, svc *Service) {
	type paint struct {
		SceneID string     `json:"sceneId"`
		Patch   maps.Patch `json:"patch"`
	}
	type objects struct {
		SceneID string      `json:"sceneId"`
		Patch   ObjectPatch `json:"patch"`
	}
	type activate struct {
		SceneID string `json:"sceneId"`
	}
	handle(r, "map.paint", func(ctx context.Context, cmd realtime.Command, in paint) error {
		return svc.Paint(ctx, cmd.CampaignID, in.SceneID, cmd.Sender.UserID, in.Patch)
	})
	handle(r, "map.objects", func(ctx context.Context, cmd realtime.Command, in objects) error {
		return svc.EditObjects(ctx, cmd.CampaignID, in.SceneID, cmd.Sender.UserID, in.Patch)
	})
	handle(r, "scene.activate", func(ctx context.Context, cmd realtime.Command, in activate) error {
		return svc.Activate(ctx, cmd.CampaignID, in.SceneID)
	})
}

// handle registers a GM-only command that decodes its payload into T.
func handle[T any](r Registrar, typ string, fn func(context.Context, realtime.Command, T) error) {
	r.Handle(typ, func(ctx context.Context, cmd realtime.Command) error {
		var in T
		if err := json.Unmarshal(cmd.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		return commandError(fn(ctx, cmd, in))
	}, realtime.GMOnly)
}

// commandError turns domain errors into messages shown to the Mestre.
func commandError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound):
		return realtime.UserError("Cena não encontrada")
	case errors.Is(err, ErrInvalidPatch):
		return realtime.UserError("Alteração inválida: " + trimPrefix(err))
	}
	return err
}

func trimPrefix(err error) string {
	msg := err.Error()
	const p = "scenes: invalid patch: "
	if len(msg) > len(p) && msg[:len(p)] == p {
		return msg[len(p):]
	}
	return msg
}
