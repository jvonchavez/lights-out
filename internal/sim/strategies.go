package sim

// Strategy returns the scripted player allocations for a named strategy, or
// nil if the name is unknown. These are the player-side counterparts to the
// rival archetypes: the balance harness runs all of them against the same
// seeds to check that no single approach dominates, and the simulate CLI
// uses them to play a season without a UI.
func Strategy(name string, season Season) []Decision {
	ds := make([]Decision, len(season.Budgets))
	for i, budget := range season.Budgets {
		switch name {
		case "even":
			ds[i] = split(budget, 25, 25, 25, 25)
		case "aggressive":
			// Everything into performance for the first three rounds, then
			// nothing at all. Development compounds, so early spending pays
			// off across more remaining races -- at maximum risk.
			if i < 3 {
				ds[i] = split(budget, 40, 40, 20, 0)
			}
		case "specialist":
			ds[i] = split(budget, 0, 80, 0, 20)
		case "reliability":
			ds[i] = split(budget, 20, 20, 20, 40)
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
