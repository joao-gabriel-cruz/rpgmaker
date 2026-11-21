// Package realtime is the per-campaign WebSocket hub: presence, command handlers and filtered events.
package realtime

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

// Message is the wire format in both directions. Client to server messages are commands,
// server to client messages are events.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	ReqID   string          `json:"reqId,omitempty"`
}

// Event types emitted by the hub itself.
const (
	EventPresence = "presence.updated"
	EventError    = "error"
	EventAck      = "ack"
	EventPong     = "pong"
)

// Sender identifies who issued a command (or who is connected).
type Sender struct {
	UserID string `json:"userId"`
	Name   string `json:"displayName"`
	Role   string `json:"role"`
	// ConnID identifies the WebSocket connection (one per tab); set by the hub, never serialized.
	ConnID string `json:"-"`
}

// Command is a validated client request handed to a registered handler.
type Command struct {
	CampaignID string
	Sender     Sender
	ConnID     string // origin connection; use with ExceptConn to skip only the sending tab
	Type       string
	Payload    json.RawMessage
	ReqID      string
}

// HandlerFunc handles a command. Return UserError for a message shown to the user;
// any other error is logged and reported as a generic failure.
type HandlerFunc func(ctx context.Context, cmd Command) error

// Guard declares who may run a command; it returns a non-nil error to deny.
type Guard func(s Sender) error

var errForbidden = UserError("Você não tem permissão para fazer isso")

// GMOnly allows only the campaign's Mestre.
func GMOnly(s Sender) error {
	if s.Role != campaigns.RoleGM {
		return errForbidden
	}
	return nil
}

// AnyMember allows every connected member (the default when no guard is given).
func AnyMember(Sender) error { return nil }

type userError struct{ msg string }

func (e userError) Error() string { return e.msg }

// UserError builds an error whose message (PT-BR) is sent to the command's sender.
func UserError(msg string) error { return userError{msg} }

func userMessage(err error) (string, bool) {
	var ue userError
	if errors.As(err, &ue) {
		return ue.msg, true
	}
	return "", false
}

// UserMessage returns the PT-BR message of an error built with UserError.
func UserMessage(err error) (string, bool) { return userMessage(err) }
