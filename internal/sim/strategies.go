package sim

// Strategy returns the scripted player picks for a named strategy, or nil
// if the name is unknown. These are the player-side counterparts to the
// rival archetypes: the balance harness runs all of them against the same
// seeds to check that no single approach dominates, and the simulate CLI
// uses them to play a season without a UI.
//
// They are ordered by how much they understand about the game, and M1
// measures the gradient across that ordering.
func Strategy(name string, season Season) []int {
	deals := DealsFor(season.Seed)
	picks := make([]int, WindowCount)

	// aeroCapUnits is the aero spend at which the multiplier saturates.
	// Past it, aero buys only its circuit-weight contribution.
	aeroCapUnits := int(MaxAeroBonus.Div(AeroScale) / GainPerUnit)
	aeroSpent := 0

	for w := 0; w < WindowCount; w++ {
		deal := deals[w]
		switch name {
		case "greedy":
			// Buys the most development available, every time, and eats
			// the compounding risk that comes with it.
			for i := 1; i < DealSize; i++ {
				if deal[i].Cost() > deal[picks[w]].Cost() {
					picks[w] = i
				}
			}
		case "cautious":
			// Spends as little as the deal allows. The floor a real
			// strategy has to beat.
			for i := 1; i < DealSize; i++ {
				if deal[i].Cost() < deal[picks[w]].Cost() {
					picks[w] = i
				}
			}
		case "aerofirst":
			// Exploits the aero multiplier: buy aero until the cap, where
			// it compounds across the most remaining races, then read the
			// calendar. The deepest insight the game rewards.
			if aeroSpent < aeroCapUnits {
				for i := 1; i < DealSize; i++ {
					if deal[i].Effect.Aero > deal[picks[w]].Effect.Aero {
						picks[w] = i
					}
				}
				aeroSpent += deal[picks[w]].Effect.Aero
				continue
			}
			wc, we, wa := weightsForWindow(w, season.Calendar)
			bestFit := cardFit(deal[0], wc, we, wa)
			for i := 1; i < DealSize; i++ {
				if f := cardFit(deal[i], wc, we, wa); f > bestFit {
					picks[w], bestFit = i, f
				}
			}
		case "adaptive":
			// What a thoughtful human does: read the races this window
			// precedes and take the card that best suits them.
			wc, we, wa := weightsForWindow(w, season.Calendar)
			bestFit := cardFit(deal[0], wc, we, wa)
			for i := 1; i < DealSize; i++ {
				if f := cardFit(deal[i], wc, we, wa); f > bestFit {
					picks[w], bestFit = i, f
				}
			}
		case "first":
			// Always takes the first card offered. The no-thought baseline
			// that replaces the old "spend nothing", which the card draft
			// makes inexpressible: you must take something.
			picks[w] = 0
		default:
			return nil
		}
	}
	return picks
}
