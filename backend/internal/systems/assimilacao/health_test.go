package assimilacao

import "testing"

var deriveCases = []struct {
	name         string
	potency, res int
	damage       int
	chars        []Characteristic
	boxes, level int
	levelName    string
	penalty      int
	deb, inc     bool
	spent, avail int
}{
	{"healthy", 2, 1, 0, nil, 4, 0, "Saudável", 0, false, false, 0, 7},
	{"laceration", 2, 1, 9, nil, 4, 2, "Laceração", 1, false, false, 0, 7},
	{"scrape", 2, 1, 5, nil, 4, 1, "Escoriação", 0, false, false, 0, 7},
	{"wounds", 1, 1, 9, nil, 3, 3, "Ferimentos", 1, false, false, 0, 7},
	{"debilitated", 1, 1, 12, nil, 3, 4, "Debilitação", 2, true, false, 0, 7},
	{"incapacitated", 1, 1, 100, nil, 3, 5, "Incapacitação", 2, true, true, 0, 7},
	{"negative damage", 1, 1, -4, nil, 3, 0, "Saudável", 0, false, false, 0, 7},
	{"xp", 1, 1, 0, []Characteristic{{Level: 1}, {Level: 2}}, 3, 0, "Saudável", 0, false, false, 6, 1},
}

func TestDerive(t *testing.T) {
	for _, tc := range deriveCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSheet()
			s.Instincts.Potency, s.Instincts.Resolution = tc.potency, tc.res
			s.Health.Damage, s.Characteristics = tc.damage, tc.chars
			d := Derive(s)
			want := Derived{tc.boxes, tc.level, tc.levelName, tc.penalty, tc.deb, tc.inc, tc.spent, tc.avail}
			if d != want {
				t.Fatalf("got %+v want %+v", d, want)
			}
		})
	}
}
