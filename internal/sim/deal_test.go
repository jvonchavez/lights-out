package sim

import (
	"reflect"
	"testing"
)

func TestDealsAreDeterministic(t *testing.T) {
	a, b := DealsFor(4242), DealsFor(4242)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed produced different deals")
	}
}

func TestDealsHaveNoDuplicatesWithinAWindow(t *testing.T) {
	for seed := int64(0); seed < 200; seed++ {
		for w, deal := range DealsFor(seed) {
			seen := map[string]bool{}
			for _, c := range deal {
				if seen[c.ID] {
					t.Fatalf("seed %d window %d deals %q twice", seed, w, c.ID)
				}
				seen[c.ID] = true
			}
		}
	}
}

func TestEveryDealtCardIsInThePool(t *testing.T) {
	inPool := map[string]bool{}
	for _, c := range CardPool {
		inPool[c.ID] = true
	}
	for seed := int64(0); seed < 100; seed++ {
		for _, deal := range DealsFor(seed) {
			for _, c := range deal {
				if !inPool[c.ID] {
					t.Fatalf("seed %d dealt %q, which is not in the pool", seed, c.ID)
				}
			}
		}
	}
}

func TestDifferentSeedsDealDifferently(t *testing.T) {
	same := 0
	for seed := int64(0); seed < 100; seed++ {
		a := DealsFor(seed)[0][0].ID
		b := DealsFor(seed + 5000)[0][0].ID
		if a == b {
			same++
		}
	}
	if same > 40 {
		t.Errorf("%d/100 seed pairs opened with the same card -- dealing is not random enough", same)
	}
}

func TestDealStreamIsSeparateFromRaceStream(t *testing.T) {
	// The behavioural property -- that editing the card pool cannot move
	// race outcomes -- lives in run_test.go, where RunSeason is available.
	// This is the structural half: the two streams must not coincide.
	if dealSalt == raceSalt {
		t.Fatal("deal and race RNG streams share a salt")
	}
	for seed := int64(0); seed < 50; seed++ {
		d := NewRNG(seed ^ dealSalt).Uint64()
		r := NewRNG(seed ^ raceSalt).Uint64()
		if d == r {
			t.Fatalf("seed %d: deal and race streams opened identically", seed)
		}
	}
}

func TestDealsForCoversEveryWindow(t *testing.T) {
	deals := DealsFor(7)
	if len(deals) != WindowCount {
		t.Fatalf("%d deals, want %d", len(deals), WindowCount)
	}
	for w, d := range deals {
		if len(d) != DealSize {
			t.Errorf("window %d deals %d cards, want %d", w, len(d), DealSize)
		}
	}
}
