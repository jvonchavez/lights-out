package sim

// Strategy returns the scripted player allocations for a named strategy, or
// nil if the name is unknown. These are the player-side counterparts to the
// rival archetypes: the balance harness runs all of them against the same
// seeds to check that no single approach dominates, and the simulate CLI
// uses them to play a season without a UI.
//
// They are ordered by how much they understand about the game, and M1
// measured championships across that gradient -- idle 0%, aggressive 0.3%,
// specialist 1.4%, even 7.5%, adaptive 11.2%, aerofirst 24.6%. Decisions
// matter enormously without determining the outcome.
func Strategy(name string, season Season) []Decision {
	ds := make([]Decision, len(season.Budgets))

	// aeroCapUnits is the aero spend at which the multiplier saturates.
	// Past it, aero buys only its circuit-weight contribution.
	aeroCapUnits := int(MaxAeroBonus.Div(AeroScale) / GainPerUnit)
	aeroSpent := 0

	// weightsAhead sums the circuit weights over the next two rounds,
	// clamped at the end of the calendar.
	weightsAhead := func(i int) (wc, we, wa Milli) {
		for off := 0; off < 2; off++ {
			j := i + off
			if j >= len(season.Calendar) {
				j = len(season.Calendar) - 1
			}
			p := season.Calendar[j].Profile
			wc += p.Chassis
			we += p.Engine
			wa += p.Aero
		}
		return
	}

	for i, budget := range season.Budgets {
		switch name {
		case "even":
			ds[i] = split(budget, 34, 33, 33)
		case "aggressive":
			// Everything into chassis and engine for the first three
			// rounds, then nothing at all. Development compounds, so early
			// spending pays off across more remaining races -- at maximum
			// risk, and blind to aero.
			if i < 3 {
				ds[i] = split(budget, 50, 50, 0)
			}
		case "specialist":
			ds[i] = split(budget, 0, 100, 0)
		case "adaptive":
			// What a thoughtful human does: read the next two circuits off
			// the published calendar and invest toward them.
			wc, we, wa := weightsAhead(i)
			sum := wc + we + wa
			ds[i] = Decision{
				Chassis: budget * int(wc) / int(sum),
				Engine:  budget * int(we) / int(sum),
				Aero:    budget * int(wa) / int(sum),
			}
		case "aerofirst":
			// Exploits the aero multiplier: buy aero up to the cap early,
			// where it compounds across the most remaining races, then
			// spend the rest by the calendar. This is the deeper insight
			// the game rewards over naive calendar-following.
			if aeroSpent < aeroCapUnits {
				want := budget * 75 / 100
				if rem := aeroCapUnits - aeroSpent; want > rem {
					want = rem
				}
				aeroSpent += want
				rest := budget - want
				ds[i] = Decision{Aero: want, Chassis: rest / 2, Engine: rest - rest/2}
				continue
			}
			wc, we, _ := weightsAhead(i)
			c := budget * int(wc) / int(wc+we)
			ds[i] = Decision{Chassis: c, Engine: budget - c}
			continue
		case "idle":
			// Spends nothing. The floor every other strategy must beat.
		default:
			return nil
		}
		if rem := budget - ds[i].Total(); rem > 0 && name != "aggressive" && name != "idle" {
			ds[i].Chassis += rem
		}
	}
	return ds
}
