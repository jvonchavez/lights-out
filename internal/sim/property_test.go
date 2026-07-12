package sim

import (
	"reflect"
	"testing"

	"pgregory.net/rapid"
)

// drawDecisions generates a valid allocation set by construction: each area
// draws from whatever budget remains, so the total can never exceed it.
func drawDecisions(rt *rapid.T) []Decision {
	ds := make([]Decision, RaceCount)
	for i := range ds {
		rem := RaceBudget
		c := rapid.IntRange(0, rem).Draw(rt, "chassis")
		rem -= c
		e := rapid.IntRange(0, rem).Draw(rt, "engine")
		rem -= e
		a := rapid.IntRange(0, rem).Draw(rt, "aero")
		rem -= a
		r := rapid.IntRange(0, rem).Draw(rt, "reliability")
		ds[i] = Decision{Chassis: c, Engine: e, Aero: a, Reliability: r}
	}
	return ds
}

func TestPropertiesOverGeneratedSeasons(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		seed := rapid.Int64().Draw(rt, "seed")
		ds := drawDecisions(rt)

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

func TestPropertyEveryAllocationRespectsItsBudget(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Any allocation exceeding the budget must be rejected, always.
		over := rapid.IntRange(RaceBudget+1, RaceBudget*4).Draw(rt, "over")
		round := rapid.IntRange(0, RaceCount-1).Draw(rt, "round")
		ds := evenDecisions()
		ds[round] = Decision{Chassis: over}
		if _, err := RunSeason(1, ds); err == nil {
			rt.Errorf("allocation of %d in round %d was accepted, budget is %d", over, round+1, RaceBudget)
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
