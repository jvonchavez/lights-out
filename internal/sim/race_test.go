package sim

import "testing"

// fieldOf builds n cars all on the baseline, for tests that care about
// ordering rather than about any particular car being fast.
func fieldOf(n int) []*carState {
	cars := make([]*carState, n)
	for i := range cars {
		cars[i] = newCarState(Team{ID: i, Start: Ratings{StartRating, StartRating, StartRating}, DriverSkill: 2 * One})
	}
	return cars
}

func TestQualifyingReturnsEveryCarOnce(t *testing.T) {
	cars := fieldOf(TeamCount)
	grid := qualify(cars, Profiles["balanced"], NewRNG(1))
	if len(grid) != TeamCount {
		t.Fatalf("grid has %d entries, want %d", len(grid), TeamCount)
	}
	seen := map[int]bool{}
	for _, id := range grid {
		if seen[id] {
			t.Errorf("team %d appears twice on the grid", id)
		}
		seen[id] = true
	}
}

func TestQualifyingFavoursTheFasterCar(t *testing.T) {
	poles := 0
	for seed := int64(0); seed < 100; seed++ {
		cars := fieldOf(TeamCount)
		// Give team 5 a commanding advantage: +30.0 in every area.
		for i := 0; i < 30; i++ {
			cars[5].apply(Decision{Chassis: 10, Engine: 10, Aero: 10})
		}
		if qualify(cars, Profiles["balanced"], NewRNG(seed))[0] == 5 {
			poles++
		}
	}
	if poles < 95 {
		t.Errorf("dominant car took pole %d/100 times, want at least 95", poles)
	}
}

func TestDNFScoresZeroAndIsNotClassified(t *testing.T) {
	cars := fieldOf(TeamCount)
	// Force certain failure for team 3 by spending absurdly on performance.
	for i := 0; i < 200; i++ {
		cars[3].apply(Decision{Engine: 100})
	}
	found := false
	for seed := int64(0); seed < 200 && !found; seed++ {
		grid := qualify(cars, Profiles["balanced"], NewRNG(seed))
		res := runRace(cars, Circuit{Name: "T", Archetype: "balanced", Profile: Profiles["balanced"]}, grid, NewRNG(seed))
		for _, c := range res.Cars {
			if c.TeamID == 3 && c.DNF {
				found = true
				if c.Points != 0 {
					t.Errorf("DNF scored %d points, want 0", c.Points)
				}
				if c.Finish != 0 {
					t.Errorf("DNF has finish position %d, want 0", c.Finish)
				}
			}
		}
	}
	if !found {
		t.Fatal("a car at the 60%% failure clamp never DNF'd across 200 races")
	}
}

func TestPointsMatchTable(t *testing.T) {
	cars := fieldOf(TeamCount)
	grid := qualify(cars, Profiles["balanced"], NewRNG(42))
	res := runRace(cars, Circuit{Name: "T", Archetype: "balanced", Profile: Profiles["balanced"]}, grid, NewRNG(42))

	byFinish := map[int]CarResult{}
	for _, c := range res.Cars {
		if !c.DNF {
			byFinish[c.Finish] = c
		}
	}
	for pos, want := range PointsTable {
		c, ok := byFinish[pos+1]
		if !ok {
			continue // that position DNF'd out of existence
		}
		if c.Points != want {
			t.Errorf("P%d scored %d, want %d", pos+1, c.Points, want)
		}
	}
	if c, ok := byFinish[11]; ok && c.Points != 0 {
		t.Errorf("P11 scored %d, want 0", c.Points)
	}
}

func TestCarsAlwaysSortedByTeamID(t *testing.T) {
	cars := fieldOf(TeamCount)
	grid := qualify(cars, Profiles["power"], NewRNG(8))
	res := runRace(cars, Circuit{Name: "T", Archetype: "power", Profile: Profiles["power"]}, grid, NewRNG(8))
	for i, c := range res.Cars {
		if c.TeamID != i {
			t.Fatalf("Cars[%d] has TeamID %d -- results must be sorted by team ID for stable JSON", i, c.TeamID)
		}
	}
}

func TestEveryCarIsAccountedFor(t *testing.T) {
	cars := fieldOf(TeamCount)
	grid := qualify(cars, Profiles["technical"], NewRNG(3))
	res := runRace(cars, Circuit{Name: "T", Archetype: "technical", Profile: Profiles["technical"]}, grid, NewRNG(3))
	if len(res.Cars) != TeamCount {
		t.Fatalf("classified %d cars, want %d", len(res.Cars), TeamCount)
	}
	finishes := map[int]bool{}
	for _, c := range res.Cars {
		if c.DNF {
			continue
		}
		if finishes[c.Finish] {
			t.Errorf("two cars share P%d", c.Finish)
		}
		finishes[c.Finish] = true
	}
}

func TestOvertakingIsHarderAtTechnicalCircuits(t *testing.T) {
	// Measure how often the pole-sitter is beaten. A power circuit's low
	// overtake difficulty should let the field through more often than a
	// technical circuit's high one.
	beaten := func(arch string) int {
		n := 0
		for seed := int64(0); seed < 400; seed++ {
			cars := fieldOf(TeamCount)
			c := Circuit{Name: arch, Archetype: arch, Profile: Profiles[arch]}
			grid := qualify(cars, c.Profile, NewRNG(seed))
			res := runRace(cars, c, grid, NewRNG(seed+7777))
			for _, r := range res.Cars {
				if r.TeamID == grid[0] && !r.DNF && r.Finish > 1 {
					n++
				}
			}
		}
		return n
	}
	power, technical := beaten("power"), beaten("technical")
	if power <= technical {
		t.Errorf("pole beaten %d times at power vs %d at technical -- power must be easier to overtake at", power, technical)
	}
}

func TestSafetyCarFiresAtTheExpectedRate(t *testing.T) {
	fired := 0
	const n = 2000
	for seed := int64(0); seed < n; seed++ {
		cars := fieldOf(TeamCount)
		c := Circuit{Name: "T", Archetype: "balanced", Profile: Profiles["balanced"]}
		grid := qualify(cars, c.Profile, NewRNG(seed))
		if runRace(cars, c, grid, NewRNG(seed)).SafetyCar {
			fired++
		}
	}
	rate := Milli(fired * int(One) / n)
	if rate < SafetyCarChance-60 || rate > SafetyCarChance+60 {
		t.Errorf("safety car rate %d, want near %d", rate, SafetyCarChance)
	}
}
