package sim

// dealSalt keeps card deals on their own RNG stream, independent of season
// generation (seeded with seed) and race resolution (seed ^ raceSalt).
// Three separate streams mean editing the card pool cannot shift race
// outcomes, and changing generation cannot shift deals. Draw order is the
// determinism contract, and this is how it stays local.
const (
	dealSalt = 0xCA4D
	raceSalt = 0x5EED
)

// DealsFor returns the cards offered at each development window. It is a
// pure function of the seed, so every player in the world is dealt the same
// hand on the same day and the server can re-derive it to verify a pick.
func DealsFor(seed int64) [WindowCount][DealSize]Card {
	rng := NewRNG(seed ^ dealSalt)

	// Copy before shuffling: the package-level pool must never be mutated,
	// or one call would change the next.
	pool := make([]Card, len(CardPool))
	copy(pool, CardPool)

	var deals [WindowCount][DealSize]Card
	for w := 0; w < WindowCount; w++ {
		// A partial Fisher-Yates over a freshly reshuffled prefix: each
		// window draws DealSize distinct cards, and windows may repeat a
		// card across the season, which keeps the pool feeling deep
		// without ever offering the same part twice in one choice.
		for i := 0; i < DealSize; i++ {
			j := i + rng.IntN(len(pool)-i)
			pool[i], pool[j] = pool[j], pool[i]
			deals[w][i] = pool[i]
		}
	}
	return deals
}
