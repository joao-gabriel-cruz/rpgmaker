package actions

import (
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

func wantDenied(t *testing.T, err error, part string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), part) {
		t.Fatalf("err = %v, want it to contain %q", err, part)
	}
}

func TestPerformPermissions(t *testing.T) {
	r := newRig()
	melee := PerformInput{ActionID: "melee_attack", Target: Target{Kind: TargetToken, TokenID: "tok-npc"}}
	wantDenied(t, perform(r, realtime.Sender{UserID: otherUsr, Role: campaigns.RolePlayer}, melee), "seus personagens")
	if err := perform(r, realtime.Sender{UserID: "gm", Role: campaigns.RoleGM}, melee); err != nil {
		t.Fatalf("Mestre must act for anyone: %v", err)
	}
	wantDenied(t, perform(r, anaSender, PerformInput{ActionID: "nope"}), "não encontrada")
	melee.RollChoice = 5
	wantDenied(t, perform(r, anaSender, melee), "Rolagem inválida")
}

func TestPerformTargetValidation(t *testing.T) {
	r := newRig()
	wantDenied(t, perform(r, anaSender, PerformInput{ActionID: "melee_attack", Target: Target{TokenID: "ghost"}}), "alvo válido")
	wantDenied(t, perform(r, anaSender, PerformInput{ActionID: "persuade", Target: Target{TokenID: "tok-far"}}), "fora de alcance")
	wantDenied(t, perform(r, anaSender, PerformInput{ActionID: "ranged_attack", Target: Target{TokenID: "tok-far"}}), "linha de visão")
	wantDenied(t, perform(r, anaSender, PerformInput{ActionID: "search", Target: Target{Kind: TargetObject, ObjectID: "x"}}), "objeto")
	x, z := 99, 0
	wantDenied(t, perform(r, anaSender, PerformInput{ActionID: "search", Target: Target{Kind: TargetCell, X: &x, Z: &z}}), "dentro do mapa")
	wantDenied(t, perform(r, anaSender, PerformInput{ActionID: "melee_attack", Target: Target{TokenID: "tok-far"}}), "fora de alcance")
	if len(r.rolls.got) != 0 || len(r.ev.list) != 0 {
		t.Fatalf("invalid actions must not roll or announce: %d %d", len(r.rolls.got), len(r.ev.list))
	}
	r.svc.scenes = fakeScenes{active: false}
	wantDenied(t, perform(r, anaSender, PerformInput{ActionID: "perceive"}), "cena ativa")
}

func TestPerformRollsAndAnnouncesWithTheContractPayload(t *testing.T) {
	r := newRig()
	r.chars.sheet = []byte(`{"inventory":{"body":[{"name":"Facão","kind":"melee","attackAnimation":"slash",
		"effects":{"keptBonus":{"actionCategory":"combat","amount":1}}}],"backpack":[]}}`)
	err := perform(r, anaSender, PerformInput{ActionID: "melee_attack", RollChoice: 1, Target: Target{TokenID: "tok-npc"},
		Modifiers: Modifiers{Effort: true}})
	if err != nil {
		t.Fatal(err)
	}
	got := r.rolls.got[0]
	if got.Instinct != "reaction" || got.Aptitude != "weapons" || got.BonusKept != 1 || !got.Effort || got.CharacterID != charA {
		t.Fatalf("roll input = %+v", got)
	}
	e, ok := r.ev.last(EventStarted)
	if !ok || !e.to(realtime.Sender{Role: campaigns.RolePlayer, UserID: otherUsr}) {
		t.Fatal("action.started must reach everyone")
	}
	p := e.payload.(map[string]any)
	tgt := p["target"].(Target)
	if p["animation"] != "melee_strike" || p["actorTokenId"] != "tok-a" || p["rollId"] != "roll-1" ||
		p["actionId"] != "melee_attack" || p["variant"] != "slash" || tgt.TokenID != "tok-npc" || *tgt.X != 4 {
		t.Fatalf("payload = %+v", p)
	}
}

func TestWeaponBonusOnlyForTheActionCategory(t *testing.T) {
	r := newRig()
	r.chars.sheet = []byte(`{"inventory":{"body":[{"name":"Lupa","kind":"tool","effects":{"keptBonus":{"actionCategory":"exploration","amount":2}}}]}}`)
	if err := perform(r, anaSender, PerformInput{ActionID: "perceive"}); err != nil {
		t.Fatal(err)
	}
	if err := perform(r, anaSender, PerformInput{ActionID: "dance"}); err != nil {
		t.Fatal(err)
	}
	if r.rolls.got[0].BonusKept != 2 || r.rolls.got[1].BonusKept != 0 {
		t.Fatalf("bonuses = %d %d", r.rolls.got[0].BonusKept, r.rolls.got[1].BonusKept)
	}
}
