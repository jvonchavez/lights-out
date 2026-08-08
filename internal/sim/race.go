package sim

import (
	"cmp"
	"slices"
)

// runRace resolves one round in three phases, in a fixed draw order that
// must never change without bumping Version:
//
//  1. Reliability -- one draw per CAR in field order. A failure is a DNF
//     and the car scores nothing. This happens FIRST so the rest of the
//     race only simulates survivors.
//  2. Race pace -- one perf roll per SURVIVING car in field order, at a
//     lower sigma than qualifying because a race averages 50+ laps and
//     regresses toward true pace. Combined with grid position via the
//     circuit's overtaking difficulty and the driver's racecraft.
//  3. Events -- one safety car draw; if it fires, one shuffle pass that
//     compresses the field.
//
// The field is 24 cars, two per team, and the top ten of those 24 score.
// Both of a team's cars feed the same constructors' total, which is what
// makes the second driver a real decision rather than a passenger.
func runRace(field []*entryState, round int, c Circuit, grid []gridKey, r *RNG) RaceResult {
	gridPos := make(map[gridKey]int, len(grid))
	for i, k := range grid {
		gridPos[k] = i + 1 // 1-based
	}

	// Phase 1: reliability, in field order, one draw each.
	type outcome struct {
		dnf    bool
		reason string
	}
	fate := make(map[gridKey]outcome, len(field))
	for _, e := range field {
		dnf, reason := e.failure(r)
		fate[gridKey{e.teamID, e.entry}] = outcome{dnf, reason}
	}

	// Phase 2: race pace for survivors, in field order.
	type row struct {
		key   gridKey
		score Milli
	}
	var running []row
	for _, e := range field {
		k := gridKey{e.teamID, e.entry}
		if fate[k].dnf {
			continue
		}
		pace := e.perf(round, c.Profile, false, r)
		running = append(running, row{k, pace - e.gridPenalty(gridPos[k], c.Profile)})
	}

	slices.SortFunc(running, func(a, b row) int {
		if a.score != b.score {
			return cmp.Compare(b.score, a.score)
		}
		if a.key.teamID != b.key.teamID {
			return cmp.Compare(a.key.teamID, b.key.teamID)
		}
		return cmp.Compare(a.key.entry, b.key.entry)
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

	finish := make(map[gridKey]int, len(running))
	points := make(map[gridKey]int, len(running))
	for i, row := range running {
		finish[row.key] = i + 1
		if i < len(PointsTable) {
			points[row.key] = PointsTable[i]
		}
	}

	entries := make([]EntryResult, 0, len(field))
	for _, e := range field {
		k := gridKey{e.teamID, e.entry}
		entries = append(entries, EntryResult{
			TeamID:    e.teamID,
			Entry:     e.entry,
			DriverID:  e.driver.ID,
			Driver:    e.driver.Name,
			Grid:      gridPos[k],
			Finish:    finish[k], // zero value 0 means DNF
			DNF:       fate[k].dnf,
			DNFReason: fate[k].reason,
			Points:    points[k],
		})
	}
	// Always sorted by team then entry so the JSON is stable regardless of
	// outcome.
	slices.SortFunc(entries, func(a, b EntryResult) int {
		if a.TeamID != b.TeamID {
			return cmp.Compare(a.TeamID, b.TeamID)
		}
		return cmp.Compare(a.Entry, b.Entry)
	})

	return RaceResult{Circuit: c.Name, SafetyCar: safetyCar, Entries: entries}
}
