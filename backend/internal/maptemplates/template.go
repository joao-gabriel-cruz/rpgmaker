// Package maptemplates holds the Mestre's map templates: deterministic generators registered by id.
package maptemplates

import (
	"fmt"
	"slices"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// Size bounds of a template, in cells per side.
const (
	MinSide = 64
	MaxSide = 128
)

// Ambience is the suggested mood (values of internal/ambience) applied to scenes created from a template.
type Ambience struct {
	TimeOfDay string `json:"timeOfDay"`
	Weather   string `json:"weather"`
}

// Template is a generated map: the same seed always yields the same grid and objects.
type Template struct {
	ID, Name, Description string
	Tags                  []string
	Width, Depth          int
	Ambience              Ambience
	Generate              func(seed int64) (maps.Grid, []maps.Object)
}

// Info is the template metadata sent to the gallery.
type Info struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Width       int      `json:"width"`
	Depth       int      `json:"depth"`
	Ambience    Ambience `json:"ambience"`
}

// Info returns the metadata of the template.
func (t Template) Info() Info {
	return Info{t.ID, t.Name, t.Description, t.Tags, t.Width, t.Depth, t.Ambience}
}

var (
	mu       sync.RWMutex
	registry = map[string]Template{}
)

// galleryOrder is the display order of the catalog (spec 0004); unknown ids follow alphabetically.
var galleryOrder = []string{
	"vila-abandonada", "rodovia-engarrafada", "cidade-em-ruinas", "floresta-clareira", "fazenda", "refugio-escola",
	"festival", "vale-vulcanico", "passo-nevado", "litoral-pier", "pantano", "desfiladeiro-ponte",
}

// Register adds a template; it panics on an invalid or duplicate one (called from init()).
func Register(t Template) {
	if t.ID == "" || t.Generate == nil || t.Name == "" {
		panic("maptemplates: template needs id, name and generator")
	}
	if t.Width < MinSide || t.Depth < MinSide || t.Width > MaxSide || t.Depth > MaxSide {
		panic(fmt.Sprintf("maptemplates: %s has an invalid size %dx%d", t.ID, t.Width, t.Depth))
	}
	mu.Lock()
	defer mu.Unlock()
	if _, dup := registry[t.ID]; dup {
		panic("maptemplates: duplicate template " + t.ID)
	}
	registry[t.ID] = t
}

// All returns every template in gallery order.
func All() []Template {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Template, 0, len(registry))
	for _, t := range registry {
		out = append(out, t)
	}
	slices.SortFunc(out, func(a, b Template) int {
		ia, ib := orderOf(a.ID), orderOf(b.ID)
		if ia != ib {
			return ia - ib
		}
		return compare(a.ID, b.ID)
	})
	return out
}

func compare(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func orderOf(id string) int {
	if i := slices.Index(galleryOrder, id); i >= 0 {
		return i
	}
	return len(galleryOrder)
}

// Get finds a template by id.
func Get(id string) (Template, bool) {
	mu.RLock()
	defer mu.RUnlock()
	t, ok := registry[id]
	return t, ok
}
