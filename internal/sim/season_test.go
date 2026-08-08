package sim

import (
	"reflect"
	"testing"
)

func TestGenerateSeasonIsDeterministic(t *testing.T) {
	a, b := GenerateSeason(999), GenerateSeason(999)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed produced different seasons")
	}
}

func TestGenerateSeasonShape(t *testing.T) {
	s := GenerateSeason(1)
	if len(s.Calendar) != RaceCount {
		t.Errorf("calendar has %d races, want %d", len(s.Calendar), RaceCount)
	}
	if len(s.Rivals) != TeamCount-1 {
		t.Errorf("field has %d rivals, want %d", len(s.Rivals), TeamCount-1)
	}
	if s.SimVersion != Version {
		t.Errorf("sim version = %q, want %q", s.SimVersion, Version)
	}
	if s.Seed != 1 {
		t.Errorf("seed = %d, want 1", s.Seed)
	}
	counts := map[string]int{}
	for _, c := range s.Calendar {
		counts[c.Archetype]++
	}
	want := map[string]int{"power": 3, "technical": 3, "balanced": 2, "highspeed": 2}
	if !reflect.DeepEqual(counts, want) {
		t.Errorf("archetype mix = %v, want %v", counts, want)
	}
}

// The rival field is a constant now, not a draw. Every player on every seed
// races the same 2026 grid, which is what makes the leaderboard a measure
// of the draft alone -- and it is worth pinning, because a stray rng call
// in GenerateSeason would silently make some days easier than others.
func TestRivalFieldIsIdenticalOnEverySeed(t *testing.T) {
	want := GenerateSeason(1).Rivals
	for seed := int64(2); seed < 200; seed++ {
		if got := GenerateSeason(seed).Rivals; !reflect.DeepEqual(got, want) {
			t.Fatalf("seed %d faces a different field from seed 1", seed)
		}
	}
}

func TestRivalsAreTheCurrentGrid(t *testing.T) {
	s := GenerateSeason(5)
	if len(s.Rivals) != len(Grid2026) {
		t.Fatalf("%d rivals, but the grid has %d teams", len(s.Rivals), len(Grid2026))
	}
	for i, tm := range s.Rivals {
		if tm.ID != i+1 {
			t.Errorf("rival at index %d has ID %d, want %d", i, tm.ID, i+1)
		}
		if tm.Name != Grid2026[i].Team {
			t.Errorf("rival %d is %q, want %q", i, tm.Name, Grid2026[i].Team)
		}
		if tm.Lineup.Car.ID != Grid2026[i].Car.ID {
			t.Errorf("%s is not driving its own car", tm.Name)
		}
	}
}

func TestCalendarHasNoDuplicates(t *testing.T) {
	s := GenerateSeason(17)
	seen := map[string]bool{}
	for _, c := range s.Calendar {
		if seen[c.Name] {
			t.Errorf("circuit %q appears twice", c.Name)
		}
		seen[c.Name] = true
	}
}

func TestDifferentSeedsShuffleTheCalendar(t *testing.T) {
	same := 0
	for seed := int64(0); seed < 50; seed++ {
		if GenerateSeason(seed).Calendar[0].Name == GenerateSeason(seed + 1000).Calendar[0].Name {
			same++
		}
	}
	if same > 20 {
		t.Errorf("%d/50 seed pairs shared round 1 -- calendar is not shuffling", same)
	}
}

func TestProfileWeightsSumToOne(t *testing.T) {
	for name, p := range Profiles {
		if sum := p.Chassis + p.Engine + p.Aero; sum != One {
			t.Errorf("profile %q weights sum to %d, want %d", name, sum, One)
		}
	}
}
