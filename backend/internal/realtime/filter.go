package realtime

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"

// Filter decides, per connected member, whether an event is delivered.
// Any func(Sender) bool works as a custom predicate (e.g. per-player vision).
type Filter func(Sender) bool

// ToAll delivers to every member.
func ToAll(Sender) bool { return true }

// ToGM delivers only to the Mestre.
func ToGM(s Sender) bool { return s.Role == campaigns.RoleGM }

// ToUser delivers only to one user (all of their tabs).
func ToUser(id string) Filter { return func(s Sender) bool { return s.UserID == id } }

// Except delivers to everyone but one user.
func Except(id string) Filter { return func(s Sender) bool { return s.UserID != id } }

// ExceptConn delivers to everyone but one connection (other tabs of the same user still receive).
func ExceptConn(id string) Filter { return func(s Sender) bool { return s.ConnID != id } }

// Or delivers when any of the filters match (e.g. Or(ToGM, ToUser(id))).
func Or(fs ...Filter) Filter {
	return func(s Sender) bool {
		for _, f := range fs {
			if f(s) {
				return true
			}
		}
		return false
	}
}
