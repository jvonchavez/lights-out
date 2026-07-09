package sim

// carState is a car's evolving condition across a season: its ratings, and
// the cumulative spending that drives its failure probability.
//
// RESCALING RULE. Budget spends are plain unit counts, not Milli, so
// multiplying a spend by a per-unit Milli rate needs NO /One -- use plain
// multiplication. Only multiplying two Milli values needs .Mul. The two
// cases look alike and confusing them silently scales the whole game wrong.
type carState struct {
	team      Team
	ratings   Ratings
	perfSpend int // cumulative chassis+engine+aero
	aeroSpend int // cumulative aero alone, for the efficiency credit
	relSpend  int // cumulative reliability
}

func newCarState(t Team) *carState {
	return &carState{team: t, ratings: t.Start}
}

// apply banks one race's development. Spend is a plain count, so the gain
// is a plain multiplication: at GainPerUnit 100, spending 100 units buys
// 10000 millis, a full +10.0 rating.
func (c *carState) apply(d Decision) {
	c.ratings.Chassis += Milli(d.Chassis) * GainPerUnit
	c.ratings.Engine += Milli(d.Engine) * GainPerUnit
	c.ratings.Aero += Milli(d.Aero) * GainPerUnit

	c.perfSpend += d.Chassis + d.Engine + d.Aero
	c.aeroSpend += d.Aero
	c.relSpend += d.Reliability
}

// perf is a performance roll at a circuit: the weighted sum of ratings by
// circuit profile, plus driver skill, plus noise.
//
// Both operands are Milli here, so .Mul is correct. The RNG draw is LAST
// and happens EXACTLY ONCE per call -- draw order is the determinism
// contract, and adding or moving a draw changes every downstream result.
func (c *carState) perf(p CircuitProfile, sigma Milli, r *RNG) Milli {
	return c.ratings.Chassis.Mul(p.Chassis) +
		c.ratings.Engine.Mul(p.Engine) +
		c.ratings.Aero.Mul(p.Aero) +
		c.team.DriverSkill +
		r.Normal(sigma)
}

// failureChance is the probability this car suffers a DNF, in Milli.
//
// Development spend raises risk permanently, not just for the next race:
// you are running a highly strung car for the rest of the season. That is
// what stops "spend everything in round 1" from being a free win.
//
// .Mul is right for pressure and relief -- it is the /1000 inside Mul that
// turns "250 per 1000 units spent" into a probability, so 1000 cumulative
// perf spend yields 1000 * 250 / 1000 = 250 millis, i.e. 25%.
func (c *carState) failureChance() Milli {
	pressure := Milli(c.perfSpend).Mul(PressurePerUnit)

	// Aero is the only area that improves efficiency: it is credited back a
	// share of the pressure its own spend caused.
	credit := Milli(c.aeroSpend).Mul(PressurePerUnit).Mul(AeroEfficiency)

	relief := Milli(c.relSpend).Mul(ReliabilityGain)

	p := BaseFailure + pressure - credit - relief
	if p > MaxFailure {
		return MaxFailure
	}
	if p < MinFailure {
		return MinFailure
	}
	return p
}
