package sim

// rivalDecision returns a rival's allocation for one round. It is a pure
// function of (archetype, round, calendar, budget) and NEVER sees the
// player. docs/Game Design.md is explicit about why: an AI that reacted to
// the player would make the daily seed unfair, because two players making
// different early choices would face different fields.
//
// Each archetype expresses a different bet about where performance comes
// from. None of them is strictly best; M1 measured the field at roughly
// 10-14% championships each, with the specialists as the deliberate trap.
func rivalDecision(t Team, round int, cal []Circuit, budget int) Decision {
	var d Decision
	switch t.Archetype {
	case "aggressive":
		// Pours everything into raw chassis and engine and neglects aero.
		// High DNF rate, high ceiling, and blind to the fact that aero
		// multiplies the whole car rather than only its own term.
		d = split(budget, 45, 45, 10)
	case "conservative":
		// Spreads evenly. Rarely fails spectacularly, rarely dominates.
		d = split(budget, 34, 33, 33)
	case "specialist":
		// Pours everything into one area, dominant at circuits that suit it
		// and badly exposed everywhere else.
		switch t.ID % 3 {
		case 0:
			d = split(budget, 100, 0, 0)
		case 1:
			d = split(budget, 0, 100, 0)
		default:
			d = split(budget, 0, 0, 100)
		}
	case "reactive":
		// Invests toward whichever archetype dominates the next two
		// circuits, clamped at the end of the calendar.
		var wc, we, wa Milli
		for off := 0; off < 2; off++ {
			i := round + off
			if i >= len(cal) {
				i = len(cal) - 1
			}
			p := cal[i].Profile
			wc += p.Chassis
			we += p.Engine
			wa += p.Aero
		}
		sum := wc + we + wa
		if sum == 0 { // unreachable with real profiles, but never divide by zero
			d = split(budget, 34, 33, 33)
			break
		}
		d = Decision{
			Chassis: budget * int(wc) / int(sum),
			Engine:  budget * int(we) / int(sum),
			Aero:    budget * int(wa) / int(sum),
		}
	default:
		d = split(budget, 34, 33, 33)
	}

	// Integer division above can leave the budget under-spent. Assign the
	// remainder to chassis so nothing is silently dropped and every rival
	// spends exactly its budget.
	if rem := budget - d.Total(); rem > 0 {
		d.Chassis += rem
	}
	return d
}

// split allocates budget by percentage across the three development areas.
// Any rounding shortfall is corrected by the caller.
func split(budget, chassis, engine, aero int) Decision {
	return Decision{
		Chassis: budget * chassis / 100,
		Engine:  budget * engine / 100,
		Aero:    budget * aero / 100,
	}
}
