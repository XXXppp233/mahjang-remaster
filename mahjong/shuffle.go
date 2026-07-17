// Package mahjong contains transport- and storage-independent game primitives.
package mahjong

const ShuffleAlgorithm = "FY-XorShift32-v1"

type XorShift32 struct{ state uint32 }

func NewXorShift32(seed uint32) *XorShift32 {
	if seed == 0 {
		seed = 0x6d2b79f5
	}
	return &XorShift32{state: seed}
}

func (r *XorShift32) Next() uint32 {
	x := r.state
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.state = x
	return x
}

// Shuffle applies the versioned Fisher-Yates protocol shared with the frontend.
func Shuffle[T any](values []T, seed uint32) {
	rng := NewXorShift32(seed)
	for i := len(values) - 1; i > 0; i-- {
		j := int(rng.Next() % uint32(i+1))
		values[i], values[j] = values[j], values[i]
	}
}
