package sim

import (
	"cmp"
	"errors"
	"slices"
	"strconv"
)

// RunSeason plays an entire season and returns the result. It is the one
// exported entry point of this package, and the same function the browser
// calls through WASM and the server calls natively to verify a submission.
//
// It is a pure function of (seed, decisions): no wall-clock time, no map
// iteration order affecting results, no unseeded randomness. That is what
// makes the leaderboard trustworthy -- a cheater would have to find
// decisions that genuinely produce a high score, which is just playing well.
func RunSeason(seed int64, decisions []Decision) (SeasonResult, error) {
	season := GenerateSeason(seed)

	if len(decisions) != len(season.Calendar) {
		return SeasonResult{}, errors.New("sim: got " + strconv.Itoa(len(decisions)) +
			" decisions, want " + strconv.Itoa(len(season.Calendar)))
	}
	for i, d := range decisions {
		if d.Chassis < 0 || d.Engine < 0 || d.Aero < 0 || d.Reliability < 0 {
			return SeasonResult{}, errors.New("sim: race " + strconv.Itoa(i+1) +
				" has a negative allocation")
		}
		if d.Total() > season.Budgets[i] {
			return SeasonResult{}, errors.New("sim: race " + strconv.Itoa(i+1) +
				" allocates " + strconv.Itoa(d.Total()) +
				", budget is " + strconv.Itoa(season.Budgets[i]))
		}
	}

	// A SECOND RNG, seeded separately from the one GenerateSeason consumed,
	// so that changing generation logic later cannot shift race outcomes.
	rng := NewRNG(seed ^ 0x5EED)

	cars := make([]*carState, len(season.Teams))
	for i, t := range season.Teams {
		cars[i] = newCarState(t)
	}

	standings := make([]Standing, len(season.Teams))
	for i, t := range season.Teams {
		standings[i] = Standing{TeamID: t.ID, Name: t.Name}
	}

	races := make([]RaceResult, 0, len(season.Calendar))
	for round, circuit := range season.Calendar {
		// Development is banked in team ID order: car 0 is the player,
		// every other car is its archetype's deterministic choice.
		cars[0].apply(decisions[round])
		for i := 1; i < len(cars); i++ {
			cars[i].apply(rivalDecision(cars[i].team, round, season.Calendar, season.Budgets[round]))
		}

		grid := qualify(cars, circuit.Profile, rng)
		res := runRace(cars, circuit, grid, rng)
		res.Round = round + 1
		races = append(races, res)

		for _, c := range res.Cars {
			s := &standings[c.TeamID]
			s.Points += c.Points
			switch {
			case c.DNF:
				s.DNFs++
			case c.Finish == 1:
				s.Wins++
				s.Podiums++
			case c.Finish <= 3:
				s.Podiums++
			}
		}
	}

	player := standings[0]

	slices.SortFunc(standings, func(a, b Standing) int {
		if a.Points != b.Points {
			return cmp.Compare(b.Points, a.Points)
		}
		if a.Wins != b.Wins {
			return cmp.Compare(b.Wins, a.Wins)
		}
		if a.Podiums != b.Podiums {
			return cmp.Compare(b.Podiums, a.Podiums)
		}
		return cmp.Compare(a.TeamID, b.TeamID)
	})

	playerPos := 0
	for i, s := range standings {
		if s.TeamID == 0 {
			playerPos = i + 1
			break
		}
	}

	return SeasonResult{
		SimVersion: Version,
		Seed:       seed,
		Races:      races,
		Standings:  standings,
		Player:     player,
		PlayerPos:  playerPos,
		Share:      shareString(seed, player, playerPos, races),
	}, nil
}
