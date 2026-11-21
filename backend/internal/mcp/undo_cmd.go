package mcp

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

// RegisterUndo wires the Mestre-only "mcp.undo" command.
func RegisterUndo(r scenes.Registrar, j *Journal, t Table) {
	r.Handle("mcp.undo", func(ctx context.Context, cmd realtime.Command) error {
		var in struct {
			ChangeID string `json:"changeId"`
		}
		if err := json.Unmarshal(cmd.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		s := Session{CampaignID: cmd.CampaignID, Sender: cmd.Sender}
		_, err := j.Undo(ctx, t, s, in.ChangeID)
		return err
	}, realtime.GMOnly)
}
