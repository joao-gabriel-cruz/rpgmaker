package actions

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

func syms(s, a, p int) assimilacao.Symbols {
	return assimilacao.Symbols{Success: s, Adaptation: a, Pressure: p}
}

func TestDamageSuggestsSuccessesWithAMinimumOfOne(t *testing.T) {
	r := newRig()
	_ = perform(r, anaSender, PerformInput{ActionID: "melee_attack", Target: Target{TokenID: "tok-npc"}})
	out := finish(r, syms(3, 0, 0))
	sg := r.ef.got[0]
	if out.Kind != KindDamage || sg.Effect != effects.EffectDamage || sg.Amount != 3 || sg.CharacterID != charNPC || sg.SceneID != sceneID {
		t.Fatalf("out=%+v sg=%+v", out, sg)
	}
	e, _ := r.ev.last(EventResolved)
	res := e.payload.(map[string]any)["result"].(map[string]int)
	if res["success"] != 3 || e.payload.(map[string]any)["animation"] != "melee_strike" {
		t.Fatalf("resolved = %+v", e.payload)
	}
}

func TestMissSuggestsNothing(t *testing.T) {
	r := newRig()
	_ = perform(r, anaSender, PerformInput{ActionID: "melee_attack", Target: Target{TokenID: "tok-npc"}})
	if out := finish(r, syms(0, 2, 1)); out.Kind != KindNone || len(r.ef.got) != 0 {
		t.Fatalf("out=%+v suggestions=%d", out, len(r.ef.got))
	}
}

func TestFirstAidSuggestsHeal(t *testing.T) {
	r := newRig()
	_ = perform(r, anaSender, PerformInput{ActionID: "first_aid", Target: Target{TokenID: "tok-npc"}})
	out := finish(r, syms(2, 0, 0))
	if out.Kind != KindHeal || r.ef.got[0].Effect != effects.EffectHeal || r.ef.got[0].Amount != 2 {
		t.Fatalf("out=%+v sg=%+v", out, r.ef.got)
	}
}

func TestPerceiveScansSuccessesPlusTwoAroundTheActor(t *testing.T) {
	r := newRig()
	r.loot.found = 1
	_ = perform(r, anaSender, PerformInput{ActionID: "perceive"})
	out := finish(r, syms(2, 0, 0))
	has := func(x, z int) bool {
		for _, c := range r.loot.target.Cells {
			if c == (maps.Coord{X: x, Z: z}) {
				return true
			}
		}
		return false
	}
	if !has(3+4, 3) || !has(3, 3-4) || has(3+5, 3) || r.loot.successes != 2 || out.Kind != KindLoot {
		t.Fatalf("radius wrong: out=%+v successes=%d", out, r.loot.successes)
	}
}

func TestSearchTargetsTheObject(t *testing.T) {
	r := newRig()
	_ = perform(r, anaSender, PerformInput{ActionID: "search", Target: Target{Kind: TargetObject, ObjectID: "crate1"}, RollChoice: 0})
	// the crate is 2 cells away from (3,3): in range 2
	out := finish(r, syms(1, 0, 0))
	if r.loot.target.ObjectID != "crate1" || out.Kind != KindNone {
		t.Fatalf("target=%+v out=%+v", r.loot.target, out)
	}
}

func TestRestTestWithSafeSpotAndCards(t *testing.T) {
	r := newRig()
	r.svc.tokens = fakeTokens{list: []tokens.Token{{ID: "tok-a", SceneID: sceneID, CharacterID: charA, Name: "Ana", X: 3, Z: 3}}}
	_ = perform(r, anaSender, PerformInput{ActionID: "rest"})
	out := finish(r, syms(2, 1, 1))
	msg := r.ef.got[0].Message
	if out.Kind != KindRest || !containsAll(msg, "recuperar 2", "local seguro", "♥ Evolutiva", "♦ Adaptativa", "♠ Inoportuna") {
		t.Fatalf("msg = %q out=%+v", msg, out)
	}
	e, ok := r.ev.last(EventCards)
	cards := e.payload.(map[string]any)["cards"].([]Card)
	if !ok || len(cards) != 3 || cards[0].Suit != "hearts" || cards[1].Suit != "diamonds" || cards[2].Suit != "spades" {
		t.Fatalf("cards = %+v", cards)
	}
	if e.to(realtime.Sender{Role: campaigns.RolePlayer}) || !e.to(realtime.Sender{Role: campaigns.RoleGM}) {
		t.Fatal("cards go to the Mestre only")
	}
}

func TestRestAwayFromFireRecoversOneAndNoSymbolsNoCards(t *testing.T) {
	r := newRig()
	r.svc.tokens = fakeTokens{list: []tokens.Token{{ID: "tok-a", SceneID: sceneID, CharacterID: charA, Name: "Ana", X: 10, Z: 10}}}
	_ = perform(r, anaSender, PerformInput{ActionID: "rest"})
	finish(r, syms(1, 0, 0))
	if !containsAll(r.ef.got[0].Message, "recuperar 1") {
		t.Fatalf("msg = %q", r.ef.got[0].Message)
	}
	_ = perform(r, anaSender, PerformInput{ActionID: "rest"})
	r.ev.list = nil
	finish(r, syms(0, 0, 0))
	if _, ok := r.ev.last(EventCards); ok || !containsAll(r.ef.got[1].Message, "nenhuma Determinação") {
		t.Fatalf("msg = %q cards=%v", r.ef.got[1].Message, ok)
	}
}

func TestMoraleNoticeAndAcceptCard(t *testing.T) {
	r := newRig()
	_ = perform(r, anaSender, PerformInput{ActionID: "dance"})
	if out := finish(r, syms(1, 0, 0)); out.Kind != KindMorale || r.ef.got[0].Effect != effects.EffectNotice {
		t.Fatalf("out=%+v", out)
	}
	_ = perform(r, anaSender, PerformInput{ActionID: "sneak"})
	if out := finish(r, syms(1, 0, 0)); out.Kind != KindNotice || len(r.ef.got) != 1 {
		t.Fatalf("notice must not create a suggestion: %+v", out)
	}
	if err := r.svc.AcceptCard(ctx, camp, charA, Card{Suit: "hearts", Name: "Pele Calejada"}); err != nil {
		t.Fatal(err)
	}
	if !containsAll(string(r.chars.sheet), "Pele Calejada", "hearts") {
		t.Fatalf("sheet = %s", r.chars.sheet)
	}
	if err := r.svc.AcceptCard(ctx, camp, charA, Card{Suit: "clubs", Name: "X"}); err == nil {
		t.Fatal("unknown suit accepted")
	}
}

func TestFinalizedIgnoresRollsWithoutAction(t *testing.T) {
	r := newRig()
	r.svc.Finalized(ctx, camp, "someone-elses-roll")
	if len(r.ev.list) != 0 {
		t.Fatal("must stay silent")
	}
}
