package sim

import (
	"cmp"
	"slices"
)

// qualify ranks the field by one performance roll and returns team IDs in
// grid order, pole first.
//
// The roll happens once per car in TEAM ID ORDER. That order is part of the
// determinism contract: changing it changes every downstream result.
func qualify(cars []*carState, p CircuitProfile, r *RNG) []int {
	type entry struct {
		teamID int
		score  Milli
	}
	entries := make([]entry, len(cars))
	for i, c := range cars {
		entries[i] = entry{teamID: c.team.ID, score: c.perf(p, QualiSigma, r)}
	}

	// A TOTAL order: ties break on team ID, never on slice position, so the
	// result cannot depend on how the field happened to be assembled.
	slices.SortFunc(entries, func(a, b entry) int {
		if a.score != b.score {
			return cmp.Compare(b.score, a.score) // higher is better
		}
		return cmp.Compare(a.teamID, b.teamID)
	})

	grid := make([]int, len(entries))
	for i, e := range entries {
		grid[i] = e.teamID
	}
	return grid
}
