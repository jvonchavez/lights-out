package sim

import "testing"

func testCalendar(archetypes ...string) []Circuit {
	cal := make([]Circuit, len(archetypes))
	for i, a := range archetypes {
		cal[i] = Circuit{Name: a, Archetype: a, Profile: Profiles[a]}
	}
	return cal
}

func TestRivalDecisionsSpendExactlyTheBudget(t *testing.T) {
	cal := GenerateSeason(1).Calendar
	for _, arch := range rivalArchetypes {
		for round := 0; round < RaceCount; round++ {
			tm := Team{ID: 1, Archetype: arch}
			d := rivalDecision(tm, round, cal, RaceBudget)
			if d.Total() != RaceBudget {
				t.Errorf("%s round %d spent %d, want exactly %d", arch, round, d.Total(), RaceBudget)
			}
			if d.Chassis < 0 || d.Engine < 0 || d.Aero < 0 {
				t.Errorf("%s round %d has a negative allocation: %+v", arch, round, d)
			}
		}
	}
}

func TestRivalDecisionsAreDeterministic(t *testing.T) {
	cal := GenerateSeason(2).Calendar
	for _, arch := range rivalArchetypes {
		tm := Team{ID: 3, Archetype: arch}
		a := rivalDecision(tm, 4, cal, RaceBudget)
		b := rivalDecision(tm, 4, cal, RaceBudget)
		if a != b {
			t.Errorf("%s is not deterministic: %+v vs %+v", arch, a, b)
		}
	}
}

func TestAggressiveNeglectsAero(t *testing.T) {
	// The aggressive archetype buys raw chassis and engine and is blind to
	// the fact that aero multiplies the whole car. That blindness is the
	// point: it is a high-risk, high-ceiling bet that M1 measured as poor.
	cal := GenerateSeason(3).Calendar
	d := rivalDecision(Team{ID: 1, Archetype: "aggressive"}, 0, cal, RaceBudget)
	if d.Aero >= d.Chassis || d.Aero >= d.Engine {
		t.Errorf("aggressive should underweight aero: %+v", d)
	}
}

func TestConservativeSpreadsEvenly(t *testing.T) {
	cal := GenerateSeason(4).Calendar
	d := rivalDecision(Team{ID: 2, Archetype: "conservative"}, 0, cal, RaceBudget)
	want := Decision{34, 33, 33}
	if d != want {
		t.Errorf("conservative = %+v, want %+v", d, want)
	}
}

func TestSpecialistConcentratesOnOneArea(t *testing.T) {
	cal := GenerateSeason(5).Calendar
	seen := map[int]bool{}
	for id := 1; id <= 10; id++ {
		d := rivalDecision(Team{ID: id, Archetype: "specialist"}, 0, cal, RaceBudget)
		perf := []int{d.Chassis, d.Engine, d.Aero}
		nonZero := 0
		for i, v := range perf {
			if v > 0 {
				nonZero++
				seen[i] = true
			}
		}
		if nonZero != 1 {
			t.Errorf("team %d specialist spread across %d areas, want 1: %+v", id, nonZero, d)
		}
	}
	if len(seen) < 3 {
		t.Errorf("specialists only ever chose %d distinct areas, want 3", len(seen))
	}
}

func TestReactiveFollowsTheCalendar(t *testing.T) {
	// Rounds 2 and 3 (0-based) are power circuits, so the round-1 decision
	// looking two ahead should favour engine over chassis.
	cal := testCalendar("balanced", "balanced", "power", "power", "technical",
		"technical", "balanced", "balanced", "power", "technical")
	d := rivalDecision(Team{ID: 4, Archetype: "reactive"}, 2, cal, RaceBudget)
	if d.Engine <= d.Chassis {
		t.Errorf("reactive at power circuits: engine %d, chassis %d -- want engine higher", d.Engine, d.Chassis)
	}

	// Mirror: technical circuits ahead should favour chassis.
	d2 := rivalDecision(Team{ID: 4, Archetype: "reactive"}, 4, cal, RaceBudget)
	if d2.Chassis <= d2.Engine {
		t.Errorf("reactive at technical circuits: chassis %d, engine %d -- want chassis higher", d2.Chassis, d2.Engine)
	}
}

func TestReactiveHandlesTheFinalRound(t *testing.T) {
	cal := GenerateSeason(6).Calendar
	d := rivalDecision(Team{ID: 4, Archetype: "reactive"}, RaceCount-1, cal, RaceBudget)
	if d.Total() != RaceBudget {
		t.Errorf("reactive final round spent %d, want %d", d.Total(), RaceBudget)
	}
}
