package sim

import "testing"

func TestCardPoolIsLargeEnoughToDeal(t *testing.T) {
	if len(CardPool) < DealSize*WindowCount {
		t.Errorf("pool has %d cards, want at least %d so no season repeats a deal",
			len(CardPool), DealSize*WindowCount)
	}
}

func TestCardCostsAreInRange(t *testing.T) {
	// Cards deliberately vary in cost: that variation is what makes the
	// risk pips mean something and keeps "how much performance will you buy
	// with how much risk" a real question. An average near 200 keeps a
	// five-window season around 1000 units, where PressureQuad is calibrated.
	total := 0
	for _, c := range CardPool {
		if c.Cost() < MinCardCost || c.Cost() > MaxCardCost {
			t.Errorf("%s costs %d, want %d..%d", c.ID, c.Cost(), MinCardCost, MaxCardCost)
		}
		total += c.Cost()
	}
	avg := total / len(CardPool)
	if avg < 185 || avg > 215 {
		t.Errorf("mean card cost %d, want near 200 so a season banks ~1000 units", avg)
	}
}

func TestCardEffectsAreNonNegative(t *testing.T) {
	for _, c := range CardPool {
		if c.Effect.Chassis < 0 || c.Effect.Engine < 0 || c.Effect.Aero < 0 {
			t.Errorf("%s has a negative allocation: %+v", c.ID, c.Effect)
		}
	}
}

func TestCardIDsAndNamesAreUnique(t *testing.T) {
	ids := map[string]bool{}
	names := map[string]bool{}
	for _, c := range CardPool {
		if c.ID == "" || c.Name == "" || c.Blurb == "" {
			t.Errorf("%q is missing an id, name or blurb", c.ID)
		}
		if ids[c.ID] {
			t.Errorf("duplicate id %q", c.ID)
		}
		if names[c.Name] {
			t.Errorf("duplicate name %q", c.Name)
		}
		ids[c.ID] = true
		names[c.Name] = true
	}
}

func TestEveryAreaIsBuildable(t *testing.T) {
	// If an area is never the dominant one on a card, no player can ever
	// commit to it and the circuit weights for that area become dead rules.
	dominant := map[string]int{}
	for _, c := range CardPool {
		e := c.Effect
		switch {
		case e.Chassis > e.Engine && e.Chassis > e.Aero:
			dominant["chassis"]++
		case e.Engine > e.Chassis && e.Engine > e.Aero:
			dominant["engine"]++
		case e.Aero > e.Chassis && e.Aero > e.Engine:
			dominant["aero"]++
		}
	}
	for _, area := range []string{"chassis", "engine", "aero"} {
		if dominant[area] < 3 {
			t.Errorf("only %d cards are dominantly %s, want at least 3", dominant[area], area)
		}
	}
}

func TestCardCostMatchesEffectTotal(t *testing.T) {
	for _, c := range CardPool {
		if c.Cost() != c.Effect.Total() {
			t.Errorf("%s: Cost() %d != Effect.Total() %d", c.ID, c.Cost(), c.Effect.Total())
		}
	}
}

func TestWindowRoundsAreValid(t *testing.T) {
	if len(WindowRounds) != WindowCount {
		t.Fatalf("%d window rounds, want %d", len(WindowRounds), WindowCount)
	}
	prev := -1
	for i, r := range WindowRounds {
		if r <= prev {
			t.Errorf("window rounds must ascend: index %d is round %d after %d", i, r, prev)
		}
		if r < 0 || r >= RaceCount {
			t.Errorf("window %d is round %d, outside the %d-race season", i, r, RaceCount)
		}
		prev = r
	}
}
