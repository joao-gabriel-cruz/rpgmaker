package assimilacao

import "testing"

func TestFaceSymbols(t *testing.T) {
	want := map[int]Symbols{
		1: {}, 2: {}, 3: {0, 0, 1}, 4: {0, 0, 1}, 5: {0, 1, 1}, 6: {1, 0, 0}, 7: {2, 0, 0},
		8: {1, 1, 0}, 9: {1, 1, 1}, 10: {2, 0, 1}, 11: {1, 2, 1}, 12: {0, 0, 2},
	}
	for face, w := range want {
		kind := D12
		if got := FaceSymbols(kind, face); got != w {
			t.Errorf("d12 face %d: got %+v want %+v", face, got, w)
		}
	}
	if FaceSymbols(D6, 7) != (Symbols{}) || FaceSymbols(D10, 11) != (Symbols{}) || FaceSymbols(D6, 0) != (Symbols{}) {
		t.Error("out-of-range faces must be empty")
	}
	if FaceSymbols(D6, 6) != (Symbols{Success: 1}) || FaceSymbols(D10, 10) != (Symbols{Success: 2, Pressure: 1}) {
		t.Error("d6/d10 faces use the same table")
	}
	if Sides("x") != 0 || Sides(D10) != 10 {
		t.Error("sides")
	}
}
