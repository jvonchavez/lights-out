package sim

import "testing"

func TestStrategiesAreValid(t *testing.T) {
	season := GenerateSeason(1)
	for _, name := range []string{"even", "aggressive", "specialist", "adaptive", "aerofirst", "idle"} {
		t.Run(name, func(t *testing.T) {
			ds := Strategy(name, season)
			if len(ds) != RaceCount {
				t.Fatalf("got %d decisions, want %d", len(ds), RaceCount)
			}
			for i, d := range ds {
				if d.Total() > season.Budgets[i] {
					t.Errorf("round %d spends %d, budget is %d", i+1, d.Total(), season.Budgets[i])
				}
				if d.Chassis < 0 || d.Engine < 0 || d.Aero < 0 {
					t.Errorf("round %d has a negative allocation: %+v", i+1, d)
				}
			}
			if _, err := RunSeason(1, ds); err != nil {
				t.Errorf("strategy %q produced decisions RunSeason rejects: %v", name, err)
			}
		})
	}
}

func TestUnknownStrategyReturnsNil(t *testing.T) {
	if Strategy("nonsense", GenerateSeason(1)) != nil {
		t.Error("unknown strategy must return nil")
	}
}

func TestIdleStrategySpendsNothing(t *testing.T) {
	for _, d := range Strategy("idle", GenerateSeason(1)) {
		if d.Total() != 0 {
			t.Errorf("idle spent %d", d.Total())
		}
	}
}
