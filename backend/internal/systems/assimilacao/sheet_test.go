package assimilacao

import (
	"encoding/json"
	"reflect"
	"testing"
)

const contractExample = `{
  "identity": {"name":"Ana","portraitUrl":null,"generation":"post_collapse","occupation":"Médica","markingEvent":"Queda",
    "personalPurposes":["a","b"],"collectivePurposes":["c","d"]},
  "instincts": {"influence":1,"perception":2,"potency":2,"reaction":1,"resolution":1,"sagacity":2},
  "knowledge": {"biology":1,"erudition":0,"engineering":0,"geography":0,"medicine":2,"security":0},
  "practices": {"weapons":0,"athletics":1,"expression":0,"stealth":0,"crafting":0,"survival":1},
  "tugOfWar": {"determinationLevel":9,"determinationPoints":8,"assimilationLevel":1,"assimilationPoints":0},
  "health": {"damage":3},
  "inventory": {"body":[{"name":"Faca","quality":1,"scarcity":0,"pressure":0,"traits":["afiada"]}],"backpack":[]},
  "characteristics": [{"name":"Rápida","level":1,"description":"x"}],
  "xp": {"total": 7},
  "assimilations": [{"name":"Garras","suit":"spades","description":"y"}],
  "notes": "n"
}`

func TestParseSheetDefaults(t *testing.T) {
	s, err := ParseSheet([]byte(`{"identity":{"name":"Bia"},"instincts":{"potency":3}}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Identity.Name != "Bia" || s.Instincts.Potency != 3 || s.Instincts.Reaction != 1 {
		t.Fatalf("unexpected merge: %+v", s)
	}
	if s.TugOfWar.DeterminationLevel != 9 || s.XP.Total != 7 || s.Inventory.Body == nil {
		t.Fatalf("defaults missing: %+v", s)
	}
}

func TestParseSheetInvalid(t *testing.T) {
	if _, err := ParseSheet([]byte(`{`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	s, err := ParseSheet([]byte(contractExample))
	if err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(s)
	var a, b any
	_ = json.Unmarshal([]byte(contractExample), &a)
	_ = json.Unmarshal(out, &b)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("round trip mismatch:\n%s", out)
	}
	if issues := Validate(s); len(issues) != 0 {
		t.Fatalf("example should be valid: %+v", issues)
	}
}

func TestNewSheetValid(t *testing.T) {
	if issues := Validate(NewSheet()); len(issues) != 0 {
		t.Fatalf("new sheet issues: %+v", issues)
	}
}
