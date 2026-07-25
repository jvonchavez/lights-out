package sim

import (
	"reflect"
	"testing"

	"pgregory.net/rapid"
)

// drawPicks generates a valid pick set by construction: one in-range card
// index per development window.
func drawPicks(rt *rapid.T) []int {
	picks := make([]int, WindowCount)
	for i := range picks {
		picks[i] = rapid.IntRange(0, DealSize-1).Draw(rt, "pick")
	}
	return picks
}

func TestPropertiesOverGeneratedSeasons(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		seed := rapid.Int64().Draw(rt, "seed")
		ds := drawPicks(rt)

		res, err := RunSeason(seed, ds)
		if err != nil {
			rt.Fatalf("valid input rejected: %v", err)
		}

		// Points are never negative.
		for _, s := range res.Standings {
			if s.Points < 0 {
				rt.Errorf("team %d has negative points %d", s.TeamID, s.Points)
			}
			if s.Podiums < s.Wins {
				rt.Errorf("team %d has %d wins but only %d podiums", s.TeamID, s.Wins, s.Podiums)
			}
			if s.DNFs > RaceCount {
				rt.Errorf("team %d has %d DNFs across %d races", s.TeamID, s.DNFs, RaceCount)
			}
		}

		// Championship totals equal the sum of race points.
		raceTotal := 0
		for _, r := range res.Races {
			for _, c := range r.Cars {
				raceTotal += c.Points
			}
		}
		standingsTotal := 0
		for _, s := range res.Standings {
			standingsTotal += s.Points
		}
		if raceTotal != standingsTotal {
			rt.Errorf("race points %d != standings points %d", raceTotal, standingsTotal)
		}

		// Every race classifies every car exactly once, with no shared
		// finishing position among survivors.
		for _, r := range res.Races {
			if len(r.Cars) != TeamCount {
				rt.Errorf("round %d classified %d cars, want %d", r.Round, len(r.Cars), TeamCount)
			}
			seen := map[int]bool{}
			for _, c := range r.Cars {
				if c.DNF {
					if c.Points != 0 || c.Finish != 0 {
						rt.Errorf("round %d: DNF has finish %d points %d", r.Round, c.Finish, c.Points)
					}
					continue
				}
				if seen[c.Finish] {
					rt.Errorf("round %d has two cars in P%d", r.Round, c.Finish)
				}
				seen[c.Finish] = true
			}
		}

		// RunSeason is deterministic.
		again, err := RunSeason(seed, ds)
		if err != nil {
			rt.Fatalf("second run errored: %v", err)
		}
		if !reflect.DeepEqual(res, again) {
			rt.Error("RunSeason is not deterministic")
		}
	})
}

func TestPropertyEveryPickMustBeInRange(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Any out-of-range card index must be rejected, always.
		bad := rapid.IntRange(DealSize, DealSize*20).Draw(rt, "bad")
		window := rapid.IntRange(0, WindowCount-1).Draw(rt, "window")
		picks := evenPicks()
		picks[window] = bad
		if _, err := RunSeason(1, picks); err == nil {
			rt.Errorf("card index %d in window %d was accepted, deal size is %d", bad, window+1, DealSize)
		}
	})
}

func TestPropertyBuildAlwaysMatchesTheDeal(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		seed := rapid.Int64().Draw(rt, "seed")
		picks := drawPicks(rt)
		res, err := RunSeason(seed, picks)
		if err != nil {
			rt.Fatalf("valid picks rejected: %v", err)
		}
		deals := DealsFor(seed)
		for w, p := range picks {
			if res.Build[w].ID != deals[w][p].ID {
				rt.Errorf("window %d built %q but was dealt %q at index %d",
					w, res.Build[w].ID, deals[w][p].ID, p)
			}
		}
	})
}

func TestPropertySeasonGenerationIsStable(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		seed := rapid.Int64().Draw(rt, "seed")
		a, b := GenerateSeason(seed), GenerateSeason(seed)
		if !reflect.DeepEqual(a, b) {
			rt.Error("GenerateSeason is not deterministic")
		}
		if len(a.Calendar) != RaceCount || len(a.Teams) != TeamCount {
			rt.Errorf("malformed season for seed %d", seed)
		}
	})
}
