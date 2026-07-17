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
}

// perf is a performance roll at a circuit: the weighted sum of ratings by
// circuit profile, plus driver skill, plus noise.
//
// Both operands are Milli here, so .Mul is correct. The RNG draw is LAST
// and happens EXACTLY ONCE per call -- draw order is the determinism
// contract, and adding or moving a draw changes every downstream result.
func (c *carState) perf(p CircuitProfile, sigma Milli, r *RNG) Milli {
	base := c.ratings.Chassis.Mul(p.Chassis) +
		c.ratings.Engine.Mul(p.Engine) +
		c.ratings.Aero.Mul(p.Aero)

	// Aero scales with both: it multiplies the whole weighted sum, not just
	// its own term. Without this the performance model is linear in the
	// ratings, concentration always beats spreading, and the optimal play
	// collapses to "put everything in whichever area the calendar weights
	// highest" -- which makes the three sliders a non-decision.
	base += base.Mul(c.aeroBonus())

	return base + c.team.DriverSkill + r.Normal(sigma)
}

// aeroBonus is the multiplier aero rating above the baseline confers,
// clamped at MaxAeroBonus. Negative aero (a car below the baseline) confers
// nothing rather than a penalty.
func (c *carState) aeroBonus() Milli {
	above := c.ratings.Aero - StartRating
	if above <= 0 {
		return 0
	}
	if bonus := above.Mul(AeroScale); bonus < MaxAeroBonus {
		return bonus
	}
	return MaxAeroBonus
}

// failureChance is the probability this car suffers a DNF, in Milli.
//
// Development spend raises risk permanently, not just for the next race:
// you are running a highly strung car for the rest of the season. That is
// what stops "spend everything in round 1" from being a free win.
//
// Risk is CONVEX in cumulative spend: the first 100 units cost far less
// risk than the last 100. Under a linear cost the trade-off has no interior
// optimum and spending the whole budget is always correct, which makes the
// central decision a non-decision. See PressureQuad in params.go.
func (c *carState) failureChance() Milli {
	// A season's full budget is 1000 units, so Milli(perfSpend) already
	// reads as the fraction of a full season's development spent, with
	// 1000 units == One. Squaring it gives the convex risk curve.
	frac := Milli(c.perfSpend)
	pressure := frac.Mul(frac).Mul(PressureQuad)

	// Aero is the only area that improves efficiency: it is credited back a
	// share of the pressure its own spend caused, in proportion to how much
	// of the development it represents.
	var credit Milli
	if c.perfSpend > 0 {
		share := Milli(c.aeroSpend).Div(Milli(c.perfSpend))
		credit = pressure.Mul(share).Mul(AeroEfficiency)
	}

	p := BaseFailure + pressure - credit
	if p > MaxFailure {
		return MaxFailure
	}
	if p < MinFailure {
		return MinFailure
	}
	return p
}
