package scenes

import (
	"encoding/json"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

type route struct {
	fn     realtime.HandlerFunc
	guards []realtime.Guard
}

type fakeRegistrar map[string]route

func (f fakeRegistrar) Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard) {
	f[typ] = route{fn, guards}
}

// dispatch mimics the hub: guards first, then the handler.
func (f fakeRegistrar) dispatch(role, typ string, payload any) error {
	r := f[typ]
	s := realtime.Sender{UserID: "u", Role: role}
	for _, g := range r.guards {
		if err := g(s); err != nil {
			return err
		}
	}
	raw, _ := json.Marshal(payload)
	return r.fn(ctx, realtime.Command{CampaignID: camp, Sender: s, Type: typ, Payload: raw})
}

func TestCommandsAreGMOnly(t *testing.T) {
	svc, _, _ := newSvc()
	reg := fakeRegistrar{}
	RegisterCommands(reg, svc)
	sc, _ := svc.Create(ctx, camp, "A", 16, 16)
	paint := map[string]any{"sceneId": sc.ID, "patch": map[string]any{"cells": []map[string]any{{"x": 1, "z": 1, "terrain": "stone", "height": 5}}}}
	for _, typ := range []string{"map.paint", "map.objects", "scene.activate"} {
		err := reg.dispatch("player", typ, paint)
		if err == nil || err.Error() != "Você não tem permissão para fazer isso" {
			t.Errorf("%s as player: %v", typ, err)
		}
	}
	if err := reg.dispatch("gm", "map.paint", paint); err != nil {
		t.Errorf("gm paint: %v", err)
	}
}

func TestCommandsReportUserErrors(t *testing.T) {
	svc, _, n := newSvc()
	reg := fakeRegistrar{}
	RegisterCommands(reg, svc)
	sc, _ := svc.Create(ctx, camp, "A", 16, 16)
	bad := map[string]any{"sceneId": sc.ID, "patch": map[string]any{"add": []map[string]any{{"id": "a", "prefab": "dragon", "x": 1, "z": 1, "rotation": 0}}}}
	if err := reg.dispatch("gm", "map.objects", bad); err == nil || err.Error() == "" {
		t.Errorf("unknown prefab accepted: %v", err)
	}
	if err := reg.dispatch("gm", "map.paint", "garbage"); err == nil {
		t.Error("garbage payload accepted")
	}
	if err := reg.dispatch("gm", "scene.activate", map[string]any{"sceneId": sc.ID}); err != nil {
		t.Fatal(err)
	}
	if _, ok := n.last(EventActivated); !ok {
		t.Error("no activated event")
	}
}
