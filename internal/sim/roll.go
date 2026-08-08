package sim

import (
	"errors"
	"strconv"
)

// rollSalt keeps draft rolls on their own RNG stream, independent of season
// generation (seeded with seed) and race resolution (seed ^ raceSalt).
// Separate streams mean editing the roster cannot shift race outcomes, and
// changing generation cannot shift what you are offered. Draw order is the
// determinism contract, and this is how it stays local.
const (
	rollSalt = 0xCA4D
	raceSalt = 0x5EED
)

// ItemKind is one of the five things a rolled team-era offers. You take
// exactly one of them and then roll again.
type ItemKind int

const (
	ItemCar ItemKind = iota
	ItemDriverA
	ItemDriverB
	ItemEngineer
	ItemPrincipal
	itemKindCount
)

// Slot is where an item lands in your team. Both drivers fill the same
// slot, which has room for two.
type Slot int

const (
	SlotCar Slot = iota
	SlotDriver
	SlotEngineer
	SlotPrincipal
)

// slotFor maps an item to the slot it fills.
func slotFor(k ItemKind) Slot {
	switch k {
	case ItemCar:
		return SlotCar
	case ItemDriverA, ItemDriverB:
		return SlotDriver
	case ItemEngineer:
		return SlotEngineer
	default:
		return SlotPrincipal
	}
}

// slotCapacity is the shape of a complete team: one car, TWO drivers, one
// engineer, one principal. It sums to RollCount, which is why there is no
// passing -- every roll must fill something.
var slotCapacity = map[Slot]int{
	SlotCar:       1,
	SlotDriver:    CarsPerTeam,
	SlotEngineer:  1,
	SlotPrincipal: 1,
}

// RollsFor returns the team-eras offered at each roll. It is a pure
// function of the seed, so every player in the world is offered the same
// five teams on the same day and the server can re-derive them to verify a
// pick.
//
// The rolls are distinct: landing on 1988 McLaren twice would waste one of
// only five decisions.
func RollsFor(seed int64) [RollCount]TeamEra {
	rng := NewRNG(seed ^ rollSalt)

	// Copy before shuffling: the package-level roster must never be
	// mutated, or one call would change the next.
	pool := make([]TeamEra, len(Roster))
	copy(pool, Roster)

	var rolls [RollCount]TeamEra
	for i := 0; i < RollCount; i++ {
		j := i + rng.IntN(len(pool)-i)
		pool[i], pool[j] = pool[j], pool[i]
		rolls[i] = pool[i]
	}
	return rolls
}

// BuildLineup replays a set of picks against a set of rolls and returns the
// team they assemble.
//
// It is the whole validation surface for a submission, and it lives here in
// the sim rather than in the API so the browser and the server enforce
// identical rules. A pick is an index into the five items a rolled team
// offers, so a client cannot even describe an item it was never shown.
func BuildLineup(rolls [RollCount]TeamEra, picks []int) (Lineup, error) {
	if len(picks) != RollCount {
		return Lineup{}, errors.New("sim: got " + strconv.Itoa(len(picks)) +
			" picks, want " + strconv.Itoa(RollCount))
	}

	var out Lineup
	filled := map[Slot]int{}
	drivers := 0

	for i, p := range picks {
		if p < 0 || p >= int(itemKindCount) {
			return Lineup{}, errors.New("sim: roll " + strconv.Itoa(i+1) +
				" took item " + strconv.Itoa(p) + ", want 0.." +
				strconv.Itoa(int(itemKindCount)-1))
		}
		kind := ItemKind(p)
		slot := slotFor(kind)
		if filled[slot] >= slotCapacity[slot] {
			return Lineup{}, errors.New("sim: roll " + strconv.Itoa(i+1) +
				" fills an already-full slot")
		}
		filled[slot]++

		te := rolls[i]
		switch kind {
		case ItemCar:
			out.Car = te.Car
		case ItemDriverA:
			out.Drivers[drivers] = te.Drivers[0]
			drivers++
		case ItemDriverB:
			out.Drivers[drivers] = te.Drivers[1]
			drivers++
		case ItemEngineer:
			out.Engineer = te.Engineer
		case ItemPrincipal:
			out.Principal = te.Principal
		}
	}

	// Because slotCapacity sums to RollCount and no slot may overfill, a
	// picks slice of the right length that never overfills has necessarily
	// filled every slot. Checking anyway: this is the invariant the rest of
	// the simulation assumes, and it costs nothing to assert.
	for slot, cap := range slotCapacity {
		if filled[slot] != cap {
			return Lineup{}, errors.New("sim: slot " + strconv.Itoa(int(slot)) +
				" has " + strconv.Itoa(filled[slot]) + " of " + strconv.Itoa(cap))
		}
	}
	return out, nil
}
