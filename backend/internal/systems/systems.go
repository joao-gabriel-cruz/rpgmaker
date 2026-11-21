// Package systems defines the game-system module contract and the registry of available systems.
package systems

import "encoding/json"

// System is the minimal contract every game-system module implements.
type System interface {
	ID() string
	Name() string
	Version() string
}

// Issue is a sheet validation finding. Severity is SeverityBlocking (always rejects) or
// SeverityLimit (a creation limit the campaign may only warn about).
type Issue struct {
	Code     string `json:"code,omitempty"`
	Severity string `json:"severity"`
	Field    string `json:"field"`
	Message  string `json:"message"`
}

// Issue severities.
const (
	SeverityBlocking = "blocking"
	SeverityLimit    = "limit"
)

// SheetTemplater is optionally implemented by systems that can create a blank sheet.
type SheetTemplater interface {
	NewSheet(name string) (json.RawMessage, error)
}

// SheetValidator is optionally implemented by systems that validate character sheets.
type SheetValidator interface {
	ValidateSheet(sheet, campaignSettings json.RawMessage) []Issue
}

// SheetDeriver is optionally implemented by systems that compute derived sheet values.
type SheetDeriver interface {
	DeriveSheet(sheet json.RawMessage) (json.RawMessage, error)
}

// Roller is optionally implemented by systems that build and resolve rolls.
type Roller interface {
	BuildRoll(sheet, request json.RawMessage) (json.RawMessage, error)
	ResolveRoll(spec json.RawMessage) (json.RawMessage, error)
}
