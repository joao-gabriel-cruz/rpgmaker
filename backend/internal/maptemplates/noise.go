package maptemplates

import "math"

// Noise is seeded 2D value noise in [0,1]; the same seed gives the same field.
type Noise struct{ seed uint64 }

// NewNoise creates a noise field for the seed.
func NewNoise(seed int64) Noise {
	return Noise{seed: uint64(seed)*0x9e3779b97f4a7c15 + 0x1234567} //nolint:gosec // wrapping is intended for hashing
}

func (n Noise) lattice(ix, iz int) float64 {
	h := n.seed ^ uint64(int64(ix))*0xbf58476d1ce4e5b9 ^ uint64(int64(iz))*0x94d049bb133111eb //nolint:gosec // wrapping is intended
	h ^= h >> 30
	h *= 0xbf58476d1ce4e5b9
	h ^= h >> 27
	h *= 0x94d049bb133111eb
	h ^= h >> 31
	return float64(h>>11) / float64(1<<53)
}

func smooth(t float64) float64 { return t * t * (3 - 2*t) }

// At samples smooth value noise at (x, z); one lattice step is one unit.
func (n Noise) At(x, z float64) float64 {
	fx, fz := math.Floor(x), math.Floor(z)
	ix, iz := int(fx), int(fz)
	tx, tz := smooth(x-fx), smooth(z-fz)
	a := n.lattice(ix, iz)*(1-tx) + n.lattice(ix+1, iz)*tx
	b := n.lattice(ix, iz+1)*(1-tx) + n.lattice(ix+1, iz+1)*tx
	return a*(1-tz) + b*tz
}

// FBM sums octaves of noise (each twice as fine, half as strong), normalized to [0,1].
func (n Noise) FBM(x, z float64, octaves int) float64 {
	sum, amp, norm, freq := 0.0, 1.0, 0.0, 1.0
	for i := 0; i < octaves; i++ {
		sum += n.At(x*freq+float64(i)*17.3, z*freq-float64(i)*9.1) * amp
		norm += amp
		amp /= 2
		freq *= 2
	}
	return sum / norm
}
