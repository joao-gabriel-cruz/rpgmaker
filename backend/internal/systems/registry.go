package systems

import (
	"sort"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// Registry holds the available systems by ID.
type Registry struct{ byID map[string]System }

// NewRegistry builds a registry from the given systems.
func NewRegistry(list ...System) *Registry {
	r := &Registry{byID: map[string]System{}}
	for _, s := range list {
		r.byID[s.ID()] = s
	}
	return r
}

// Default returns the registry with every built-in system.
func Default() *Registry { return NewRegistry(assimilacaoModule{assimilacao.New()}) }

// Get returns the system with the given ID.
func (r *Registry) Get(id string) (System, bool) {
	s, ok := r.byID[id]
	return s, ok
}

// List returns the systems ordered by ID.
func (r *Registry) List() []System {
	out := make([]System, 0, len(r.byID))
	for _, s := range r.byID {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}
