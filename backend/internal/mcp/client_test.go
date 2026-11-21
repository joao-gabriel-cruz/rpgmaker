package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// callTool runs one tool through a real in-memory MCP client/server pair.
func callTool(t *testing.T, tbl Table, j *Journal, name string, args any) (string, bool) {
	t.Helper()
	return callToolAs(t, tbl, j, Session{CampaignID: "c1"}, name, args)
}

func callToolAs(t *testing.T, tbl Table, j *Journal, s Session, name string, args any) (string, bool) {
	t.Helper()
	ctx := context.Background()
	st, ct := sdk.NewInMemoryTransports()
	ss, err := newServer(tbl, j, s).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ss.Close() }()
	cs, err := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cs.Close() }()
	res, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*sdk.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String(), res.IsError
}

func TestShapeFieldsAreDescribedInSchema(t *testing.T) {
	ctx := context.Background()
	st, ct := sdk.NewInMemoryTransports()
	ss, err := newServer(newRig("s1").t, NewJournal(&recNotifier{}), Session{CampaignID: "c1"}).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ss.Close() }()
	cs, err := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cs.Close() }()
	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tl := range tools.Tools {
		if tl.Name != "paint_terrain" {
			continue
		}
		raw, _ := json.Marshal(tl.InputSchema)
		for _, want := range []string{"first corner (inclusive)", "thickness in cells", "do not combine with heightDelta"} {
			if !strings.Contains(string(raw), want) {
				t.Fatalf("schema lacks %q: %s", want, raw)
			}
		}
		return
	}
	t.Fatal("paint_terrain not listed")
}
