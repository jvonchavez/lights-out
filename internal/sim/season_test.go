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
	if len(s.Teams) != TeamCount {
		t.Errorf("field has %d teams, want %d", len(s.Teams), TeamCount)
	}
	if len(s.Budgets) != RaceCount {
		t.Errorf("budgets has %d entries, want %d", len(s.Budgets), RaceCount)
	}
	if s.Teams[0].ID != 0 || s.Teams[0].Archetype != "" {
		t.Error("team 0 must be the player, with no AI archetype")
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

func TestEveryTeamHasAnID(t *testing.T) {
	s := GenerateSeason(5)
	for i, tm := range s.Teams {
		if tm.ID != i {
			t.Errorf("team at index %d has ID %d", i, tm.ID)
		}
		if tm.Name == "" {
			t.Errorf("team %d has no name", i)
		}
		if i > 0 && tm.Archetype == "" {
			t.Errorf("rival %d has no archetype", i)
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
		if GenerateSeason(seed).Calendar[0].Name == GenerateSeason(seed+1000).Calendar[0].Name {
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
