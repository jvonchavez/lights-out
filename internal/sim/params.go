package sim

// Every balance constant in the game lives in this file. M1 tunes these and
// nothing else; docs/Build Plan.md gates M1 on choosing the sigma values
// from 100k simulated seasons rather than guessing them.
const (
	RaceCount  = 10
	TeamCount  = 11  // the player plus 10 rivals
	RaceBudget = 100 // budget units available each race

	StartRating = 50 * One // every team starts at 50.0 in each area
	StartSpread = 4 * One  // rivals vary around it by this sigma

	// GainPerUnit is the rating bought by one budget unit. Spends are plain
	// counts, so this multiplies directly with no rescale: 100 units buys
	// 100 * 100 = 10000 millis, a full +10.0.
	GainPerUnit = Milli(100)

	QualiSigma = Milli(2500) // 2.5 rating points -- M1 GATE
	RaceSigma  = Milli(1500) // lower: a race averages 50+ laps and regresses

	BaseFailure = Milli(30) // 3.0% before any development

	// PressureQuad is the failure probability added at FULL season spend.
	// Risk grows with the SQUARE of cumulative development, not linearly.
	//
	// This is an M1 finding, not a preference. Under the linear model the
	// design document specifies, risk and performance are both linear in
	// spend, so one always dominates globally and there is no interior
	// optimum: measured across pressure coefficients from 250 to 850,
	// spending 100% of the budget was optimal every time, even at 4.19
	// DNFs per 10-race season. A convex cost is what makes "how much
	// performance will you buy with how much risk" a real question.
	PressureQuad   = Milli(280)
	AeroEfficiency = Milli(300) // aero offsets 30% of the pressure it causes

	// Aero "scales with both" (docs/Game Design.md): rating above the
	// baseline multiplies the whole weighted performance sum rather than
	// only adding through its own circuit weight. AeroScale sets the rate,
	// MaxAeroBonus caps it.
	//
	// The cap is what makes the sliders a real decision. Without it, a pure
	// aero car compounds without limit and simply replaces engine as the
	// dominant single answer. With it, aero is worth buying up to the cap
	// and worthless past it, so the interesting question becomes what to do
	// with the budget that remains.
	AeroScale    = Milli(12)
	MaxAeroBonus = Milli(150) // at most +25% performance

	SafetyCarChance    = Milli(250) // 25% per race
	SafetyCarThreshold = 2 * One    // cars within 2.0 pace may be shuffled

	// A safety rail, not an active constraint: it sits above the failure
	// chance a full-budget season produces (BaseFailure + PressureQuad) so
	// that clamping never silently truncates the risk model.
	MaxFailure = Milli(850) // never worse than 85%
	MinFailure = Milli(5)   // never better than 0.5%
)

// PointsTable is the standard allocation for the top ten finishers.
var PointsTable = [10]int{25, 18, 15, 12, 10, 8, 6, 4, 2, 1}

// Profiles maps circuit archetype to its weights. This is a map, which is
// safe only because it is indexed by key and never ranged over in a path
// that affects results -- map iteration order is randomised in Go and would
// destroy determinism. TestSimHasNoIOImports and the determinism tests
// guard the rule.
var Profiles = map[string]CircuitProfile{
	"power":     {Chassis: 200, Engine: 550, Aero: 250, OvertakeDifficulty: 120},
	"technical": {Chassis: 550, Engine: 200, Aero: 250, OvertakeDifficulty: 450},
	"balanced":  {Chassis: 350, Engine: 350, Aero: 300, OvertakeDifficulty: 280},
	"highspeed": {Chassis: 300, Engine: 350, Aero: 350, OvertakeDifficulty: 200},
}
