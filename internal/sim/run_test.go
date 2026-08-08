package sim

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

// legalPicks is a complete, valid draft: car, both drivers, engineer,
// principal, one per roll.
func legalPicks() []int {
	return []int{int(ItemCar), int(ItemDriverA), int(ItemDriverB),
		int(ItemEngineer), int(ItemPrincipal)}
}

func picksWith(roll, item int) []int {
	p := legalPicks()
	p[roll] = item
	return p
}

func TestRunSeasonIsDeterministic(t *testing.T) {
	picks := legalPicks()
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
		{"too few", make([]int, RollCount-1)},
		{"too many", make([]int, RollCount+1)},
		{"none", nil},
		{"negative index", picksWith(0, -1)},
		{"index past the last item", picksWith(3, int(itemKindCount))},
		{"wildly out of range", picksWith(2, 9999)},
		// Slot legality, which a card draft could not express: five cars is
		// five in-range indices and still not a team.
		{"five cars", []int{0, 0, 0, 0, 0}},
		{"three drivers, no car", []int{1, 2, 1, 3, 4}},
		{"no principal", []int{0, 1, 2, 3, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := RunSeason(1, tt.picks); err == nil {
				t.Error("want error, got nil")
			}
		})
	}
}

func TestLineupMatchesPicks(t *testing.T) {
	picks := []int{int(ItemDriverB), int(ItemCar), int(ItemPrincipal),
		int(ItemDriverA), int(ItemEngineer)}
	res, err := RunSeason(777, picks)
	if err != nil {
		t.Fatal(err)
	}
	rolls := RollsFor(777)
	if res.Lineup.Car.ID != rolls[1].Car.ID {
		t.Errorf("car came from %q, want the roll-2 team's", res.Lineup.Car.ID)
	}
	if res.Lineup.Principal.ID != rolls[2].Principal.ID {
		t.Errorf("principal came from %q, want the roll-3 team's", res.Lineup.Principal.ID)
	}
	if res.Lineup.Engineer.ID != rolls[4].Engineer.ID {
		t.Errorf("engineer came from %q, want the roll-5 team's", res.Lineup.Engineer.ID)
	}
	// Drivers fill in the order they were taken.
	if res.Lineup.Drivers[0].ID != rolls[0].Drivers[1].ID {
		t.Errorf("driver 1 is %q, want the roll-1 team's second driver", res.Lineup.Drivers[0].ID)
	}
	if res.Lineup.Drivers[1].ID != rolls[3].Drivers[0].ID {
		t.Errorf("driver 2 is %q, want the roll-4 team's first driver", res.Lineup.Drivers[1].ID)
	}
}

// The whole team is locked in before round one. Nothing about the lineup
// may change mid-season -- the only thing that moves is what the principal
// develops onto the car, and that is a pure function of the round.
func TestDevelopmentIsTheOnlyThingThatChangesInSeason(t *testing.T) {
	e := entriesFor(testTeam(0))[0]
	e.principal.Development = 90

	p := Profiles["balanced"]
	prev := e.carBase(0, p)
	for round := 1; round < RaceCount; round++ {
		got := e.carBase(round, p)
		if got <= prev {
			t.Fatalf("round %d: car did not improve (%d -> %d)", round, prev, got)
		}
		prev = got
	}

	// A principal exactly at the baseline stands still all year.
	flat := entriesFor(testTeam(0))[0]
	if flat.carBase(0, p) != flat.carBase(RaceCount-1, p) {
		t.Error("a principal at DevBaseline developed the car anyway")
	}
}

func TestRunSeasonShape(t *testing.T) {
	res, err := RunSeason(77, legalPicks())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Races) != RaceCount {
		t.Errorf("%d races, want %d", len(res.Races), RaceCount)
	}
	if len(res.Standings) != TeamCount {
		t.Errorf("%d standings, want %d", len(res.Standings), TeamCount)
	}
	if len(res.Drivers) != FieldSize {
		t.Errorf("%d driver standings, want %d", len(res.Drivers), FieldSize)
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
	if !reflect.DeepEqual(res.Picks, legalPicks()) {
		t.Errorf("result picks %v differ from the ones submitted", res.Picks)
	}
}

