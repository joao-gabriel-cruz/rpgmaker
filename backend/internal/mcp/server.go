package mcp

import (
	"context"
	"errors"
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type sessionKey struct{}

const maxBodyBytes = 1 << 20

// NewHandler serves the MCP endpoint; the {secret} path value picks the campaign and acting Mestre.
func NewHandler(l *Links, t Table, j *Journal) http.Handler {
	inner := sdk.NewStreamableHTTPHandler(func(r *http.Request) *sdk.Server {
		s, _ := r.Context().Value(sessionKey{}).(Session)
		return newServer(t, j, s)
	}, &sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := l.Resolve(r.Context(), r.PathValue("secret"))
		if errors.Is(err, ErrInvalidLink) {
			writeError(w, http.StatusNotFound, "Link inválido")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Erro interno")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		inner.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), sessionKey{}, s)))
	})
}

// newServer builds the per-request MCP server; each tools file adds one register call here.
func newServer(t Table, j *Journal, s Session) *sdk.Server {
	srv := sdk.NewServer(&sdk.Implementation{Name: "rpgmaker-mesa", Version: "1"}, nil)
	registerReadTools(srv, t, j, s)
	registerTerrainTools(srv, t, j, s)
	registerObjectTools(srv, t, j, s)
	registerTokenTools(srv, t, j, s)
	registerSceneTools(srv, t, j, s)
	return srv
}

func textResult(text string) *sdk.CallToolResult {
	return &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: text}}}
}
