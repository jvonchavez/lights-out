package sim

import (
	"cmp"
	"slices"
)

// gridKey identifies one car on the grid.
type gridKey struct {
	teamID int
	entry  int
}

// qualify ranks the whole 24-car field by one performance roll each and
// returns the grid, pole first.
//
// The roll happens once per car in FIELD ORDER -- team ID, then entry index
// within the team. That order is part of the determinism contract:
// changing it changes every downstream result.
func qualify(field []*entryState, round int, p CircuitProfile, r *RNG) []gridKey {
	type row struct {
		key   gridKey
		score Milli
	}
	rows := make([]row, len(field))
	for i, e := range field {
		rows[i] = row{
			key:   gridKey{e.teamID, e.entry},
			score: e.perf(round, p, true, r),
		}
	}

	// A TOTAL order: ties break on team ID then entry, never on slice
	// position, so the result cannot depend on how the field happened to be
	// assembled.
	slices.SortFunc(rows, func(a, b row) int {
		if a.score != b.score {
			return cmp.Compare(b.score, a.score) // higher is better
		}
		if a.key.teamID != b.key.teamID {
			return cmp.Compare(a.key.teamID, b.key.teamID)
		}
		return cmp.Compare(a.key.entry, b.key.entry)
	})

	grid := make([]gridKey, len(rows))
	for i, row := range rows {
		grid[i] = row.key
	}
	return grid
}
