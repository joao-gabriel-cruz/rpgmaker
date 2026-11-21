package assimilacao

import "testing"

func TestValidateNPC(t *testing.T) {
	if got := ValidateNPC(NewNPCSheet()); len(got) != 0 {
		t.Fatalf("blank NPC issues = %+v", got)
	}
	s := NewNPCSheet()
	s.Instincts.Potency = 12 // beyond PC limits, fine for NPCs
	if got := ValidateNPC(s); len(got) != 0 {
		t.Errorf("high instinct rejected: %+v", got)
	}
	s.Health.BoxesPerLevel = 0
	s.Practices.Weapons = -1
	if got := ValidateNPC(s); len(got) != 2 || !got[0].Blocking {
		t.Errorf("issues = %+v", got)
	}
}

func TestParseNPCSheetDefaultsAndDerive(t *testing.T) {
	s, err := ParseNPCSheet([]byte(`{"health":{"damage":7,"boxesPerLevel":2}}`))
	if err != nil || s.Threats == nil || s.Instincts.Sagacity != 1 {
		t.Fatalf("parse = %+v %v", s, err)
	}
	if d := DeriveNPC(s); d.CurrentLevel != 3 || d.BoxesPerLevel != 2 {
		t.Errorf("derived = %+v", d)
	}
}
