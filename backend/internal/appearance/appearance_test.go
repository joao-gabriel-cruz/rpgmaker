package appearance

import (
	"encoding/json"
	"testing"
)

const valid = `{"appearance":{"body":{"build":"broad","height":1.05,"skin":"#aabbcc"},"hair":{"style":"mohawk","color":"#112233"},
"beard":"full","outfit":{"top":"coat","topColor":"#000000","bottom":"skirt","bottomColor":"#ffffff","shoes":"boots","shoesColor":"#123456"},
"accessories":{"backpack":true,"weapon":"bow","mask":"gas","headwear":"helmet","glasses":false},"tokenColor":"#34d399","infected":3}}`

func TestValidate(t *testing.T) {
	cases := map[string]struct {
		sheet string
		bad   bool
	}{
		"no appearance":   {`{"identity":{}}`, false},
		"null":            {`{"appearance":null}`, false},
		"empty object":    {`{"appearance":{}}`, false},
		"full valid":      {valid, false},
		"unknown enum":    {`{"appearance":{"hair":{"style":"afro"}}}`, true},
		"bad color":       {`{"appearance":{"body":{"skin":"red"}}}`, true},
		"short color":     {`{"appearance":{"tokenColor":"#fff"}}`, true},
		"height high":     {`{"appearance":{"body":{"height":1.5}}}`, true},
		"height low":      {`{"appearance":{"body":{"height":0.5}}}`, true},
		"infected range":  {`{"appearance":{"infected":11}}`, true},
		"infected frac":   {`{"appearance":{"infected":2.5}}`, true},
		"wrong bool":      {`{"appearance":{"accessories":{"backpack":"yes"}}}`, true},
		"not an object":   {`{"appearance":"x"}`, true},
		"enum wrong type": {`{"appearance":{"beard":3}}`, true},
	}
	for name, c := range cases {
		got := Validate(json.RawMessage(c.sheet))
		if (len(got) > 0) != c.bad {
			t.Errorf("%s: issues = %+v", name, got)
		}
	}
}
