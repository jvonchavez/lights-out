package sim

// GenerateSeason derives an entire season descriptor from a seed. It is a
// pure function: the same seed always yields the same calendar and the same
// five rolls, which is what lets every player in the world face an
// identical season and makes the leaderboard a measure of decisions alone.
//
// It no longer draws any ratings. The field is Grid2026 -- the real eleven
// teams, identical on every seed -- so the benchmark is constant and no
// seed is easier than another. The old model drew eleven procedural teams
// from a distribution and had to draw the player from the same one to give
// them a ceiling (an M1 finding); with a fixed field and a drafted player,
// that asymmetry cannot arise.
func GenerateSeason(seed int64) Season {
	rng := NewRNG(seed)

	// Copy the pool before shuffling; the package-level slice must not be
	// mutated, or one call would change the next.
	calendar := make([]Circuit, len(circuitPool))
	copy(calendar, circuitPool)
	for i := range calendar {
		calendar[i].Profile = Profiles[calendar[i].Archetype]
	}
	// Fisher-Yates, walking down so each index is chosen exactly once.
	for i := len(calendar) - 1; i > 0; i-- {
		j := rng.IntN(i + 1)
		calendar[i], calendar[j] = calendar[j], calendar[i]
	}

	return Season{
		Seed:       seed,
		SimVersion: Version,
		Calendar:   calendar,
		Rolls:      RollsFor(seed),
		Rivals:     rivalTeams(),
	}
}

// rivalTeams is the 2026 grid as entrants, at team IDs 1..TeamCount-1 in
// Grid2026 order. It takes no seed because it does not vary.
func rivalTeams() []Team {
	teams := make([]Team, 0, TeamCount-1)
	for i, te := range Grid2026 {
		teams = append(teams, Team{
			ID:     i + 1,
			Name:   te.Team,
			Livery: te.Livery,
			Lineup: Lineup{
				Car:       te.Car,
				Drivers:   te.Drivers,
				Engineer:  te.Engineer,
				Principal: te.Principal,
			},
		})
	}
	return teams
}
