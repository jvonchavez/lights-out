package sim

import "testing"

func baseTeam() Team {
	return Team{ID: 0, Start: Ratings{StartRating, StartRating, StartRating}, DriverSkill: 2 * One}
}

func TestSpendingRaisesRatingsAndRisk(t *testing.T) {
	c := newCarState(baseTeam())
	before := c.failureChance()
	c.apply(Decision{Chassis: 100})
	if want := StartRating + 100*GainPerUnit; c.ratings.Chassis != want {
		t.Errorf("chassis = %d, want %d", c.ratings.Chassis, want)
	}
	if c.ratings.Engine != StartRating {
		t.Errorf("engine moved to %d without being spent on", c.ratings.Engine)
	}
	if c.failureChance() <= before {
		t.Error("performance spend must raise failure chance")
	}
}

func TestReliabilitySpendLowersRisk(t *testing.T) {
	c := newCarState(baseTeam())
	c.apply(Decision{Engine: 100})
	risky := c.failureChance()
	c.apply(Decision{Reliability: 100})
	if c.failureChance() >= risky {
		t.Error("reliability spend must lower failure chance")
	}
}

func TestReliabilitySpendBuysNoPerformance(t *testing.T) {
	c := newCarState(baseTeam())
	c.apply(Decision{Reliability: 100})
	if c.ratings != (Ratings{StartRating, StartRating, StartRating}) {
		t.Errorf("reliability spend changed ratings: %+v", c.ratings)
	}
}

func TestFailureChanceIsClamped(t *testing.T) {
	c := newCarState(baseTeam())
	for i := 0; i < 20; i++ {
		c.apply(Decision{Engine: 100})
	}
	if got := c.failureChance(); got > MaxFailure {
		t.Errorf("failure %d exceeds clamp %d", got, MaxFailure)
	}
	c2 := newCarState(baseTeam())
	for i := 0; i < 20; i++ {
		c2.apply(Decision{Reliability: 100})
	}
	if got := c2.failureChance(); got < MinFailure {
		t.Errorf("failure %d below clamp %d", got, MinFailure)
	}
}

func TestBaseFailureAppliesBeforeAnySpend(t *testing.T) {
	c := newCarState(baseTeam())
	if got := c.failureChance(); got != BaseFailure {
		t.Errorf("untouched car has failure %d, want %d", got, BaseFailure)
	}
}

func TestAeroReducesPressureRelativeToEngine(t *testing.T) {
	a, e := newCarState(baseTeam()), newCarState(baseTeam())
	a.apply(Decision{Aero: 100})
	e.apply(Decision{Engine: 100})
	if a.failureChance() >= e.failureChance() {
		t.Errorf("aero failure %d must be below engine failure %d for equal spend",
			a.failureChance(), e.failureChance())
	}
}

func TestPressureScalesWithCumulativeSpend(t *testing.T) {
	// 1000 units of perf spend should add PressurePerUnit (25%) of risk.
	c := newCarState(baseTeam())
	for i := 0; i < 10; i++ {
		c.apply(Decision{Engine: 100})
	}
	want := BaseFailure + PressurePerUnit
	if got := c.failureChance(); got != want {
		t.Errorf("failure after 1000 engine spend = %d, want %d", got, want)
	}
}

func TestPerfWeightsByCircuitProfile(t *testing.T) {
	// A chassis-heavy car scores higher at a technical circuit than a power
	// one. Compare with the same RNG stream so the noise draw is identical.
	c := newCarState(baseTeam())
	c.apply(Decision{Chassis: 100})
	tech := c.perf(Profiles["technical"], QualiSigma, NewRNG(1))
	power := c.perf(Profiles["power"], QualiSigma, NewRNG(1))
	if tech <= power {
		t.Errorf("chassis car: technical %d, power %d -- want technical higher", tech, power)
	}
}

func TestPerfDrawsExactlyOnce(t *testing.T) {
	c := newCarState(baseTeam())
	r1, r2 := NewRNG(5), NewRNG(5)
	c.perf(Profiles["balanced"], QualiSigma, r1)
	r2.Normal(QualiSigma)
	if r1.Uint64() != r2.Uint64() {
		t.Fatal("perf consumed a different number of RNG draws than one Normal")
	}
}
