package sim

import "testing"

func TestStrategiesProduceLegalDrafts(t *testing.T) {
	season := GenerateSeason(1)
	for _, name := range StrategyNames {
		t.Run(name, func(t *testing.T) {
			picks := Strategy(name, season)
			if len(picks) != RollCount {
				t.Fatalf("got %d picks, want %d", len(picks), RollCount)
			}
			for i, p := range picks {
				if p < 0 || p >= int(itemKindCount) {
					t.Errorf("roll %d picked %d, outside [0,%d)", i, p, itemKindCount)
				}
			}
			// Legal by construction, not by luck: every scripted line must
			// fill exactly one car, two drivers, an engineer and a
			// principal.
			if _, err := RunSeason(season.Seed, picks); err != nil {
				t.Errorf("strategy %q produced a draft RunSeason rejects: %v", name, err)
			}
		})
	}
}

// Legality must hold on every seed, not just a convenient one: it is the
// last roll, where only one slot is open, that a careless scorer breaks on.
func TestStrategiesAreLegalOnEverySeed(t *testing.T) {
	for _, name := range StrategyNames {
		for seed := int64(0); seed < 300; seed++ {
			season := GenerateSeason(seed)
			if _, err := RunSeason(seed, Strategy(name, season)); err != nil {
				t.Fatalf("%s on seed %d: %v", name, seed, err)
			}
		}
	}
}

func TestUnknownStrategyReturnsNil(t *testing.T) {
	if Strategy("nonsense", GenerateSeason(1)) != nil {
		t.Error("unknown strategy must return nil")
	}
}

func TestStrategiesAreDeterministic(t *testing.T) {
	season := GenerateSeason(99)
	for _, name := range StrategyNames {
		a := Strategy(name, season)
		b := Strategy(name, season)
		for i := range a {
			if a[i] != b[i] {
				t.Errorf("%s is not deterministic at roll %d", name, i)
			}
		}
	}
}

func TestCarfirstTakesTheCarFirst(t *testing.T) {
	for seed := int64(0); seed < 100; seed++ {
		picks := Strategy("carfirst", GenerateSeason(seed))
		if picks[0] != int(ItemCar) {
			t.Fatalf("seed %d: carfirst opened with item %d, want the car", seed, picks[0])
		}
	}
}

// Starpower fills both driver slots before it takes anything else, and
// takes the better of the two drivers on offer each time.
//
// Note what this does NOT assert. Measured across 200 seeds, starpower ends
// up with WORSE drivers than carfirst -- 35288 rating against 35895 --
// because filling both driver slots from the first two rolls forfeits the
// choice that later rolls would have offered. Grabbing the thing you want
// first is not the same as ending up with the best of it, and that is a
// real property of the draft rather than a bug in the strategy.
func TestStarpowerTakesDriversFirst(t *testing.T) {
	for seed := int64(0); seed < 200; seed++ {
		season := GenerateSeason(seed)
		picks := Strategy("starpower", season)
		for i := 0; i < CarsPerTeam; i++ {
			if slotFor(ItemKind(picks[i])) != SlotDriver {
				t.Fatalf("seed %d: roll %d took a %v, want a driver", seed, i+1, slotFor(ItemKind(picks[i])))
			}
		}
		// And it takes the better one of the pair on offer.
		for i := 0; i < CarsPerTeam; i++ {
			te := season.Rolls[i]
			want := ItemDriverA
			if te.Drivers[1].Overall() > te.Drivers[0].Overall() {
				want = ItemDriverB
			}
			if ItemKind(picks[i]) != want {
				t.Errorf("seed %d roll %d: took %v from %s, want the higher-rated driver",
					seed, i+1, ItemKind(picks[i]), te.Label())
			}
		}
	}
}

func TestCautiousBuysReliability(t *testing.T) {
	cautious, best := 0, 0
	for seed := int64(0); seed < 200; seed++ {
		season := GenerateSeason(seed)
		c, err := BuildLineup(season.Rolls, Strategy("cautious", season))
		if err != nil {
			t.Fatal(err)
		}
		b, err := BuildLineup(season.Rolls, Strategy("bestavailable", season))
		if err != nil {
			t.Fatal(err)
		}
		cautious += c.Car.Reliability
		best += b.Car.Reliability
	}
	if cautious <= best {
		t.Errorf("cautious drafted %d reliability, bestavailable %d", cautious, best)
	}
}
