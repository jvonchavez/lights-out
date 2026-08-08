package sim

import (
	"reflect"
	"testing"

	"pgregory.net/rapid"
)

// drawPicks generates a legal draft by construction: at each roll it takes
// only items whose slot still has room and whose removal still leaves
// enough rolls to fill everything else.
func drawPicks(rt *rapid.T) []int {
	picks := make([]int, RollCount)
	filled := map[Slot]int{}
	for i := 0; i < RollCount; i++ {
		var legal []int
		for k := ItemCar; k < itemKindCount; k++ {
			s := slotFor(k)
			if filled[s] >= slotCapacity[s] || !fillable(filled, s, RollCount-i) {
				continue
			}
			legal = append(legal, int(k))
		}
		p := rapid.SampledFrom(legal).Draw(rt, "pick")
		picks[i] = p
		filled[slotFor(ItemKind(p))]++
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

		for _, s := range res.Standings {
			if s.Points < 0 {
				rt.Errorf("team %d has negative points %d", s.TeamID, s.Points)
			}
			if s.Podiums < s.Wins {
				rt.Errorf("team %d has %d wins but only %d podiums", s.TeamID, s.Wins, s.Podiums)
			}
			// Two cars per team, so a team can retire twice per round.
			if s.DNFs > RaceCount*CarsPerTeam {
				rt.Errorf("team %d has %d DNFs across %d races", s.TeamID, s.DNFs, RaceCount)
			}
		}

		// Championship totals equal the sum of race points, and the driver
		// table accounts for exactly the same points as the constructors'.
		raceTotal := 0
		for _, r := range res.Races {
			for _, c := range r.Entries {
				raceTotal += c.Points
			}
		}
		standingsTotal, driversTotal := 0, 0
		for _, s := range res.Standings {
			standingsTotal += s.Points
		}
		for _, d := range res.Drivers {
			driversTotal += d.Points
		}
		if raceTotal != standingsTotal {
			rt.Errorf("race points %d != standings points %d", raceTotal, standingsTotal)
		}
		if driversTotal != standingsTotal {
			rt.Errorf("driver points %d != constructor points %d", driversTotal, standingsTotal)
		}

		for _, r := range res.Races {
			if len(r.Entries) != FieldSize {
				rt.Errorf("round %d classified %d cars, want %d", r.Round, len(r.Entries), FieldSize)
			}
			seen := map[int]bool{}
			for _, c := range r.Entries {
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
		bad := rapid.IntRange(int(itemKindCount), int(itemKindCount)*20).Draw(rt, "bad")
		roll := rapid.IntRange(0, RollCount-1).Draw(rt, "roll")
		picks := legalPicks()
		picks[roll] = bad
		if _, err := RunSeason(1, picks); err == nil {
			rt.Errorf("item index %d at roll %d was accepted, there are only %d items",
				bad, roll+1, itemKindCount)
		}
	})
}

// A draft that overfills a slot -- three drivers and no car, say -- must be
// rejected however plausible each individual index looks. This is the check
// the old range test could not express, because a card draft had only one
// kind of slot.
func TestPropertyIllegalSlotShapesAreRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		picks := make([]int, RollCount)
		for i := range picks {
			picks[i] = rapid.IntRange(0, int(itemKindCount)-1).Draw(rt, "pick")
		}
		filled := map[Slot]int{}
		for _, p := range picks {
			filled[slotFor(ItemKind(p))]++
		}
		legal := true
		for s, cap := range slotCapacity {
			if filled[s] != cap {
				legal = false
			}
		}
		_, err := RunSeason(7, picks)
		if legal && err != nil {
			rt.Errorf("legal draft %v rejected: %v", picks, err)
		}
		if !legal && err == nil {
			rt.Errorf("illegal draft %v accepted", picks)
		}
	})
}

// What the season simulates must be exactly what the rolls offered and the
// picks selected. This is the trust boundary stated as a property.
func TestPropertyLineupAlwaysMatchesTheRolls(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		seed := rapid.Int64().Draw(rt, "seed")
		picks := drawPicks(rt)
		res, err := RunSeason(seed, picks)
		if err != nil {
			rt.Fatalf("valid picks rejected: %v", err)
		}
		rolls := RollsFor(seed)
		if rolls != res.Rolls {
			rt.Error("result rolls differ from RollsFor(seed)")
		}
		want, err := BuildLineup(rolls, picks)
		if err != nil {
			rt.Fatalf("BuildLineup rejected picks RunSeason accepted: %v", err)
		}
		if !reflect.DeepEqual(want, res.Lineup) {
			rt.Error("result lineup differs from replaying the picks against the rolls")
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
		if len(a.Calendar) != RaceCount || len(a.Rivals) != TeamCount-1 {
			rt.Errorf("malformed season for seed %d", seed)
		}
	})
}
