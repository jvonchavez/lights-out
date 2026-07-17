package sim

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func evenDecisions() []Decision {
	ds := make([]Decision, RaceCount)
	for i := range ds {
		ds[i] = Decision{34, 33, 33}
	}
	return ds
}

func decisionsWith(round int, d Decision) []Decision {
	ds := evenDecisions()
	ds[round] = d
	return ds
}

func TestRunSeasonIsDeterministic(t *testing.T) {
	ds := evenDecisions()
	first, err := RunSeason(4242, ds)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		got, err := RunSeason(4242, ds)
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
		name string
		ds   []Decision
	}{
		{"too few", make([]Decision, RaceCount-1)},
		{"too many", make([]Decision, RaceCount+1)},
		{"none", nil},
		{"negative chassis", decisionsWith(0, Decision{Chassis: -1})},
		{"negative aero", decisionsWith(3, Decision{Aero: -5})},
		{"over budget", decisionsWith(0, Decision{Chassis: RaceBudget + 1})},
		{"over budget in aggregate", decisionsWith(9, Decision{50, 50, 50})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := RunSeason(1, tt.ds); err == nil {
				t.Error("want error, got nil")
			}
		})
	}
}

func TestRunSeasonAcceptsUnderspending(t *testing.T) {
	ds := decisionsWith(0, Decision{}) // spend nothing in round 1
	if _, err := RunSeason(1, ds); err != nil {
		t.Errorf("underspending must be legal: %v", err)
	}
}

func TestRunSeasonShape(t *testing.T) {
	res, err := RunSeason(77, evenDecisions())
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
		t.Errorf("sim version %q, want %q", res.SimVersion, Version)
	}
	if res.PlayerPos < 1 || res.PlayerPos > TeamCount {
		t.Errorf("player position %d out of range", res.PlayerPos)
	}
	if res.Player.TeamID != 0 {
		t.Errorf("player standing is team %d, want 0", res.Player.TeamID)
	}
}

func TestStandingsSumToRacePoints(t *testing.T) {
	res, _ := RunSeason(77, evenDecisions())
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
	res, _ := RunSeason(31, evenDecisions())
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

func TestSpendingMoreBeatsSpendingNothing(t *testing.T) {
	// Over many seeds, developing the car must beat never developing it.
	// This is the sanity check that the whole game economy points the right
	// way; it says nothing about which strategy is best.
	nothing := make([]Decision, RaceCount)
	better := 0
	for seed := int64(0); seed < 200; seed++ {
		spent, _ := RunSeason(seed, evenDecisions())
		idle, _ := RunSeason(seed, nothing)
		if spent.Player.Points > idle.Player.Points {
			better++
		}
	}
	if better < 140 {
		t.Errorf("developing beat idling only %d/200 times", better)
	}
}

func TestShareStringShape(t *testing.T) {
	res, _ := RunSeason(142, evenDecisions())
	lines := strings.Split(res.Share, "\n")
	if len(lines) != 3 {
		t.Fatalf("share has %d lines, want 3:\n%s", len(lines), res.Share)
	}
	if !strings.HasPrefix(lines[0], "Lights Out · Season ") {
		t.Errorf("bad header: %q", lines[0])
	}
	if !strings.Contains(lines[1], "pts") {
		t.Errorf("bad summary: %q", lines[1])
	}
	if n := utf8.RuneCountInString(lines[2]); n < RaceCount {
		t.Errorf("emoji row has %d runes, want at least %d: %q", n, RaceCount, lines[2])
	}
}

func TestShareStringHasNoSpoilers(t *testing.T) {
	res, _ := RunSeason(142, evenDecisions())
	for _, leak := range []string{"chassis", "engine", "aero", "reliability", "seed"} {
		if strings.Contains(strings.ToLower(res.Share), leak) {
			t.Errorf("share string leaks strategy detail %q:\n%s", leak, res.Share)
		}
	}
}
