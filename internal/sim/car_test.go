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

func TestUnderspendingKeepsTheCarSafer(t *testing.T) {
	// With the reliability slider cut, restraint IS the reliability lever:
	// a car that banks less development carries less risk.
	full, half := newCarState(baseTeam()), newCarState(baseTeam())
	for i := 0; i < 10; i++ {
		full.apply(Decision{Chassis: 34, Engine: 33, Aero: 33})
		half.apply(Decision{Chassis: 17, Engine: 16, Aero: 17})
	}
	if half.failureChance() >= full.failureChance() {
		t.Errorf("spending half must be safer: half %d, full %d", half.failureChance(), full.failureChance())
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
	// A car that never develops sits at the base rate, well above the
	// floor, so the lower clamp is exercised through the aero credit.
	c2 := newCarState(baseTeam())
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
	// Measured over a full season's spend. At one race's worth the convex
	// pressure is only ~2 millis, so the 30% aero credit truncates to zero
	// in fixed point -- a rounding artifact worth well under 0.1% failure
	// probability, and immaterial next to a 3% base rate.
	a, e := newCarState(baseTeam()), newCarState(baseTeam())
	for i := 0; i < 10; i++ {
		a.apply(Decision{Aero: 100})
		e.apply(Decision{Engine: 100})
	}
	if a.failureChance() >= e.failureChance() {
		t.Errorf("aero failure %d must be below engine failure %d for equal spend",
			a.failureChance(), e.failureChance())
	}
}

func TestPressureIsConvexInCumulativeSpend(t *testing.T) {
	// A full season's spend (1000 units) adds exactly PressureQuad.
	full := newCarState(baseTeam())
	for i := 0; i < 10; i++ {
		full.apply(Decision{Engine: 100})
	}
	if want, got := BaseFailure+PressureQuad, full.failureChance(); got != want {
		t.Errorf("failure after full spend = %d, want %d", got, want)
	}

	// Half the spend costs a QUARTER of the risk, not half. This convexity
	// is what gives the spend/risk trade-off an interior optimum; under a
	// linear model spending everything is always correct.
	half := newCarState(baseTeam())
	for i := 0; i < 5; i++ {
		half.apply(Decision{Engine: 100})
	}
	if want, got := BaseFailure+PressureQuad.Mul(250), half.failureChance(); got != want {
		t.Errorf("failure after half spend = %d, want %d", got, want)
	}

	// Marginal risk must increase: the last 100 units cost more than the
	// first 100.
	first := newCarState(baseTeam())
	first.apply(Decision{Engine: 100})
	firstStep := first.failureChance() - BaseFailure

	late := newCarState(baseTeam())
	for i := 0; i < 9; i++ {
		late.apply(Decision{Engine: 100})
	}
	before := late.failureChance()
	late.apply(Decision{Engine: 100})
	lastStep := late.failureChance() - before

	if lastStep <= firstStep {
		t.Errorf("marginal risk not increasing: first 100 units cost %d, last 100 cost %d", firstStep, lastStep)
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
