package sim

import "testing"

func TestRNGDeterministic(t *testing.T) {
	a, b := NewRNG(42), NewRNG(42)
	for i := 0; i < 1000; i++ {
		if x, y := a.Uint64(), b.Uint64(); x != y {
			t.Fatalf("draw %d diverged: %d != %d", i, x, y)
		}
	}
}

func TestRNGDifferentSeedsDiverge(t *testing.T) {
	a, b := NewRNG(1), NewRNG(2)
	if a.Uint64() == b.Uint64() {
		t.Fatal("seeds 1 and 2 produced the same first draw")
	}
}

func TestNormalIsCenteredAndScaled(t *testing.T) {
	r := NewRNG(7)
	const n = 100000
	var sum, sumSq int64
	for i := 0; i < n; i++ {
		v := int64(r.Normal(FromInt(2))) // sigma = 2.0
		sum += v
		sumSq += v * v
	}
	mean := sum / n
	if mean < -60 || mean > 60 {
		t.Errorf("mean = %d millis, want near 0", mean)
	}
	variance := sumSq/n - mean*mean
	if variance < 3_600_000 || variance > 4_400_000 {
		t.Errorf("variance = %d, want approx 4000000", variance)
	}
}

func TestNormalIsBounded(t *testing.T) {
	r := NewRNG(11)
	sigma := FromInt(2)
	for i := 0; i < 100000; i++ {
		v := r.Normal(sigma)
		if v < -6*sigma || v > 6*sigma {
			t.Fatalf("draw %d = %d, outside +/-6 sigma", i, v)
		}
	}
}

func TestIntNInRange(t *testing.T) {
	r := NewRNG(3)
	for i := 0; i < 10000; i++ {
		if v := r.IntN(11); v < 0 || v >= 11 {
			t.Fatalf("IntN(11) = %d, out of range", v)
		}
	}
}

func TestIntNIsUniform(t *testing.T) {
	r := NewRNG(23)
	counts := make([]int, 10)
	const n = 200000
	for i := 0; i < n; i++ {
		counts[r.IntN(10)]++
	}
	for i, c := range counts {
		if c < n/10-2000 || c > n/10+2000 {
			t.Errorf("bucket %d got %d, want approx %d", i, c, n/10)
		}
	}
}

func TestChanceBounds(t *testing.T) {
	r := NewRNG(9)
	for i := 0; i < 100; i++ {
		if r.Chance(0) {
			t.Fatal("Chance(0) must never fire")
		}
		if !r.Chance(One) {
			t.Fatal("Chance(One) must always fire")
		}
	}
}

// TestRNGFrozenStream pins the generator's output forever. If this fails,
// the RNG changed and every historical leaderboard is invalidated. The fix
// is to revert the change, or to bump sim.Version deliberately and accept
// that old seasons are frozen under the old version. Never edit want[] to
// make this pass.
func TestRNGFrozenStream(t *testing.T) {
	r := NewRNG(12345)
	want := []uint64{
		1411482639,
		3165192603,
		3360792183,
		2433038347,
		628889468,
	}
	for i, w := range want {
		if got := r.Uint64(); got != w {
			t.Errorf("draw %d = %d, want %d", i, got, w)
		}
	}
}
