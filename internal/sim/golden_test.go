package sim

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden fixtures")

// goldenCases are committed seed/decision pairs whose full expected result
// lives in testdata/golden. If a rules change alters any of them, this test
// fails loudly -- which is the entire point. Regenerate with -update ONLY
// when the change was intended, and bump sim.Version when you do.
var goldenCases = []struct {
	name      string
	seed      int64
	decisions []Decision
}{
	{"seed-1001-even", 1001, repeatDecision(Decision{25, 25, 25, 25})},
	{"seed-2002-aggressive", 2002, frontLoaded()},
	{"seed-3003-specialist", 3003, repeatDecision(Decision{Engine: 80, Reliability: 20})},
	{"seed-4004-reliability", 4004, repeatDecision(Decision{20, 20, 20, 40})},
	{"seed-5005-idle", 5005, repeatDecision(Decision{})},
}

func repeatDecision(d Decision) []Decision {
	ds := make([]Decision, RaceCount)
	for i := range ds {
		ds[i] = d
	}
	return ds
}

// frontLoaded spends everything on performance early, then nothing.
func frontLoaded() []Decision {
	ds := make([]Decision, RaceCount)
	for i := range ds {
		if i < 3 {
			ds[i] = Decision{Chassis: 40, Engine: 40, Aero: 20}
		}
	}
	return ds
}

func TestGolden(t *testing.T) {
	for _, tc := range goldenCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RunSeason(tc.seed, tc.decisions)
			if err != nil {
				t.Fatalf("RunSeason: %v", err)
			}
			encoded, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			path := filepath.Join("testdata", "golden", tc.name+".json")
			if *update {
				if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("wrote %s", path)
				return
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading fixture: %v (run with -update to create it)", err)
			}
			if string(append(encoded, '\n')) != string(want) {
				t.Errorf("result differs from the committed fixture %s.\n"+
					"If this change was intended, bump sim.Version and rerun with -update.", path)
			}
		})
	}
}
