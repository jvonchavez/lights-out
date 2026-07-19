package scheduler

import (
	"testing"
	"time"
)

func TestSeedForDateIsStable(t *testing.T) {
	day := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	a, b := SeedForDate(day), SeedForDate(day)
	if a != b {
		t.Errorf("same date gave %d and %d", a, b)
	}
	// Time of day must not matter -- only the calendar date.
	noon := time.Date(2026, 9, 2, 12, 34, 56, 789, time.UTC)
	if SeedForDate(noon) != a {
		t.Error("seed changed with the time of day")
	}
}

func TestSeedDiffersAcrossDates(t *testing.T) {
	seen := map[int64]string{}
	for i := 0; i < 400; i++ {
		day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
		s := SeedForDate(day)
		if prev, dup := seen[s]; dup {
			t.Errorf("seed %d collides between %s and %s", s, prev, day.Format("2006-01-02"))
		}
		seen[s] = day.Format("2006-01-02")
	}
}

func TestSeedFitsInJavaScriptSafeRange(t *testing.T) {
	// Above 2^53 a JS number loses precision and the browser would play a
	// different season from the one the server verifies.
	for i := 0; i < 5000; i++ {
		day := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
		s := SeedForDate(day)
		if s < 0 || s > maxSafeInt {
			t.Fatalf("%s produced seed %d, outside [0, 2^53-1]", day.Format("2006-01-02"), s)
		}
		if float64(s) != float64(int64(float64(s))) {
			t.Fatalf("seed %d does not survive a float64 round trip", s)
		}
	}
}

func TestDayTruncatesToUTCMidnight(t *testing.T) {
	got := Day(time.Date(2026, 9, 2, 23, 59, 59, 999, time.FixedZone("x", 3*3600)))
	want := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Day = %s, want %s", got, want)
	}
}
