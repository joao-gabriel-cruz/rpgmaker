package realtime

import (
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

// Register mounts GET /api/campaigns/{id}/ws behind requireMember (campaigns.Handlers.RequireMember),
// which authenticates via the session cookie and refuses non-members before the upgrade.
func (h *Hub) Register(mux *http.ServeMux, requireMember func(http.Handler) http.Handler) {
	mux.Handle("GET /api/campaigns/{id}/ws", requireMember(http.HandlerFunc(h.handle)))
}

func (h *Hub) handle(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r)
	m, _ := campaigns.MemberFrom(r)
	h.Serve(w, r, m.CampaignID, Sender{UserID: u.ID, Name: u.DisplayName, Role: m.Role})
}
