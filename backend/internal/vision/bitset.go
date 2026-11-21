package vision

import (
	"encoding/base64"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// Bits is a compact per-cell flag set; cell (x,z) is bit z*w+x, least significant bit first.
type Bits struct {
	w, d int
	b    []byte
}

// NewBits makes an empty set for a w x d grid.
func NewBits(w, d int) Bits { return Bits{w: w, d: d, b: make([]byte, (w*d+7)/8)} }

// DecodeBits reads the base64 wire format; malformed input yields an empty set.
func DecodeBits(s string, w, d int) Bits {
	out := NewBits(w, d)
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return out
	}
	copy(out.b, raw)
	return out
}

// Encode returns the base64 wire format.
func (s Bits) Encode() string { return base64.StdEncoding.EncodeToString(s.b) }

func (s Bits) idx(x, z int) (int, bool) {
	if x < 0 || z < 0 || x >= s.w || z >= s.d {
		return 0, false
	}
	return z*s.w + x, true
}

// Has reports whether the cell is set.
func (s Bits) Has(x, z int) bool {
	i, ok := s.idx(x, z)
	return ok && s.b[i/8]&(1<<(i%8)) != 0
}

// Put sets or clears a cell; out-of-bounds cells are ignored.
func (s Bits) Put(x, z int, on bool) {
	i, ok := s.idx(x, z)
	if !ok {
		return
	}
	if on {
		s.b[i/8] |= 1 << (i % 8)
	} else {
		s.b[i/8] &^= 1 << (i % 8)
	}
}

// Or merges another set of the same size into s.
func (s Bits) Or(o Bits) {
	for i := range s.b {
		if i < len(o.b) {
			s.b[i] |= o.b[i]
		}
	}
}

// FromSet converts a visibility set.
func FromSet(set maps.Set, w, d int) Bits {
	out := NewBits(w, d)
	for c := range set {
		out.Put(c.X, c.Z, true)
	}
	return out
}
