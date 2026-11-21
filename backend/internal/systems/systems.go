// Package systems defines the game-system module contract and the registry of available systems.
package systems

import "encoding/json"

// System is the minimal contract every game-system module implements.
type System interface {
	ID() string
	Name() string
	Version() string
}

// Issue is a sheet validation finding; the GM decides whether it blocks or only warns.
type Issue struct {
	Severity string `json:"severity"`
	Field    string `json:"field"`
	Message  string `json:"message"`
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
