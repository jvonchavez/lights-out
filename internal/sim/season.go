package sim

// GenerateSeason derives an entire season descriptor from a seed. It is a
// pure function: the same seed always yields the same calendar, field, and
// budgets, which is what lets every player in the world face an identical
// season and makes the leaderboard a measure of decisions alone.
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

	// The player is drawn from the SAME distribution as the rivals, in team
	// ID order. This is a balance finding from M1, not an aesthetic choice:
	// a player pinned to the exact mean has a good average and no ceiling,
	// because winning a championship means being at the top of the field
	// and a car with zero variance never gets there. Measured at 100.5
	// average points but only 1.5% of titles, against a rival on the same
	// strategy scoring 102.9 for 11.2%.
	//
	// Fairness is preserved because the draw is seeded: every player in the
	// world gets the same car on the same day. Some days it is a good one.
	teams := make([]Team, 0, TeamCount)
	teams = append(teams, Team{
		ID:        0,
		Name:      "Your Team",
		Archetype: "",
		Start: Ratings{
			Chassis: StartRating + rng.Normal(StartSpread),
			Engine:  StartRating + rng.Normal(StartSpread),
			Aero:    StartRating + rng.Normal(StartSpread),
		},
		DriverSkill: 2*One + rng.Normal(One),
	})
	for i, name := range rivalNames {
		teams = append(teams, Team{
			ID:        i + 1,
			Name:      name,
			Archetype: rivalArchetypes[i%len(rivalArchetypes)],
			Start: Ratings{
				Chassis: StartRating + rng.Normal(StartSpread),
				Engine:  StartRating + rng.Normal(StartSpread),
				Aero:    StartRating + rng.Normal(StartSpread),
			},
			DriverSkill: 2*One + rng.Normal(One),
		})
	}

	// A slice rather than a constant so M1 can try a front-loaded or
	// back-loaded budget curve without changing any type.
	budgets := make([]int, RaceCount)
	for i := range budgets {
		budgets[i] = RaceBudget
	}

	return Season{
		Seed:       seed,
		SimVersion: Version,
		Calendar:   calendar,
		Teams:      teams,
		Budgets:    budgets,
	}
}
