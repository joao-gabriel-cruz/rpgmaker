// Package rolls owns dice rolls: the two-step Assimilação flow and the generic XdY+Z roller.
package rolls

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// Roll kinds, statuses and visibilities.
const (
	KindAssimilacao = "assimilacao"
	KindGeneric     = "generic"
	StatusPending   = "pending"
	StatusFinal     = "final"
	VisPublic       = "public"
	VisGM           = "gm"
)

// Events broadcast to the table.
const (
	EventRolled    = "roll.rolled"
	EventFinalized = "roll.finalized"
)

var (
	// ErrNotFound means the roll does not exist in the campaign.
	ErrNotFound = errors.New("roll not found")
)

// Record is a stored roll with the names resolved for display.
type Record struct {
	ID            string
	CampaignID    string
	PlaySessionID *string
	UserID        string
	UserName      string
	CharacterID   *string
	CharacterName string
	Visibility    string
	Kind          string
	Spec          json.RawMessage
	Dice          json.RawMessage
	Result        json.RawMessage
	Status        string
	Seed          int64
	CreatedAt     time.Time
}

// assimilacaoSpec is the stored spec of an Assimilação roll.
type assimilacaoSpec struct {
	Request assimilacao.RollRequest `json:"request"`
	Pool    assimilacao.RollSpec    `json:"pool"`
}

// genericSpec is the stored spec of an XdY+Z roll.
type genericSpec struct {
	Expression string `json:"expression"`
	Count      int    `json:"count"`
	Sides      int    `json:"sides"`
	Modifier   int    `json:"modifier"`
}

// Entry is a roll as sent to clients, both in roll.rolled events and in the log.
type Entry struct {
	RollID        string                   `json:"rollId"`
	UserID        string                   `json:"userId"`
	UserName      string                   `json:"userName"`
	CharacterID   *string                  `json:"characterId"`
	CharacterName string                   `json:"characterName,omitempty"`
	Kind          string                   `json:"kind"`
	Visibility    string                   `json:"visibility"`
	Status        string                   `json:"status"`
	Dice          []assimilacao.Die        `json:"dice"`
	KeptCount     int                      `json:"keptCount,omitempty"`
	Suggested     []int                    `json:"suggested,omitempty"`
	Request       *assimilacao.RollRequest `json:"request,omitempty"`
	Expression    string                   `json:"expression,omitempty"`
	Modifier      int                      `json:"modifier,omitempty"`
	Total         *int                     `json:"total,omitempty"`
	Result        json.RawMessage          `json:"result,omitempty"`
	Seed          int64                    `json:"seed"`
	CreatedAt     time.Time                `json:"createdAt"`
}

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// finalResult is the stored and broadcast outcome; HealthLevel names the level behind the penalty.
type finalResult struct {
	assimilacao.Result
	HealthLevel string `json:"healthLevel,omitempty"`
}
