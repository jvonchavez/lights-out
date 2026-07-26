package sim

// Every balance constant in the game lives in this file. M1 tunes these and
// nothing else; docs/Build Plan.md gates M1 on choosing the sigma values
// from 100k simulated seasons rather than guessing them.
const (
	RaceCount  = 10
	TeamCount  = 11  // the player plus 10 rivals
	RaceBudget = 100 // budget units available each race

	// Development happens in five windows rather than every race. Each
	// window deals DealSize cards and the player takes one, so a season is
	// five clicks. Card costs average ~200 units, keeping a season near the
	// 1000 units PressureQuad is calibrated against.
	WindowCount = 5
	DealSize    = 3

	MinCardCost = 140
	MaxCardCost = 260

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
	// The cap is what stops a pure aero car compounding without limit and
	// simply replacing engine as the single dominant answer. Aero is worth
	// buying up to the cap and much less past it, so the question becomes
	// what to take with the windows that remain.
	//
	// Raised from 150 to 310 for the card draft. With sliders a player
	// could pour unlimited budget into aero, so the cap had to be tight.
	// A deal of three cards constrains that on its own, and at 150 the
	// deepest strategy (aerofirst, 14.4% titles) barely separated from the
	// crudest one that simply buys the most (greedy, 13.9%). At 310 the
	// gradient reads 4.2 / 8.1 / 10.4 / 14.3 / 20.8 across the five
	// scripted strategies, and DNF rates are untouched because this lever
	// moves performance rather than risk.
	AeroScale    = Milli(12)
	MaxAeroBonus = Milli(310) // at most +31% performance

	SafetyCarChance    = Milli(250) // 25% per race
	SafetyCarThreshold = 2 * One    // cars within 2.0 pace may be shuffled

	// A safety rail, not an active constraint: it sits above the failure
	// chance a full-budget season produces (BaseFailure + PressureQuad) so
	// that clamping never silently truncates the risk model.
	MaxFailure = Milli(850) // never worse than 85%
	MinFailure = Milli(5)   // never better than 0.5%
)

// WindowRounds are the 0-based rounds a development window precedes: cards
// are dealt before races 1, 3, 5, 7 and 9, and the intervening races run on
// whatever the car already is.
var WindowRounds = [WindowCount]int{0, 2, 4, 6, 8}

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
