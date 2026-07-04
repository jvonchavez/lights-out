package sim

// RNG is PCG64-XSL-RR, vendored so that no Go release can change the
// stream. A leaderboard is only trustworthy if today's seed means the same
// thing next year, and math/rand's stream is only guaranteed within a Go
// release line. See docs/Game Design.md, "Determinism contract".
type RNG struct {
	state uint64
	inc   uint64
}

const (
	pcgMultiplier = 6364136223846793005
	pcgIncrement  = 1442695040888963407
)

// NewRNG returns a generator seeded deterministically from seed.
func NewRNG(seed int64) *RNG {
	r := &RNG{state: 0, inc: pcgIncrement}
	r.Uint64()
	r.state += uint64(seed)
	r.Uint64()
	return r
}

// Uint64 advances the state and returns the next value in the stream.
func (r *RNG) Uint64() uint64 {
	old := r.state
	r.state = old*pcgMultiplier + r.inc
	xorshifted := ((old >> 18) ^ old) >> 27
	rot := uint32(old >> 59)
	lo := uint32(xorshifted)
	return uint64(lo>>rot | lo<<((32-rot)&31))
}

// IntN returns a uniform value in [0, n). Rejection sampling keeps it
// unbiased; the modulo alone would favour the low end of the range.
// Panics on n <= 0.
func (r *RNG) IntN(n int) int {
	if n <= 0 {
		panic("sim: IntN requires n > 0")
	}
	bound := uint64(n)
	threshold := (-bound) % bound
	for {
		v := r.Uint64()
		if v >= threshold {
			return int(v % bound)
		}
	}
}

// Milli returns a uniform value in [0, One).
func (r *RNG) Milli() Milli { return Milli(r.IntN(int(One))) }

// Normal returns an approximately normal draw with mean 0 and the given
// standard deviation, via Irwin-Hall: the sum of 12 uniforms on [0,1) has
// mean 6 and variance 1, so subtracting 6 gives a unit normal.
//
// Chosen over Box-Muller because it needs no log, sqrt, or cos. Go's
// assembly implementations of those differ from the pure-Go ones used on
// js/wasm, and that difference is exactly what would break native/WASM
// parity. Draws are bounded to +/-6 sigma, which is a feature rather than a
// limitation: a season should not turn on an absurd tail outlier.
func (r *RNG) Normal(sigma Milli) Milli {
	var sum Milli
	for i := 0; i < 12; i++ {
		sum += r.Milli()
	}
	return (sum - 6*One).Mul(sigma)
}

// Chance reports whether an event with the given probability, expressed in
// Milli where One is certainty, occurs.
func (r *RNG) Chance(p Milli) bool {
	if p <= 0 {
		return false
	}
	if p >= One {
		return true
	}
	return r.Milli() < p
}