func TestStandingsSumToRacePoints(t *testing.T) {
	res, _ := RunSeason(77, legalPicks())
	total := 0
	for _, r := range res.Races {
		for _, c := range r.Entries {
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

// Both cars feed one constructors' total. This is the mechanical reason the
// second driver is a real decision rather than a passenger.
func TestBothCarsScoreForTheConstructor(t *testing.T) {
	res, err := RunSeason(2468, legalPicks())
	if err != nil {
		t.Fatal(err)
	}
	perEntry := map[int]int{}
	for _, r := range res.Races {
		for _, c := range r.Entries {
			if c.TeamID == 0 {
				perEntry[c.Entry] += c.Points
			}
		}
	}
	if perEntry[0]+perEntry[1] != res.Player.Points {
		t.Errorf("entries scored %d + %d but the constructor has %d",
			perEntry[0], perEntry[1], res.Player.Points)
	}
	if len(perEntry) != CarsPerTeam {
		t.Errorf("only %d entries scored for the player, want %d", len(perEntry), CarsPerTeam)
	}
}

func TestStandingsAreSortedByTotalOrder(t *testing.T) {
	res, _ := RunSeason(31, legalPicks())
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
	res, _ := RunSeason(142, legalPicks())
	lines := strings.Split(res.Share, "\n")
	if len(lines) != 3 {
		t.Fatalf("share has %d lines, want 3:\n%s", len(lines), res.Share)
	}
	if !strings.HasPrefix(lines[0], "Lights Out · Season ") {
		t.Errorf("bad header: %q", lines[0])
	}
	// The summary must carry its own denominator to be legible cold.
	if !strings.Contains(lines[1], " of ") || !strings.Contains(lines[1], "pts") {
		t.Errorf("summary %q should read like \"P3 of 12 · 120 pts\"", lines[1])
	}
	if n := utf8.RuneCountInString(lines[2]); n < RaceCount {
		t.Errorf("emoji row has %d runes, want at least %d: %q", n, RaceCount, lines[2])
	}
}

func TestShareStringHasNoSpoilers(t *testing.T) {
	res, _ := RunSeason(142, legalPicks())
	for _, leak := range []string{"chassis", "engine", "aero", "seed"} {
		if strings.Contains(strings.ToLower(res.Share), leak) {
			t.Errorf("share string leaks strategy detail %q:\n%s", leak, res.Share)
		}
	}
	// On a shared daily seed your LINEUP is the strategy, so the default
	// share must not name any of it. Copying with the build is opt-in.
	named := []string{res.Lineup.Car.Name, res.Lineup.Engineer.Name,
		res.Lineup.Principal.Name, res.Lineup.Drivers[0].Name, res.Lineup.Drivers[1].Name}
	for _, n := range named {
		if strings.Contains(res.Share, n) {
			t.Errorf("default share names %q", n)
		}
	}
}

func TestRollsDoNotDisturbRaceResolution(t *testing.T) {
	// The behavioural half of the stream-separation rule: adding a team-era
	// to the roster changes what is rolled, but must not move the RNG the
	// races consume. If this fails, editing the roster silently rewrites
	// every historical result.
	season := GenerateSeason(31337)
	field := make([]*entryState, 0, FieldSize)
	for _, t := range append([]Team{{ID: 0, Lineup: testTeam(0).Lineup}}, season.Rivals...) {
		for _, e := range entriesFor(t) {
			field = append(field, e)
		}
	}
	gridBefore := qualify(field, 0, season.Calendar[0].Profile, NewRNG(31337^raceSalt))

	original := Roster
	t.Cleanup(func() { Roster = original })
	extra := original[0]
	extra.ID = "test-only-1900"
	extra.Year = 1900
	Roster = append(append([]TeamEra{}, original...), extra)

	gridAfter := qualify(field, 0, season.Calendar[0].Profile, NewRNG(31337^raceSalt))
	if !reflect.DeepEqual(gridBefore, gridAfter) {
		t.Error("changing the roster moved the race RNG stream")
	}
	if RollsFor(31337) != RollsFor(31337) {
		t.Error("rolls became non-deterministic")
	}
}

func TestRollsAreDistinctAndDeterministic(t *testing.T) {
	for seed := int64(0); seed < 200; seed++ {
		rolls := RollsFor(seed)
		if rolls != RollsFor(seed) {
			t.Fatalf("seed %d: rolls are not deterministic", seed)
		}
		seen := map[string]bool{}
		for _, te := range rolls {
			if seen[te.ID] {
				t.Errorf("seed %d offers %s twice, wasting one of %d decisions", seed, te.ID, RollCount)
			}
			seen[te.ID] = true
		}
	}
}

func TestDifferentSeedsRollDifferently(t *testing.T) {
	same := 0
	for seed := int64(0); seed < 100; seed++ {
		if RollsFor(seed)[0].ID == RollsFor(seed + 1000)[0].ID {
			same++
		}
	}
	if same > 20 {
		t.Errorf("%d/100 seed pairs opened on the same team -- rolls are not varying", same)
	}
}
