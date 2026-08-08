package sim

import "testing"

// baselineCar and friends are deliberately flat: every rating sits at a
// mid value and the principal sits exactly at DevBaseline so development
// contributes nothing. Tests that care about ordering rather than about any
// particular car being fast build their field from these.
func baselineCar() CarSpec {
	return CarSpec{ID: "base", Name: "Base", Power: 75, Cornering: 75, Aero: 75, Reliability: 85}
}

func baselineDriver() DriverSpec {
	return DriverSpec{ID: "d", Name: "D", Pace: 75, Racecraft: 75,
		Consistency: ConsistencyBaseline, Composure: 80}
}

func testTeam(id int) Team {
	return Team{ID: id, Name: "T", Lineup: Lineup{
		Car:       baselineCar(),
		Drivers:   [2]DriverSpec{baselineDriver(), baselineDriver()},
		Engineer:  EngineerSpec{ID: "e", Name: "E", Setup: 75, Strategy: 75, Ops: 75},
		Principal: PrincipalSpec{ID: "p", Name: "P", Development: DevBaseline, Leadership: 75, Nerve: 75},
	}}
}

// fieldOf builds n teams of CarsPerTeam cars each, in field order.
func fieldOf(n int) []*entryState {
	field := make([]*entryState, 0, n*CarsPerTeam)
	for i := 0; i < n; i++ {
		for _, e := range entriesFor(testTeam(i)) {
			field = append(field, e)
		}
	}
	return field
}

func balanced() Circuit {
	return Circuit{Name: "T", Archetype: "balanced", Profile: Profiles["balanced"]}
}

func TestQualifyingReturnsEveryCarOnce(t *testing.T) {
	field := fieldOf(TeamCount)
	grid := qualify(field, 0, Profiles["balanced"], NewRNG(1))
	if len(grid) != FieldSize {
		t.Fatalf("grid has %d entries, want %d", len(grid), FieldSize)
	}
	seen := map[gridKey]bool{}
	for _, k := range grid {
		if seen[k] {
			t.Errorf("car %v appears twice on the grid", k)
		}
		seen[k] = true
	}
}

func TestQualifyingFavoursTheFasterCar(t *testing.T) {
	poles := 0
	for seed := int64(0); seed < 100; seed++ {
		field := fieldOf(TeamCount)
		// Give team 5 a commanding advantage in every area.
		for _, e := range field {
			if e.teamID == 5 {
				e.car = CarSpec{ID: "fast", Name: "Fast", Power: 99, Cornering: 99, Aero: 99, Reliability: 99}
			}
		}
		if qualify(field, 0, Profiles["balanced"], NewRNG(seed))[0].teamID == 5 {
			poles++
		}
	}
	if poles < 95 {
		t.Errorf("dominant car took pole %d/100 times, want at least 95", poles)
	}
}

func TestDNFScoresZeroAndIsNotClassified(t *testing.T) {
	field := fieldOf(TeamCount)
	// A 1950s-grade car with a poor operation: high failure, guaranteed to
	// bite somewhere across 200 races.
	for _, e := range field {
		if e.teamID == 3 {
			e.car.Reliability = 40
			e.engineer.Ops = 60
		}
	}
	found := false
	for seed := int64(0); seed < 200 && !found; seed++ {
		grid := qualify(field, 0, Profiles["balanced"], NewRNG(seed))
		res := runRace(field, 0, balanced(), grid, NewRNG(seed))
		for _, c := range res.Entries {
			if c.TeamID == 3 && c.DNF {
				found = true
				if c.Points != 0 {
					t.Errorf("DNF scored %d points, want 0", c.Points)
				}
				if c.Finish != 0 {
					t.Errorf("DNF has finish position %d, want 0", c.Finish)
				}
				if c.DNFReason != DNFMechanical && c.DNFReason != DNFDriver {
					t.Errorf("DNF reason %q, want a named cause", c.DNFReason)
				}
			}
		}
	}
	if !found {
		t.Fatal("an unreliable car never DNF'd across 200 races")
	}
}

// A finisher must never carry a retirement reason, and a retirement must
// always carry one. The reel reads this field directly.
func TestDNFReasonMatchesDNF(t *testing.T) {
	field := fieldOf(TeamCount)
	for seed := int64(0); seed < 50; seed++ {
		grid := qualify(field, 0, Profiles["balanced"], NewRNG(seed))
		for _, c := range runRace(field, 0, balanced(), grid, NewRNG(seed)).Entries {
			if c.DNF && c.DNFReason == "" {
				t.Errorf("seed %d: DNF with no reason", seed)
			}
			if !c.DNF && c.DNFReason != "" {
				t.Errorf("seed %d: finisher carries reason %q", seed, c.DNFReason)
			}
		}
	}
}

