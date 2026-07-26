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
func RunSeason(seed int64, picks []int) (SeasonResult, error) {
	if len(picks) != WindowCount {
		return SeasonResult{}, errors.New("sim: got " + strconv.Itoa(len(picks)) +
			" picks, want " + strconv.Itoa(WindowCount))
	}
	return RunPartial(seed, picks)
}

// RunPartial resolves only the races a prefix of picks unlocks.
//
// Races 1-2 depend on pick 1 alone, races 3-4 on picks 1-2, and so on,
// because the RNG advances strictly in round order. So after k picks the
// first 2k races are fully determined and can be shown before the next deal
// -- which is what lets the client give each window a consequence instead
// of dumping ten races at the end.
//
// A full-length picks slice makes this identical to RunSeason.
func RunPartial(seed int64, picks []int) (SeasonResult, error) {
	season := GenerateSeason(seed)
	deals := DealsFor(seed)

	if len(picks) > WindowCount {
		return SeasonResult{}, errors.New("sim: got " + strconv.Itoa(len(picks)) +
			" picks, want at most " + strconv.Itoa(WindowCount))
	}
	for i, p := range picks {
		if p < 0 || p >= DealSize {
			return SeasonResult{}, errors.New("sim: window " + strconv.Itoa(i+1) +
				" picked card " + strconv.Itoa(p) + ", want 0.." + strconv.Itoa(DealSize-1))
		}
	}

	// Rounds are unlocked two at a time, one pair per window.
	unlocked := len(picks) * (RaceCount / WindowCount)

	// A SECOND RNG, seeded separately from the one GenerateSeason consumed,
	// so that changing generation logic later cannot shift race outcomes.
	rng := NewRNG(seed ^ raceSalt)

	cars := make([]*carState, len(season.Teams))
	for i, t := range season.Teams {
		cars[i] = newCarState(t)
	}

	standings := make([]Standing, len(season.Teams))
	for i, t := range season.Teams {
		standings[i] = Standing{TeamID: t.ID, Name: t.Name}
	}

	build := make([]Card, 0, WindowCount)

	races := make([]RaceResult, 0, unlocked)
	for round, circuit := range season.Calendar {
		if round >= unlocked {
			break
		}
		// Development is banked only at window rounds, in team ID order:
		// car 0 is the player, every other car is its archetype's
		// deterministic choice from the same deal.
		if w, ok := windowAt(round); ok {
			chosen := deals[w][picks[w]]
			cars[0].apply(chosen.Effect)
			build = append(build, chosen)
			for i := 1; i < len(cars); i++ {
				j := rivalPick(cars[i].team, deals[w], w, season.Calendar)
				cars[i].apply(deals[w][j].Effect)
			}
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
		Build:      build,
		Races:      races,
		Standings:  standings,
		Player:     player,
		PlayerPos:  playerPos,
		Share:      shareString(seed, player, playerPos, races),
	}, nil
}

// windowAt reports whether a development window precedes this round, and
// which window it is.
func windowAt(round int) (int, bool) {
	for w, r := range WindowRounds {
		if r == round {
			return w, true
		}
	}
	return 0, false
}
