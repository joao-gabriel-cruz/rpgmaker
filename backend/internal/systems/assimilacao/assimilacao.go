// Package assimilacao is the Assimilação RPG system module (rules arrive in later tickets).
package assimilacao

// System implements systems.System for Assimilação RPG.
type System struct{}

// New returns the Assimilação module.
func New() System { return System{} }

// ID is the stable identifier stored in campaigns.
func (System) ID() string { return "assimilacao" }

// Name is the display name.
func (System) Name() string { return "Assimilação RPG" }

// Version is the rules module version.
func (System) Version() string { return "0.1.0" }
