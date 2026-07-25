package sim

import "testing"

func testCalendar(archetypes ...string) []Circuit {
	cal := make([]Circuit, len(archetypes))
	for i, a := range archetypes {
		cal[i] = Circuit{Name: a, Archetype: a, Profile: Profiles[a]}
	}
	return cal
}

// dealOf builds a deal with known shapes for testing archetype rules.
func dealOf(cards ...Card) [DealSize]Card {
	var d [DealSize]Card
	copy(d[:], cards)
	return d
}

var (
	cheapCard  = Card{ID: "cheap", Name: "Cheap", Blurb: "b", Effect: Decision{Chassis: 50, Engine: 50, Aero: 50}}
	richCard   = Card{ID: "rich", Name: "Rich", Blurb: "b", Effect: Decision{Chassis: 90, Engine: 90, Aero: 80}}
	engineCard = Card{ID: "eng", Name: "Engine", Blurb: "b", Effect: Decision{Engine: 250}}
	chassCard  = Card{ID: "cha", Name: "Chassis", Blurb: "b", Effect: Decision{Chassis: 250}}
	aeroCard   = Card{ID: "aer", Name: "Aero", Blurb: "b", Effect: Decision{Aero: 250}}
)

func TestRivalPicksAreInRange(t *testing.T) {
	cal := GenerateSeason(1).Calendar
	deals := DealsFor(1)
	for _, arch := range rivalArchetypes {
		for w := 0; w < WindowCount; w++ {
			tm := Team{ID: 1, Archetype: arch}
			got := rivalPick(tm, deals[w], w, cal)
			if got < 0 || got >= DealSize {
				t.Errorf("%s window %d picked %d, outside [0,%d)", arch, w, got, DealSize)
			}
		}
	}
}

func TestRivalPicksAreDeterministic(t *testing.T) {
	cal := GenerateSeason(2).Calendar
	deal := DealsFor(2)[1]
	for _, arch := range rivalArchetypes {
		tm := Team{ID: 3, Archetype: arch}
		if a, b := rivalPick(tm, deal, 1, cal), rivalPick(tm, deal, 1, cal); a != b {
			t.Errorf("%s is not deterministic: %d vs %d", arch, a, b)
		}
	}
}

func TestAggressiveTakesTheCostliestCard(t *testing.T) {
	cal := GenerateSeason(3).Calendar
	deal := dealOf(cheapCard, richCard, engineCard) // costs 150, 260, 250
	got := rivalPick(Team{ID: 1, Archetype: "aggressive"}, deal, 0, cal)
	if deal[got].Cost() != 260 {
		t.Errorf("aggressive took %s at %d units, want the 260-unit card", deal[got].ID, deal[got].Cost())
	}
}

func TestConservativeTakesTheCheapestCard(t *testing.T) {
	cal := GenerateSeason(3).Calendar
	deal := dealOf(richCard, engineCard, cheapCard)
	got := rivalPick(Team{ID: 2, Archetype: "conservative"}, deal, 0, cal)
	if deal[got].ID != "cheap" {
		t.Errorf("conservative took %s, want the cheapest", deal[got].ID)
	}
}

func TestAggressiveOutspendsConservative(t *testing.T) {
	cal := GenerateSeason(4).Calendar
	deals := DealsFor(4)
	agg, con := 0, 0
	for w := 0; w < WindowCount; w++ {
		agg += deals[w][rivalPick(Team{ID: 1, Archetype: "aggressive"}, deals[w], w, cal)].Cost()
		con += deals[w][rivalPick(Team{ID: 2, Archetype: "conservative"}, deals[w], w, cal)].Cost()
	}
	if agg <= con {
		t.Errorf("aggressive banked %d units, conservative %d -- aggressive must spend more", agg, con)
	}
}

func TestSpecialistFollowsItsArea(t *testing.T) {
	cal := GenerateSeason(5).Calendar
	deal := dealOf(chassCard, engineCard, aeroCard)
	// team.ID % 3 selects the area: 0 chassis, 1 engine, 2 aero.
	for id, wantID := range map[int]string{3: "cha", 7: "eng", 8: "aer"} {
		got := rivalPick(Team{ID: id, Archetype: "specialist"}, deal, 0, cal)
		if deal[got].ID != wantID {
			t.Errorf("specialist %d took %s, want %s", id, deal[got].ID, wantID)
		}
	}
}

func TestReactiveFollowsTheCalendar(t *testing.T) {
	// Windows 0 and 1 precede rounds 0-1 and 2-3. Make rounds 2-3 power
	// circuits and the window-1 reactive rival should take the engine card.
	cal := testCalendar("balanced", "balanced", "power", "power", "technical",
		"technical", "balanced", "balanced", "power", "technical")
	deal := dealOf(chassCard, engineCard, aeroCard)

	got := rivalPick(Team{ID: 4, Archetype: "reactive"}, deal, 1, cal)
	if deal[got].ID != "eng" {
		t.Errorf("reactive at power circuits took %s, want the engine card", deal[got].ID)
	}

	// Window 2 precedes rounds 4-5, both technical.
	got = rivalPick(Team{ID: 4, Archetype: "reactive"}, deal, 2, cal)
	if deal[got].ID != "cha" {
		t.Errorf("reactive at technical circuits took %s, want the chassis card", deal[got].ID)
	}
}

func TestReactiveHandlesTheFinalWindow(t *testing.T) {
	cal := GenerateSeason(6).Calendar
	deal := DealsFor(6)[WindowCount-1]
	got := rivalPick(Team{ID: 4, Archetype: "reactive"}, deal, WindowCount-1, cal)
	if got < 0 || got >= DealSize {
		t.Errorf("reactive final window picked %d", got)
	}
}

func TestTiesBreakOnLowestIndex(t *testing.T) {
	cal := GenerateSeason(7).Calendar
	same := Card{ID: "a", Name: "A", Blurb: "b", Effect: Decision{Chassis: 60, Engine: 60, Aero: 60}}
	dupe := Card{ID: "b", Name: "B", Blurb: "b", Effect: Decision{Chassis: 60, Engine: 60, Aero: 60}}
	third := Card{ID: "c", Name: "C", Blurb: "b", Effect: Decision{Chassis: 60, Engine: 60, Aero: 60}}
	deal := dealOf(same, dupe, third)
	for _, arch := range rivalArchetypes {
		if got := rivalPick(Team{ID: 1, Archetype: arch}, deal, 0, cal); got != 0 {
			t.Errorf("%s broke a three-way tie to index %d, want 0", arch, got)
		}
	}
}
