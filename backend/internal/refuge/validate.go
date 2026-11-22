package refuge

import "slices"

const (
	maxCount     = 99999
	maxPopMax    = 100
	maxMobility  = 99
	defaultCap   = 10
	defaultWeek  = 1
	maxWeekValue = 99999
)

// DefaultSheet is the state of a brand-new Refúgio.
func DefaultSheet() Sheet {
	s := Sheet{Population: Population{Max: defaultCap}, ReservesMax: defaultCap, Week: defaultWeek}
	s.normalize()
	return s
}

// normalize fills missing resource and abundance keys so clients and deliveries can index freely.
func (s *Sheet) normalize() {
	res := make(map[string]int, len(ResourceKeys))
	for _, k := range ResourceKeys {
		res[k] = s.Resources[k]
	}
	s.Resources = res
	ab := make(map[string]string, len(NaturalKeys))
	for _, k := range NaturalKeys {
		ab[k] = s.Region.Abundance[k]
		if ab[k] == "" {
			ab[k] = "none"
		}
	}
	s.Region.Abundance = ab
}

func inRange(v, lo, hi int) bool { return v >= lo && v <= hi }

func (s Sheet) validate() error {
	r := s.Region
	levels := []int{s.Defense, s.Morale, s.Belligerence, r.Size, r.Danger, r.Habitation, r.NaturalResources, r.Contamination, r.Travel}
	for _, v := range levels {
		if !inRange(v, 0, MaxLevel) {
			return ErrInvalid
		}
	}
	p := s.Population
	if !inRange(s.Mobility, 0, maxMobility) || !inRange(p.Max, 1, maxPopMax) || !inRange(p.Level, 0, p.Max) ||
		!inRange(s.ReservesMax, 0, maxCount) || !inRange(s.Week, 1, maxWeekValue) {
		return ErrInvalid
	}
	return s.validateMaps()
}

func (s Sheet) validateMaps() error {
	for k, v := range s.Resources {
		if !IsResource(k) || !inRange(v, 0, maxCount) {
			return ErrInvalid
		}
	}
	for k, v := range s.Region.Abundance {
		if !slices.Contains(NaturalKeys, k) || !slices.Contains(Abundances, v) {
			return ErrInvalid
		}
	}
	return nil
}

func (e Editable) validate() error {
	if len(e.Name) == 0 || len(e.Name) > maxName || len(e.Description) > maxText || len(e.Notes) > maxText {
		return ErrInvalid
	}
	return e.Sheet.validate()
}
