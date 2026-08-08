package sim

// Every balance constant in the game lives in this file.
//
// SCALE. Roster ratings are 0-99 plain integers and map into the Milli
// fixed-point space with FromInt, so a 95-rated car corner-weights at 95.0.
// The circuit profile weights sum to One, so a car's weighted performance
// lands in roughly the 55..97 band -- the same order of magnitude the
// sigma values below were tuned against under the old rating model.
//
// RESCALING RULE. Ratings are plain counts, not Milli. Multiplying a rating
// by a per-point rate is plain integer multiplication with NO /One; only
// multiplying two Milli values uses .Mul. The two cases look alike and
// confusing them silently scales the whole game wrong.
const (
	RaceCount   = 10
	TeamCount   = 12 // the player plus the eleven-team 2026 grid
	CarsPerTeam = 2
	FieldSize   = TeamCount * CarsPerTeam // 24 cars

	// RollCount is the whole input surface: five rolls, five slots, one
	// item taken per roll and no passing. It is a constant here rather
	// than a literal so cmd/balance can sweep it.
	RollCount = 5

	QualiSigma = Milli(2500) // 2.5 rating points
	RaceSigma  = Milli(1500) // lower: a race averages 50+ laps and regresses

	// Driver pace is measured from a baseline rather than absolutely, so
	// the numbers below read as "rating points above a backmarker". The
	// spread across the roster is about 8 rating points per driver and so
	// about 16 across a two-car team -- deliberately smaller than the ~36
	// points separating the best car from the worst, because in Formula 1
	// the car dominates, and a draft where the car did not would be a lie.
	PaceBaseline = 60
	PaceRate     = 250

	// Staff ratings share one baseline. Setup applies in qualifying only
	// and Strategy in the race only, which is what makes an engineer worth
	// more to a car that qualifies badly than to one that does not.
	StaffBaseline  = 60
	SetupRate      = 80
	StrategyRate   = 110
	LeadershipRate = 120

	// Nerve applies only in the closing rounds, where a championship is
	// actually decided.
	NerveRate   = 100
	NerveRounds = 3

	// Development is the only thing that changes the car after lights out
	// in round one, because everything else was locked in at the draft. A
	// principal at DevBaseline stands still; Newey at 95 adds about four
	// rating points by the final round, which is the difference between
	// fourth and a win.
	DevBaseline = 70
	DevRate     = 18

	// Consistency widens or narrows a driver's own performance sigma. A
	// fast erratic driver and a slower metronome can carry the same
	// Overall and play completely differently across ten races.
	ConsistencyBaseline = 82
	ConsistencyRate     = 20
	MinSigmaFactor      = Milli(600)  // never less than 0.6x
	MaxSigmaFactor      = Milli(1600) // never more than 1.6x

	// Racecraft buys back a share of the grid-position penalty, capped so
	// that no driver can start last and be unaffected by it.
	RacecraftRate = 11
	MaxGridRelief = Milli(500) // at most 50% of the penalty

	// Failure. Two independent causes share ONE rng draw per car per race
	// (see team.go), so the reel can name a cause for free. Both are
	// computed at ten times scale and divided down, which keeps the slopes
	// expressible as integers.
	//
	// Mechanical risk is QUADRATIC in the car's shortfall from perfect
	// reliability, not linear. A linear slope cannot do both jobs at once:
	// fitted to make a 2026 car reliable it makes a 1967 car reliable too,
	// and fitted to make a 1967 car fragile it retires the modern grid four
	// times a season. Measured at the linear settings, Cadillac averaged
	// 3.99 retirements per season -- 20% of starts -- which is not a
	// Formula 1 team, it is attrition. The curve below gives roughly 2% a
	// start at Reliability 95, 6% at 75, 16% at 55 and 21% at 48, so the
	// modern grid finishes and the era trade-off the roster is built on
	// stays real.
	RelCeiling = 100
	MechFloor  = Milli(18)
	MechQuad   = 7
	// The engineer's operation fixes a FRACTION of the risk rather than a
	// fixed amount: a great engineer is worth much more on a fragile car
	// than on one that was never going to break.
	MechOpsRate    = 10
	DriverErrBase  = Milli(524)
	ComposureRate  = 12 // per point of driver Composure above StaffBaseline
	FailureDivisor = 10

	SafetyCarChance    = Milli(250) // 25% per race
	SafetyCarThreshold = 2 * One    // cars within 2.0 pace may be shuffled

	// Safety rails, not active constraints.
	MaxFailure = Milli(850) // never worse than 85%
	MinFailure = Milli(5)   // never better than 0.5%
)

// PointsTable is the standard allocation for the top ten finishers, now
// applied across a 24-car field rather than an 11-car one.
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
