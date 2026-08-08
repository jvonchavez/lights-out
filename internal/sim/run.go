package sim

import (
	"cmp"
	"slices"
)

// PlayerTeamName is what the player's constructor is called in standings.
const PlayerTeamName = "Your Team"

// RunSeason plays an entire season and returns the result. It is the one
// exported entry point of this package, and the same function the browser
// calls through WASM and the server calls natively to verify a submission.
//
// It is a pure function of (seed, picks): no wall-clock time, no map
// iteration order affecting results, no unseeded randomness. The server is
// the sole authority on what a set of picks is worth -- it re-derives the
// rolls from the seed and replays the picks itself, so scores cannot be
// fabricated. They can be searched for; docs/_README.md says so plainly.
//
// RunPartial is gone. It existed to show the two races each in-season
// window unlocked, and there are no in-season decisions left to give
// feedback on: the team is locked in before round one and the whole season
// resolves at once. The race-by-race reveal is now presentation, done in
// the client against a result it already holds.
func RunSeason(seed int64, picks []int) (SeasonResult, error) {
	season := GenerateSeason(seed)

	lineup, err := BuildLineup(season.Rolls, picks)
	if err != nil {
		return SeasonResult{}, err
	}

	teams := make([]Team, 0, TeamCount)
	teams = append(teams, Team{ID: 0, Name: PlayerTeamName, Livery: "#E10600", Lineup: lineup})
	teams = append(teams, season.Rivals...)

	// The field is built in team ID order, then entry order within a team.
	// That order is the determinism contract for every draw below.
	field := make([]*entryState, 0, FieldSize)
	for _, t := range teams {
		for _, e := range entriesFor(t) {
			field = append(field, e)
		}
	}

	// A SECOND RNG, seeded separately from the one GenerateSeason consumed,
	// so that changing generation logic later cannot shift race outcomes.
	rng := NewRNG(seed ^ raceSalt)

	standings := make([]Standing, len(teams))
	for i, t := range teams {
		standings[i] = Standing{TeamID: t.ID, Name: t.Name}
	}

	// Drivers are keyed by team and entry rather than by driver ID: the
	// same driver can legitimately appear twice on the grid, once in your
	// team and once in the real one you took them from.
	type driverKey struct{ teamID, entry int }
	driverPoints := map[driverKey]*DriverStanding{}
	driverOrder := make([]driverKey, 0, FieldSize)
	for _, e := range field {
		k := driverKey{e.teamID, e.entry}
		driverPoints[k] = &DriverStanding{
			DriverID: e.driver.ID,
			Name:     e.driver.Name,
			TeamID:   e.teamID,
		}
		driverOrder = append(driverOrder, k)
	}

	races := make([]RaceResult, 0, RaceCount)
	for round, circuit := range season.Calendar {
		grid := qualify(field, round, circuit.Profile, rng)
		res := runRace(field, round, circuit, grid, rng)
		res.Round = round + 1
		races = append(races, res)

		for _, c := range res.Entries {
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

			d := driverPoints[driverKey{c.TeamID, c.Entry}]
			d.Points += c.Points
			if c.Finish == 1 {
				d.Wins++
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

	drivers := make([]DriverStanding, 0, FieldSize)
	for _, k := range driverOrder {
		drivers = append(drivers, *driverPoints[k])
	}
	slices.SortFunc(drivers, func(a, b DriverStanding) int {
		if a.Points != b.Points {
			return cmp.Compare(b.Points, a.Points)
		}
		if a.Wins != b.Wins {
			return cmp.Compare(b.Wins, a.Wins)
		}
		if a.TeamID != b.TeamID {
			return cmp.Compare(a.TeamID, b.TeamID)
		}
		return cmp.Compare(a.DriverID, b.DriverID)
	})

	out := make([]int, len(picks))
	copy(out, picks)

	return SeasonResult{
		SimVersion: Version,
		Seed:       seed,
		Rolls:      season.Rolls,
		Picks:      out,
		Lineup:     lineup,
		Races:      races,
		Standings:  standings,
		Drivers:    drivers,
		Player:     player,
		PlayerPos:  playerPos,
		Share:      shareString(seed, player, playerPos, races),
	}, nil
}