func TestPointsMatchTable(t *testing.T) {
	field := fieldOf(TeamCount)
	grid := qualify(field, 0, Profiles["balanced"], NewRNG(42))
	res := runRace(field, 0, balanced(), grid, NewRNG(42))

	byFinish := map[int]EntryResult{}
	for _, c := range res.Entries {
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
	if c, ok := byFinish[len(PointsTable)+1]; ok && c.Points != 0 {
		t.Errorf("P%d scored %d, want 0", len(PointsTable)+1, c.Points)
	}
}

func TestEntriesAlwaysSortedByTeamThenEntry(t *testing.T) {
	field := fieldOf(TeamCount)
	grid := qualify(field, 0, Profiles["power"], NewRNG(8))
	c := Circuit{Name: "T", Archetype: "power", Profile: Profiles["power"]}
	res := runRace(field, 0, c, grid, NewRNG(8))
	for i, e := range res.Entries {
		wantTeam, wantEntry := i/CarsPerTeam, i%CarsPerTeam
		if e.TeamID != wantTeam || e.Entry != wantEntry {
			t.Fatalf("Entries[%d] is team %d entry %d, want %d/%d -- results must be sorted for stable JSON",
				i, e.TeamID, e.Entry, wantTeam, wantEntry)
		}
	}
}

func TestEveryCarIsAccountedFor(t *testing.T) {
	field := fieldOf(TeamCount)
	c := Circuit{Name: "T", Archetype: "technical", Profile: Profiles["technical"]}
	grid := qualify(field, 0, c.Profile, NewRNG(3))
	res := runRace(field, 0, c, grid, NewRNG(3))
	if len(res.Entries) != FieldSize {
		t.Fatalf("classified %d cars, want %d", len(res.Entries), FieldSize)
	}
	finishes := map[int]bool{}
	for _, e := range res.Entries {
		if e.DNF {
			continue
		}
		if finishes[e.Finish] {
			t.Errorf("two cars share P%d", e.Finish)
		}
		finishes[e.Finish] = true
	}
}

func TestOvertakingIsHarderAtTechnicalCircuits(t *testing.T) {
	beaten := func(arch string) int {
		n := 0
		for seed := int64(0); seed < 400; seed++ {
			field := fieldOf(TeamCount)
			c := Circuit{Name: arch, Archetype: arch, Profile: Profiles[arch]}
			grid := qualify(field, 0, c.Profile, NewRNG(seed))
			res := runRace(field, 0, c, grid, NewRNG(seed+7777))
			for _, r := range res.Entries {
				if r.TeamID == grid[0].teamID && r.Entry == grid[0].entry && !r.DNF && r.Finish > 1 {
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

// Racecraft buys back a share of the grid penalty, so a driver who can
// overtake should recover from the back more often than one who cannot.
// This is the whole reason racecraft is a separate rating from pace.
func TestRacecraftRecoversFromABadGrid(t *testing.T) {
	p := Profiles["technical"]
	slow := &entryState{driver: DriverSpec{Racecraft: 60}}
	quick := &entryState{driver: DriverSpec{Racecraft: 97}}
	if quick.gridPenalty(20, p) >= slow.gridPenalty(20, p) {
		t.Errorf("racecraft 97 penalty %d not below racecraft 60 penalty %d",
			quick.gridPenalty(20, p), slow.gridPenalty(20, p))
	}
	// And it is capped: nobody starts last for free.
	if quick.gridPenalty(20, p) <= 0 {
		t.Error("grid penalty vanished entirely; MaxGridRelief is not holding")
	}
}

func TestSafetyCarFiresAtTheExpectedRate(t *testing.T) {
	fired := 0
	const n = 2000
	for seed := int64(0); seed < n; seed++ {
		field := fieldOf(TeamCount)
		grid := qualify(field, 0, Profiles["balanced"], NewRNG(seed))
		if runRace(field, 0, balanced(), grid, NewRNG(seed)).SafetyCar {
			fired++
		}
	}
	rate := Milli(fired * int(One) / n)
	if rate < SafetyCarChance-60 || rate > SafetyCarChance+60 {
		t.Errorf("safety car rate %d, want near %d", rate, SafetyCarChance)
	}
}
