package sim

// Strategy returns the scripted player picks for a named strategy, or nil
// if the name is unknown. The balance harness runs all of them against the
// same seeds to check that no single approach dominates and that the
// gradient between the deepest and the crudest is wide enough for decisions
// to plainly matter; the simulate CLI uses them to play a season without a
// UI.
//
// They are ordered by how much they understand about the game. Every one of
// them produces a legal lineup, because the picker below only ever chooses
// among items whose slot still has room.
func Strategy(name string, season Season) []int {
	score, ok := strategyScorers[name]
	if !ok {
		return nil
	}

	picks := make([]int, RollCount)
	filled := map[Slot]int{}

	for i := 0; i < RollCount; i++ {
		te := season.Rolls[i]
		remaining := RollCount - i

		best, bestScore := -1, 0
		for k := ItemCar; k < itemKindCount; k++ {
			slot := slotFor(k)
			if filled[slot] >= slotCapacity[slot] {
				continue
			}
			// Never take an item that leaves a slot unfillable. With five
			// rolls and five slot-places this only bites on the last roll,
			// but it is what makes every scripted line legal by
			// construction rather than by luck.
			if !fillable(filled, slot, remaining) {
				continue
			}
			if s := score(te, k); best < 0 || s > bestScore {
				best, bestScore = int(k), s
			}
		}
		picks[i] = best
		filled[slotFor(ItemKind(best))]++
	}
	return picks
}

// fillable reports whether taking a place in slot leaves enough rolls to
// fill everything else.
func fillable(filled map[Slot]int, slot Slot, remaining int) bool {
	need := 0
	for s, cap := range slotCapacity {
		short := cap - filled[s]
		if s == slot {
			short--
		}
		if short > 0 {
			need += short
		}
	}
	return need <= remaining-1
}

// strategyScorers rank the items a rolled team offers. Ties fall to the
// lowest item index, because the loop above only replaces on a strict
// improvement -- so every scripted line is a total order.
var strategyScorers = map[string]func(TeamEra, ItemKind) int{
	// The no-thought baseline: always take the first legal item. Replaces
	// the old "spend nothing", which a draft makes inexpressible -- you
	// must take something from every roll.
	"first": func(TeamEra, ItemKind) int { return 0 },

	// Take the highest-rated thing available, whatever it is. The obvious
	// line, and the one a deeper strategy has to beat.
	"bestavailable": itemOverall,

	// The car is worth more than any single driver, so grab it early and
	// take the best of what is left. Tests whether "the car is everything"
	// is as true here as it is in Formula 1.
	"carfirst": func(te TeamEra, k ItemKind) int {
		if k == ItemCar {
			return 1000
		}
		return itemOverall(te, k)
	},

	// Chase the names. Two great drivers in a mediocre car -- the line
	// every fantasy player tries first.
	"starpower": func(te TeamEra, k ItemKind) int {
		if k == ItemDriverA || k == ItemDriverB {
			return 1000 + itemOverall(te, k)
		}
		return itemOverall(te, k)
	},

	// Value finishing over pace: reliability, a clean operation, a composed
	// driver. Ten races is not many to recover a retirement in.
	"cautious": func(te TeamEra, k ItemKind) int {
		return abovePar(k, rawCaution(te, k), cautionPar)
	},
}

func itemOverall(te TeamEra, k ItemKind) int {
	return abovePar(k, rawOverall(te, k), overallPar)
}

func rawOverall(te TeamEra, k ItemKind) int {
	switch k {
	case ItemCar:
		return te.Car.Overall()
	case ItemDriverA:
		return te.Drivers[0].Overall()
	case ItemDriverB:
		return te.Drivers[1].Overall()
	case ItemEngineer:
		return te.Engineer.Overall()
	default:
		return te.Principal.Overall()
	}
}

func rawCaution(te TeamEra, k ItemKind) int {
	switch k {
	case ItemCar:
		return te.Car.Reliability
	case ItemDriverA:
		return te.Drivers[0].Composure
	case ItemDriverB:
		return te.Drivers[1].Composure
	case ItemEngineer:
		return te.Engineer.Ops
	default:
		return te.Principal.Development
	}
}

// A greedy picker comparing raw numbers across five different kinds of item
// is not comparing like with like: car Reliability averages in the
// seventies across the roster while principal Development averages in the
// mid-eighties, so a naive "take the most reliable thing" never takes a car
// at all and ends up with whichever chassis the last roll happened to
// offer. Scoring an item by how far it sits above the roster mean for its
// own kind fixes that, and is what makes each scripted line actually do
// what its name says.
//
// Par is grouped by SLOT, not by item. The two drivers a team offers are
// the same kind of thing, and giving them separate pars would quietly bias
// every line toward the second seat -- lead drivers rate higher on average,
// so a per-item par makes the number two look like the better buy.
func abovePar(k ItemKind, raw int, par [4]int) int {
	return raw - par[slotFor(k)]
}

func parOf(raw func(TeamEra, ItemKind) int) [4]int {
	var sums, counts, out [4]int
	for _, te := range Roster {
		for k := ItemCar; k < itemKindCount; k++ {
			s := slotFor(k)
			sums[s] += raw(te, k)
			counts[s]++
		}
	}
	for s := range out {
		out[s] = sums[s] / counts[s]
	}
	return out
}

var (
	overallPar = parOf(rawOverall)
	cautionPar = parOf(rawCaution)
)

// StrategyNames are every scripted line, ordered from crudest to deepest.
var StrategyNames = []string{"first", "cautious", "starpower", "carfirst", "bestavailable"}
