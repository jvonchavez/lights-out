package sim

import "testing"

var strategyNames = []string{"greedy", "cautious", "aerofirst", "adaptive", "first"}

func TestStrategiesProduceValidPicks(t *testing.T) {
	season := GenerateSeason(1)
	for _, name := range strategyNames {
		t.Run(name, func(t *testing.T) {
			picks := Strategy(name, season)
			if len(picks) != WindowCount {
				t.Fatalf("got %d picks, want %d", len(picks), WindowCount)
			}
			for w, p := range picks {
				if p < 0 || p >= DealSize {
					t.Errorf("window %d picked %d, outside [0,%d)", w, p, DealSize)
				}
			}
			if _, err := RunSeason(season.Seed, picks); err != nil {
				t.Errorf("strategy %q produced picks RunSeason rejects: %v", name, err)
			}
		})
	}
}

func TestUnknownStrategyReturnsNil(t *testing.T) {
	if Strategy("nonsense", GenerateSeason(1)) != nil {
		t.Error("unknown strategy must return nil")
	}
}

func TestStrategiesAreDeterministic(t *testing.T) {
	season := GenerateSeason(99)
	for _, name := range strategyNames {
		a := Strategy(name, season)
		b := Strategy(name, season)
		for i := range a {
			if a[i] != b[i] {
				t.Errorf("%s is not deterministic at window %d", name, i)
			}
		}
	}
}

func TestGreedyOutspendsCautious(t *testing.T) {
	// Across many seeds, greedy must bank strictly more development.
	greedier := 0
	for seed := int64(0); seed < 100; seed++ {
		season := GenerateSeason(seed)
		deals := DealsFor(seed)
		g, c := 0, 0
		for w, p := range Strategy("greedy", season) {
			g += deals[w][p].Cost()
		}
		for w, p := range Strategy("cautious", season) {
			c += deals[w][p].Cost()
		}
		if g > c {
			greedier++
		}
	}
	if greedier < 95 {
		t.Errorf("greedy outspent cautious in only %d/100 seasons", greedier)
	}
}

func TestAerofirstBuysAero(t *testing.T) {
	// It must take meaningfully more aero than the no-thought baseline.
	aero, base := 0, 0
	for seed := int64(0); seed < 100; seed++ {
		season := GenerateSeason(seed)
		deals := DealsFor(seed)
		for w, p := range Strategy("aerofirst", season) {
			aero += deals[w][p].Effect.Aero
		}
		for w, p := range Strategy("first", season) {
			base += deals[w][p].Effect.Aero
		}
	}
	if aero <= base {
		t.Errorf("aerofirst banked %d aero units, baseline %d", aero, base)
	}
}
