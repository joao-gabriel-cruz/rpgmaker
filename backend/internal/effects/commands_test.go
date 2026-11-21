package effects

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

type fakeRegistrar struct{ guards map[string][]realtime.Guard }

func (f *fakeRegistrar) Handle(typ string, _ realtime.HandlerFunc, g ...realtime.Guard) {
	f.guards[typ] = g
}

func TestCommandsAreGMOnly(t *testing.T) {
	reg := &fakeRegistrar{guards: map[string][]realtime.Guard{}}
	RegisterCommands(reg, newRig().svc)
	player := realtime.Sender{UserID: "p", Role: campaigns.RolePlayer}
	for _, typ := range []string{"effect.confirm", "effect.dismiss"} {
		g := reg.guards[typ]
		if len(g) != 1 || g[0](player) == nil || g[0](gm) != nil {
			t.Fatalf("%s must be GM only", typ)
		}
	}
}
