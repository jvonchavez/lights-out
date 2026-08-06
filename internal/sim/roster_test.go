package sim

import "testing"

// The roster is content, so it is validated by test rather than by a
// parser -- the same choice cards_test.go makes for the card pool. These
// tests are the only thing standing between a typo in a rating and a
// silently unbalanced season.

func TestRosterIDsAreUnique(t *testing.T) {
	seenTeamEra := map[string]bool{}
	seenEntity := map[string]bool{}
	seenTeamYear := map[string]bool{}

	for _, te := range Roster {
		if seenTeamEra[te.ID] {
			t.Errorf("duplicate team-era ID %q", te.ID)
		}
		seenTeamEra[te.ID] = true

		ty := te.Team + "/" + itoa(te.Year)
		if seenTeamYear[ty] {
			t.Errorf("duplicate team-year %q", ty)
		}
		seenTeamYear[ty] = true

		// Every person and car ID is suffixed with its team-era's year, so
		// Newey at Williams in 1992 and Newey at Aston Martin in 2026 are
		// distinct entities that can carry different ratings.
		ids := []string{te.Car.ID, te.Engineer.ID, te.Principal.ID,
			te.Drivers[0].ID, te.Drivers[1].ID}
		for _, id := range ids {
			if id == "" {
				t.Errorf("%s: empty entity ID", te.ID)
			}
			if seenEntity[id] {
				t.Errorf("%s: duplicate entity ID %q", te.ID, id)
			}
			seenEntity[id] = true
		}
	}
}

func TestRosterRatingsAreInRange(t *testing.T) {
	// 40 is the floor because a rating maps straight into the performance
	// sum: anything lower would put a car so far off the back that the race
	// stops being a race. 99 is the ceiling because these are sports-game
	// ratings and nothing is perfect.
	check := func(t *testing.T, what string, vals ...int) {
		t.Helper()
		for _, v := range vals {
			if v < 40 || v > 99 {
				t.Errorf("%s: rating %d out of range 40..99", what, v)
			}
		}
	}
	for _, te := range Roster {
		check(t, te.ID+" car", te.Car.Power, te.Car.Cornering, te.Car.Aero, te.Car.Reliability)
		check(t, te.ID+" engineer", te.Engineer.Setup, te.Engineer.Strategy, te.Engineer.Ops)
		check(t, te.ID+" principal", te.Principal.Development, te.Principal.Leadership, te.Principal.Nerve)
		for _, d := range te.Drivers {
			check(t, te.ID+" "+d.ID, d.Pace, d.Racecraft, d.Consistency, d.Composure)
		}
	}
}

func TestRosterOverallIsDerivedAndInRange(t *testing.T) {
	for _, te := range Roster {
		overalls := map[string]int{
			"car":       te.Car.Overall(),
			"engineer":  te.Engineer.Overall(),
			"principal": te.Principal.Overall(),
			"driver0":   te.Drivers[0].Overall(),
			"driver1":   te.Drivers[1].Overall(),
		}
		for what, ov := range overalls {
			if ov < 40 || ov > 99 {
				t.Errorf("%s %s: Overall %d out of range 40..99", te.ID, what, ov)
			}
		}
	}
}

func TestRosterNamesAreNonEmpty(t *testing.T) {
	for _, te := range Roster {
		if te.Team == "" || te.Car.Name == "" || te.Engineer.Name == "" || te.Principal.Name == "" {
			t.Errorf("%s: empty display name", te.ID)
		}
		if te.Livery == "" || te.Livery[0] != '#' || len(te.Livery) != 7 {
			t.Errorf("%s: livery %q is not a 7-character hex colour", te.ID, te.Livery)
		}
		for _, d := range te.Drivers {
			if d.Name == "" {
				t.Errorf("%s: empty driver name", te.ID)
			}
		}
	}
}

func TestRosterEraIDsAreKnown(t *testing.T) {
	known := map[string]bool{}
	for _, e := range Eras {
		known[e.ID] = true
	}
	covered := map[string]int{}
	for _, te := range Roster {
		if !known[te.EraID] {
			t.Errorf("%s: unknown era %q", te.ID, te.EraID)
		}
		covered[te.EraID]++
	}
	for _, e := range Eras {
		if covered[e.ID] == 0 {
			t.Errorf("era %q has no team-eras", e.ID)
		}
	}
}

// The 2026 grid is the field, so its size is not a content choice -- it is
// the number of rival teams the season simulates.
func TestGrid2026IsTheRivalField(t *testing.T) {
	const want = 11
	if len(Grid2026) != want {
		t.Fatalf("Grid2026 has %d teams, want %d", len(Grid2026), want)
	}
	for _, te := range Grid2026 {
		if te.EraID != "2026" {
			t.Errorf("%s: grid entry in era %q", te.ID, te.EraID)
		}
	}
}

// Grid2026 is appended to Roster, so the current grid is rollable like any
// other team-era. Taking McLaren's car means racing a McLaren with a hole
// in it, which is the point.
func TestGrid2026IsRollable(t *testing.T) {
	inRoster := map[string]bool{}
	for _, te := range Roster {
		inRoster[te.ID] = true
	}
	for _, te := range Grid2026 {
		if !inRoster[te.ID] {
			t.Errorf("%s is not in Roster and so can never be rolled", te.ID)
		}
	}
}

// A roster where everything is excellent is a roster with no decisions in
// it. The spread is the content equivalent of the balance sweep.
func TestRosterHasSpread(t *testing.T) {
	lo, hi, sum := 99, 0, 0
	for _, te := range Roster {
		ov := te.Car.Overall()
		if ov < lo {
			lo = ov
		}
		if ov > hi {
			hi = ov
		}
		sum += ov
	}
	if hi-lo < 25 {
		t.Errorf("car Overall spans only %d points (%d..%d); the roster is too uniform", hi-lo, lo, hi)
	}
	mean := sum / len(Roster)
	if mean < 70 || mean > 88 {
		t.Errorf("mean car Overall is %d, want 70..88", mean)
	}
}

// The design claim that eras differ by reliability rather than by pace is
// the reason a 1988 car is worth taking at all. If this test fails the
// trade-off the roster is built on has quietly gone away.
func TestReliabilityRisesAcrossEras(t *testing.T) {
	mean := func(eraID string) int {
		sum, n := 0, 0
		for _, te := range Roster {
			if te.EraID == eraID {
				sum += te.Car.Reliability
				n++
			}
		}
		if n == 0 {
			t.Fatalf("no entries for era %q", eraID)
		}
		return sum / n
	}
	early, late := mean("1950s"), mean("2020s")
	if early >= late {
		t.Errorf("mean reliability: 1950s %d >= 2020s %d; the era trade-off is gone", early, late)
	}
	if late-early < 20 {
		t.Errorf("mean reliability gap 1950s..2020s is only %d points, want at least 20", late-early)
	}
}

func TestTeamEraLabel(t *testing.T) {
	te := TeamEra{Team: "McLaren", Year: 1988}
	if got := te.Label(); got != "1988 McLaren" {
		t.Errorf("Label() = %q, want %q", got, "1988 McLaren")
	}
}
