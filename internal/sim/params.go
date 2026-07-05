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

	BaseFailure     = Milli(30)  // 3.0% before any development
	PressurePerUnit = Milli(250) // per 1000 cumulative perf spend -> +25%
	ReliabilityGain = Milli(400) // per 1000 spend on reliability -> -40%
	AeroEfficiency  = Milli(300) // aero offsets 30% of the pressure it causes

	SafetyCarChance    = Milli(250) // 25% per race
	SafetyCarThreshold = 2 * One    // cars within 2.0 pace may be shuffled

	MaxFailure = Milli(600) // never worse than 60%
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
