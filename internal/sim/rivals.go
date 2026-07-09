package sim

// rivalDecision returns a rival's allocation for one round. It is a pure
// function of (archetype, round, calendar, budget) and NEVER sees the
// player. docs/Game Design.md is explicit about why: an AI that reacted to
// the player would make the daily seed unfair, because two players making
// different early choices would face different fields.
func rivalDecision(t Team, round int, cal []Circuit, budget int) Decision {
	var d Decision
	switch t.Archetype {
	case "aggressive":
		// Spends heavily and early: high DNF rate, high ceiling.
		if round < 4 {
			d = split(budget, 40, 40, 20, 0)
		} else {
			d = split(budget, 24, 24, 12, 40)
		}
	case "conservative":
		// Spreads evenly. Rarely fails, rarely wins.
		d = split(budget, 25, 25, 25, 25)
	case "specialist":
		// Pours everything into one area, dominant where it suits.
		switch t.ID % 3 {
		case 0:
			d = split(budget, 80, 0, 0, 20)
		case 1:
			d = split(budget, 0, 80, 0, 20)
		default:
			d = split(budget, 0, 0, 80, 20)
		}
	case "reactive":
		// Invests toward whichever archetype dominates the next two
		// circuits. Clamped at the end of the calendar.
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
			d = split(budget, 25, 25, 25, 25)
			break
		}
		perf := budget * 80 / 100
		d = Decision{
			Chassis:     perf * int(wc) / int(sum),
			Engine:      perf * int(we) / int(sum),
			Aero:        perf * int(wa) / int(sum),
			Reliability: budget - perf,
		}
	default:
		d = split(budget, 25, 25, 25, 25)
	}

	// Integer division above can leave the budget under-spent. Assign the
	// remainder to chassis so nothing is silently dropped and every rival
	// spends exactly its budget.
	if rem := budget - d.Total(); rem > 0 {
		d.Chassis += rem
	}
	return d
}

// split allocates budget by percentage. The percentages should sum to 100;
// any rounding shortfall is corrected by the caller.
func split(budget, chassis, engine, aero, reliability int) Decision {
	return Decision{
		Chassis:     budget * chassis / 100,
		Engine:      budget * engine / 100,
		Aero:        budget * aero / 100,
		Reliability: budget * reliability / 100,
	}
}
