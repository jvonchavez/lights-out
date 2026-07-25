package sim

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

// evenPicks takes the middle card every window.
func evenPicks() []int {
	p := make([]int, WindowCount)
	for i := range p {
		p[i] = 1
	}
	return p
}

func picksWith(window, card int) []int {
	p := evenPicks()
	p[window] = card
	return p
}

func TestRunSeasonIsDeterministic(t *testing.T) {
	picks := evenPicks()
	first, err := RunSeason(4242, picks)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		got, err := RunSeason(4242, picks)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d diverged from the first", i)
		}
	}
}

func TestRunSeasonRejectsBadInput(t *testing.T) {
	tests := []struct {
		name  string
		picks []int
	}{
		{"too few", make([]int, WindowCount-1)},
		{"too many", make([]int, WindowCount+1)},
		{"none", nil},
		{"negative index", picksWith(0, -1)},
		{"index past the deal", picksWith(3, DealSize)},
		{"wildly out of range", picksWith(2, 9999)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := RunSeason(1, tt.picks); err == nil {
				t.Error("want error, got nil")
			}
		})
	}
}

func TestBuildMatchesPicks(t *testing.T) {
	picks := []int{0, 1, 2, 1, 0}
	res, err := RunSeason(777, picks)
	if err != nil {
		t.Fatal(err)
	}
	deals := DealsFor(777)
	if len(res.Build) != WindowCount {
		t.Fatalf("build has %d cards, want %d", len(res.Build), WindowCount)
	}
	for w, want := range picks {
		if res.Build[w].ID != deals[w][want].ID {
			t.Errorf("window %d built %q, want %q", w, res.Build[w].ID, deals[w][want].ID)
		}
	}
}

func TestDevelopmentOnlyHappensAtWindows(t *testing.T) {
	// The player's car must change across exactly WindowCount rounds.
	season := GenerateSeason(11)
	deals := DealsFor(11)
	car := newCarState(season.Teams[0])

	changed := 0
	for round := 0; round < RaceCount; round++ {
		before := car.ratings
		if w, ok := windowAt(round); ok {
			car.apply(deals[w][0].Effect)
		}
		if car.ratings != before {
			changed++
		}
	}
	if changed != WindowCount {
		t.Errorf("ratings changed in %d rounds, want %d", changed, WindowCount)
	}
}

func TestSeasonBanksRoughlyAFullBudget(t *testing.T) {
	// PressureQuad is calibrated at 1000 units. If typical builds drift far
	// from that, the risk curve is no longer measuring what it was tuned to.
	total, n := 0, 0
	for seed := int64(0); seed < 300; seed++ {
		res, err := RunSeason(seed, evenPicks())
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range res.Build {
			total += c.Cost()
		}
		n++
	}
	avg := total / n
	if avg < 850 || avg > 1150 {
		t.Errorf("mean season spend %d units, want near 1000", avg)
	}
}

func TestRunSeasonShape(t *testing.T) {
	res, err := RunSeason(77, evenPicks())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Races) != RaceCount {
		t.Errorf("%d races, want %d", len(res.Races), RaceCount)
	}
	if len(res.Standings) != TeamCount {
		t.Errorf("%d standings, want %d", len(res.Standings), TeamCount)
	}
	for i, r := range res.Races {
		if r.Round != i+1 {
			t.Errorf("race %d has round %d", i, r.Round)
		}
	}
	if res.SimVersion != Version {
		t.Errorf("sim version = %q, want %q", res.SimVersion, Version)
	}
	if res.PlayerPos < 1 || res.PlayerPos > TeamCount {
		t.Errorf("player position %d out of range", res.PlayerPos)
	}
}

func TestStandingsSumToRacePoints(t *testing.T) {
	res, _ := RunSeason(77, evenPicks())
	total := 0
	for _, r := range res.Races {
		for _, c := range r.Cars {
			total += c.Points
		}
	}
	standingsTotal := 0
	for _, s := range res.Standings {
		standingsTotal += s.Points
	}
	if total != standingsTotal {
		t.Errorf("races total %d, standings total %d", total, standingsTotal)
	}
}

func TestStandingsAreSortedByTotalOrder(t *testing.T) {
	res, _ := RunSeason(31, evenPicks())
	for i := 0; i+1 < len(res.Standings); i++ {
		a, b := res.Standings[i], res.Standings[i+1]
		switch {
		case a.Points != b.Points:
			if a.Points < b.Points {
				t.Errorf("standings %d/%d out of order on points", i, i+1)
			}
		case a.Wins != b.Wins:
			if a.Wins < b.Wins {
				t.Errorf("standings %d/%d out of order on wins", i, i+1)
			}
		case a.Podiums != b.Podiums:
			if a.Podiums < b.Podiums {
				t.Errorf("standings %d/%d out of order on podiums", i, i+1)
			}
		default:
			if a.TeamID > b.TeamID {
				t.Errorf("standings %d/%d out of order on team ID", i, i+1)
			}
		}
	}
}

func TestShareStringShape(t *testing.T) {
	res, _ := RunSeason(142, evenPicks())
	lines := strings.Split(res.Share, "\n")
	if len(lines) != 3 {
		t.Fatalf("share has %d lines, want 3:\n%s", len(lines), res.Share)
	}
	if !strings.HasPrefix(lines[0], "Lights Out · Season ") {
		t.Errorf("bad header: %q", lines[0])
	}
	// The summary must carry its own denominator to be legible cold.
	if !strings.Contains(lines[1], " of ") || !strings.Contains(lines[1], "pts") {
		t.Errorf("summary %q should read like \"P3 of 11 · 120 pts\"", lines[1])
	}
	if n := utf8.RuneCountInString(lines[2]); n < RaceCount {
		t.Errorf("emoji row has %d runes, want at least %d: %q", n, RaceCount, lines[2])
	}
}

func TestShareStringHasNoSpoilers(t *testing.T) {
	res, _ := RunSeason(142, evenPicks())
	for _, leak := range []string{"chassis", "engine", "aero", "seed"} {
		if strings.Contains(strings.ToLower(res.Share), leak) {
			t.Errorf("share string leaks strategy detail %q:\n%s", leak, res.Share)
		}
	}
	// The default share must not name the build either: on a shared daily
	// seed your parts ARE the strategy. Copying with the build is opt-in.
	for _, c := range res.Build {
		if strings.Contains(res.Share, c.Name) {
			t.Errorf("default share names the part %q", c.Name)
		}
	}
}

func TestDealsDoNotDisturbRaceResolution(t *testing.T) {
	// The behavioural half of the three-stream rule: adding a card to the
	// pool changes what is dealt, but must not move the RNG the races
	// consume. If this fails, editing the pool silently rewrites history.
	season := GenerateSeason(31337)
	cars := make([]*carState, len(season.Teams))
	for i, tm := range season.Teams {
		cars[i] = newCarState(tm)
	}
	gridBefore := qualify(cars, season.Calendar[0].Profile, NewRNG(31337^raceSalt))

	original := CardPool
	t.Cleanup(func() { CardPool = original })
	CardPool = append(append([]Card{}, original...), Card{
		ID: "test-only", Name: "Test Only", Blurb: "not a real part",
		Effect: Decision{Chassis: 50, Engine: 50, Aero: 50},
	})

	gridAfter := qualify(cars, season.Calendar[0].Profile, NewRNG(31337^raceSalt))
	if !reflect.DeepEqual(gridBefore, gridAfter) {
		t.Error("changing the card pool moved the race RNG stream")
	}
	if reflect.DeepEqual(DealsFor(31337), DealsFor(31337)) == false {
		t.Error("deals became non-deterministic")
	}
}
