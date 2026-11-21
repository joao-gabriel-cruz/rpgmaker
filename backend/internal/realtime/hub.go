package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"sort"
	"sync"
)

// Hub owns all campaign rooms and the command registry. Safe for concurrent use.
type Hub struct {
	mu       sync.Mutex
	rooms    map[string]map[*conn]struct{}
	handlers map[string]route
	opts     Options
}

type route struct {
	fn     HandlerFunc
	guards []Guard
}

// Options tunes the hub.
type Options struct {
	// OriginPatterns are extra allowed Origin hosts (path.Match style); same-host is always allowed.
	OriginPatterns []string
}

// NewHub creates a hub with the built-in "ping" command registered.
func NewHub(opts Options) *Hub {
	h := &Hub{rooms: map[string]map[*conn]struct{}{}, handlers: map[string]route{}, opts: opts}
	h.Handle("ping", func(_ context.Context, cmd Command) error {
		h.Broadcast(cmd.CampaignID, ToUser(cmd.Sender.UserID), EventPong, nil)
		return nil
	})
	return h
}

// Handle registers the handler for a command type. Guards run in order before it
// (none means any member). Register at startup; re-registering a type panics.
func (h *Hub) Handle(typ string, fn HandlerFunc, guards ...Guard) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, dup := h.handlers[typ]; dup {
		panic("realtime: duplicate handler for " + typ)
	}
	h.handlers[typ] = route{fn: fn, guards: guards}
}

// Broadcast sends an event to the campaign's connected members accepted by to.
// It never blocks: connections with a full queue are dropped and will reconnect.
func (h *Hub) Broadcast(campaignID string, to Filter, typ string, payload any) {
	data, err := encode(typ, payload)
	if err != nil {
		slog.Error("realtime encode", "type", typ, "err", err)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sendLocked(campaignID, to, data)
}

func encode(typ string, payload any) ([]byte, error) {
	m := Message{Type: typ}
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		m.Payload = raw
	}
	return json.Marshal(m)
}

func (h *Hub) sendLocked(campaignID string, to Filter, data []byte) {
	for c := range h.rooms[campaignID] {
		if to(c.sender) {
			c.enqueue(data)
		}
	}
}

func (h *Hub) join(campaignID string, c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[campaignID]
	if room == nil {
		room = map[*conn]struct{}{}
		h.rooms[campaignID] = room
	}
	room[c] = struct{}{}
	h.presenceLocked(campaignID)
}

func (h *Hub) leave(campaignID string, c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[campaignID]
	delete(room, c)
	if len(room) == 0 {
		delete(h.rooms, campaignID)
		return
	}
	for other := range room {
		if other.sender.UserID == c.sender.UserID {
			return
		}
	}
	h.presenceLocked(campaignID)
}

// Online lists distinct connected users of a campaign, sorted by name.
func (h *Hub) Online(campaignID string) []Sender {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.onlineLocked(campaignID)
}

func (h *Hub) onlineLocked(campaignID string) []Sender {
	seen := map[string]bool{}
	users := []Sender{}
	for c := range h.rooms[campaignID] {
		if !seen[c.sender.UserID] {
			seen[c.sender.UserID] = true
			users = append(users, c.sender)
		}
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].Name != users[j].Name {
			return users[i].Name < users[j].Name
		}
		return users[i].UserID < users[j].UserID
	})
	return users
}

func (h *Hub) presenceLocked(campaignID string) {
	data, err := encode(EventPresence, map[string]any{"users": h.onlineLocked(campaignID)})
	if err != nil {
		slog.Error("realtime presence", "err", err)
		return
	}
	h.sendLocked(campaignID, ToAll, data)
}

func (h *Hub) lookup(typ string) (route, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	r, ok := h.handlers[typ]
	return r, ok
}
