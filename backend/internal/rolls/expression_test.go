package rolls

import "testing"

func TestParseExpression(t *testing.T) {
	ok := map[string]genericSpec{
		"2d6+3":    {Expression: "2d6+3", Count: 2, Sides: 6, Modifier: 3},
		"1d20":     {Expression: "1d20", Count: 1, Sides: 20},
		"d8":       {Expression: "d8", Count: 1, Sides: 8},
		"4d6-1":    {Expression: "4d6-1", Count: 4, Sides: 6, Modifier: -1},
		" 2D6 + 3": {Expression: "2d6+3", Count: 2, Sides: 6, Modifier: 3},
		"50d1000":  {Expression: "50d1000", Count: 50, Sides: 1000},
	}
	for in, want := range ok {
		got, err := parseExpression(in)
		if err != nil || got != want {
			t.Errorf("%q = %+v, %v; want %+v", in, got, err, want)
		}
	}
	for _, in := range []string{"", "abc", "2d", "d", "2d6+", "0d6", "51d6", "2d1", "2d1001", "2d6+x", "2d6+99999", "2d6*3"} {
		if _, err := parseExpression(in); err == nil {
			t.Errorf("%q accepted", in)
		}
	}
}
