package sim

// rivalPick chooses one card from a window's deal for a rival team.
//
// Rivals are dealt the SAME three cards as the player and choose by
// archetype, which makes their behaviour legible -- you can see that
// Bellweather took the gearbox you passed on. They never see the player's
// picks: docs/Game Design.md is explicit that an adaptive AI would make the
// daily seed unfair, because two players making different early choices
// would face different fields.
//
// Every rule ends by breaking ties on the LOWEST index, so the choice is a
// total order and cannot depend on how the deal happened to be assembled.
func rivalPick(t Team, deal [DealSize]Card, window int, cal []Circuit) int {
	best := 0
	switch t.Archetype {
	case "aggressive":
		// Spends heavily and early: high DNF rate, high ceiling.
		for i := 1; i < DealSize; i++ {
			if deal[i].Cost() > deal[best].Cost() {
				best = i
			}
		}
	case "conservative":
		// Rarely fails spectacularly, rarely dominates.
		for i := 1; i < DealSize; i++ {
			if deal[i].Cost() < deal[best].Cost() {
				best = i
			}
		}
	case "specialist":
		// Pours everything into one area, dominant where it suits and
		// badly exposed everywhere else.
		area := t.ID % 3
		for i := 1; i < DealSize; i++ {
			if areaUnits(deal[i], area) > areaUnits(deal[best], area) {
				best = i
			}
		}
	case "reactive":
		// Invests toward whichever archetype dominates the races this
		// window precedes, clamped at the end of the calendar.
		wc, we, wa := weightsForWindow(window, cal)
		bestScore := cardFit(deal[0], wc, we, wa)
		for i := 1; i < DealSize; i++ {
			if s := cardFit(deal[i], wc, we, wa); s > bestScore {
				best, bestScore = i, s
			}
		}
	default:
		// An unknown archetype takes the middle option rather than
		// panicking; GenerateSeason only ever assigns the four above.
		best = 0
	}
	return best
}

// areaUnits is the budget a card puts into one area, indexed as
// 0 chassis, 1 engine, 2 aero.
func areaUnits(c Card, area int) int {
	switch area {
	case 0:
		return c.Effect.Chassis
	case 1:
		return c.Effect.Engine
	default:
		return c.Effect.Aero
	}
}

// weightsForWindow sums the circuit weights of the races a window precedes.
func weightsForWindow(window int, cal []Circuit) (wc, we, wa Milli) {
	start := WindowRounds[window]
	end := len(cal)
	if window+1 < WindowCount {
		end = WindowRounds[window+1]
	}
	for i := start; i < end && i < len(cal); i++ {
		p := cal[i].Profile
		wc += p.Chassis
		we += p.Engine
		wa += p.Aero
	}
	return
}

// cardFit scores a card against a set of circuit weights: how much useful
// performance its units buy at the races just ahead.
func cardFit(c Card, wc, we, wa Milli) Milli {
	return Milli(c.Effect.Chassis)*wc + Milli(c.Effect.Engine)*we + Milli(c.Effect.Aero)*wa
}
