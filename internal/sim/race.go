package sim

import (
	"cmp"
	"slices"
)

// runRace resolves one round in three phases, in a fixed draw order that
// must never change without bumping Version:
//
//  1. Reliability -- one Chance per car in team ID order. A failure is a
//     DNF and the car scores nothing. This happens FIRST so the rest of the
//     race only simulates survivors.
//  2. Race pace -- one perf roll per SURVIVING car in team ID order, at a
//     lower sigma than qualifying because a race averages 50+ laps and
//     regresses toward true pace. Combined with grid position via the
//     circuit's overtaking difficulty.
//  3. Events -- one safety car draw; if it fires, one shuffle pass that
//     compresses the field.
func runRace(cars []*carState, c Circuit, grid []int, r *RNG) RaceResult {
	gridPos := make(map[int]int, len(grid))
	for i, id := range grid {
		gridPos[id] = i + 1 // 1-based
	}

	// Phase 1: reliability, in team ID order.
	dnf := make(map[int]bool, len(cars))
	for _, car := range cars {
		if r.Chance(car.failureChance()) {
			dnf[car.team.ID] = true
		}
	}

	// Phase 2: race pace for survivors, in team ID order.
	type entry struct {
		teamID int
		score  Milli
	}
	var running []entry
	for _, car := range cars {
		if dnf[car.team.ID] {
			continue
		}
		pace := car.perf(c.Profile, RaceSigma, r)
		// A bad grid slot costs more where overtaking is hard. At a power
		// circuit a fast car recovers; at a technical one it stays stuck.
		//
		// FromInt, not Milli: grid position is a plain count and must be
		// SCALED to fixed-point before multiplying. Milli() would treat 10
		// as 0.01 and make the penalty vanish against the noise.
		penalty := FromInt(gridPos[car.team.ID] - 1).Mul(c.Profile.OvertakeDifficulty)
		running = append(running, entry{teamID: car.team.ID, score: pace - penalty})
	}

	slices.SortFunc(running, func(a, b entry) int {
		if a.score != b.score {
			return cmp.Compare(b.score, a.score)
		}
		return cmp.Compare(a.teamID, b.teamID)
	})

	// Phase 3: the safety car compresses the field, partially randomising
	// order among cars already close together. A deliberate equaliser that
	// keeps a dominant season from becoming boring.
	safetyCar := r.Chance(SafetyCarChance)
	if safetyCar {
		for i := 0; i+1 < len(running); i++ {
			if running[i].score-running[i+1].score <= SafetyCarThreshold && r.Chance(One/2) {
				running[i], running[i+1] = running[i+1], running[i]
			}
		}
	}

	results := make([]CarResult, 0, len(cars))
	finish := make(map[int]int, len(running))
	points := make(map[int]int, len(running))
	for i, e := range running {
		finish[e.teamID] = i + 1
		if i < len(PointsTable) {
			points[e.teamID] = PointsTable[i]
		}
	}
	for _, car := range cars {
		id := car.team.ID
		results = append(results, CarResult{
			TeamID: id,
			Grid:   gridPos[id],
			Finish: finish[id], // zero value 0 means DNF
			DNF:    dnf[id],
			Points: points[id],
		})
	}
	// Always sorted by team ID so the JSON is stable regardless of outcome.
	slices.SortFunc(results, func(a, b CarResult) int { return cmp.Compare(a.TeamID, b.TeamID) })

	return RaceResult{Circuit: c.Name, SafetyCar: safetyCar, Cars: results}
}
