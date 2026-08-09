package sim

// entryState is one CAR on the grid: a chassis, the driver in it, and the
// engineer and principal it shares with its team-mate. There are two per
// team and FieldSize in total.
//
// This replaces the old carState, which tracked cumulative development
// SPEND. Nothing is bought during a season any more -- the team is locked
// in before round one -- so the only thing that changes across a season is
// what the principal develops.
//
// RESCALING RULE. Roster ratings are plain integer counts, so multiplying a
// rating by a per-point rate is plain multiplication with NO /One, and the
// product is cast to Milli at the end. Only Milli x Milli uses .Mul.
type entryState struct {
	teamID int
	entry  int

	car       CarSpec
	driver    DriverSpec
	engineer  EngineerSpec
	principal PrincipalSpec
}

func entriesFor(t Team) [CarsPerTeam]*entryState {
	var out [CarsPerTeam]*entryState
	for i := 0; i < CarsPerTeam; i++ {
		out[i] = &entryState{
			teamID:    t.ID,
			entry:     i,
			car:       t.Lineup.Car,
			driver:    t.Lineup.Drivers[i],
			engineer:  t.Lineup.Engineer,
			principal: t.Lineup.Principal,
		}
	}
	return out
}

// development is the rating the principal has added to the car by this
// round. It is a pure function of the round with no RNG draw, so the car
// visibly climbs across the season.
//
// Worth what, exactly: the roster's best principal (95) beats its worst
// (74) by 1.70 rating points on season average and 3.40 by the final round,
// which is 2.3 sigma of race noise. Enough to overturn a small car deficit
// and to decide a close title; NOT enough to make a mid car beat a great
// one, because car Overall spans 61 to 97. TestDevelopmentIsWorthWhatTheDocsSay
// pins those two numbers so the claim cannot drift.
func (e *entryState) development(round int) Milli {
	return Milli((e.principal.Development - DevBaseline) * round * DevRate)
}

// carBase is the chassis's contribution at a circuit: the weighted sum of
// its three performance ratings, plus whatever has been developed onto it.
//
// Aero is a plain third weight here. Under the old model it MULTIPLIED the
// whole sum, because performance was linear in player-allocated ratings and
// concentration therefore always beat spreading. Nothing is allocated any
// more -- a car is taken whole -- so that degenerate case cannot arise and
// the multiplier is gone with it.
func (e *entryState) carBase(round int, p CircuitProfile) Milli {
	return FromInt(e.car.Cornering).Mul(p.Chassis) +
		FromInt(e.car.Power).Mul(p.Engine) +
		FromInt(e.car.Aero).Mul(p.Aero) +
		e.development(round)
}

// perf is a performance roll at a circuit.
//
// The RNG draw is LAST and happens EXACTLY ONCE per call -- draw order is
// the determinism contract, and adding or moving a draw changes every
// downstream result. RNG.Normal consumes a fixed twelve uniforms whatever
// sigma it is given, so a per-driver sigma changes no draw counts.
func (e *entryState) perf(round int, p CircuitProfile, quali bool, r *RNG) Milli {
	v := e.carBase(round, p)
	v += Milli((e.driver.Pace - PaceBaseline) * PaceRate)
	v += Milli((e.principal.Leadership - StaffBaseline) * LeadershipRate)

	sigma := RaceSigma
	if quali {
		sigma = QualiSigma
		v += Milli((e.engineer.Setup - StaffBaseline) * SetupRate)
	} else {
		v += Milli((e.engineer.Strategy - StaffBaseline) * StrategyRate)
		// Nerve is the closing-rounds rating: a principal who holds it
		// together when the championship is actually being decided.
		if round >= RaceCount-NerveRounds {
			v += Milli((e.principal.Nerve - StaffBaseline) * NerveRate)
		}
	}

	return v + r.Normal(e.sigma(sigma))
}

// sigma widens the performance distribution for erratic drivers and narrows
// it for metronomes. This is what stops Overall from being the whole story:
// Gilles Villeneuve and Alain Prost are not the same bet.
func (e *entryState) sigma(base Milli) Milli {
	f := One + Milli((ConsistencyBaseline-e.driver.Consistency)*ConsistencyRate)
	if f < MinSigmaFactor {
		f = MinSigmaFactor
	}
	if f > MaxSigmaFactor {
		f = MaxSigmaFactor
	}
	return base.Mul(f)
}

// gridPenalty is what starting out of position costs at this circuit, after
// the driver's racecraft buys some of it back. At a power circuit a fast
// car recovers from a bad slot; at a technical one it stays stuck, and a
// driver who can actually overtake is worth most exactly there.
func (e *entryState) gridPenalty(gridPos int, p CircuitProfile) Milli {
	// FromInt, not Milli: grid position is a plain count and must be
	// SCALED to fixed-point before multiplying. Milli() would treat 10 as
	// 0.01 and make the penalty vanish against the noise.
	raw := FromInt(gridPos - 1).Mul(p.OvertakeDifficulty)

	relief := Milli((e.driver.Racecraft - StaffBaseline) * RacecraftRate)
	if relief < 0 {
		relief = 0
	}
	if relief > MaxGridRelief {
		relief = MaxGridRelief
	}
	return raw - raw.Mul(relief)
}

// failure resolves this car's retirement with EXACTLY ONE rng draw.
//
// Mechanical failure and driver error are separate probabilities but share
// the draw: the car retires if the uniform falls below their sum, and the
// cause is whichever band it landed in. One draw keeps the determinism
// contract simple and still gives the race reel a reason to name, which is
// the difference between "DNF" and "engine, on lap 41".
func (e *entryState) failure(r *RNG) (bool, string) {
	short := RelCeiling - e.car.Reliability
	if short < 0 {
		short = 0
	}
	mech := MechFloor + Milli(short*short*MechQuad/100)
	// A good race operation removes a share of whatever risk is there.
	if credit := mech * Milli((e.engineer.Ops-StaffBaseline)*MechOpsRate) / One; credit > 0 {
		mech -= credit
	}
	if mech < 0 {
		mech = 0
	}

	derr := (DriverErrBase -
		Milli((e.driver.Composure-StaffBaseline)*ComposureRate)) / FailureDivisor
	if derr < 0 {
		derr = 0
	}

	// Rails, not active constraints: real roster values land between about
	// 2% and 26%. Clamping the total keeps the mechanical share
	// proportional so the reason stays honest.
	total := mech + derr
	if total > MaxFailure {
		mech = mech * MaxFailure / total
		total = MaxFailure
	}
	if total < MinFailure {
		total = MinFailure
		if mech > total {
			mech = total
		}
	}

	u := r.Milli()
	if u >= total {
		return false, ""
	}
	if u < mech {
		return true, DNFMechanical
	}
	return true, DNFDriver
}
