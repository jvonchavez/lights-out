package sim

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden fixtures")

// goldenCases are committed seed/pick pairs whose full expected result
// lives in testdata/golden. If a rules change alters any of them, this test
// fails loudly -- which is the entire point. Regenerate with -update ONLY
// when the change was intended, and bump sim.Version when you do.
var goldenCases = []struct {
	name  string
	seed  int64
	picks []int
}{
	{"seed-1001-inorder", 1001, []int{0, 1, 2, 3, 4}},
	{"seed-2002-reversed", 2002, []int{4, 3, 2, 1, 0}},
	{"seed-3003-drivers-last", 3003, []int{0, 3, 4, 1, 2}},
	{"seed-4004-car-late", 4004, []int{1, 4, 2, 3, 0}},
	{"seed-5005-bestavailable", 5005, Strategy("bestavailable", GenerateSeason(5005))},
}

func TestGolden(t *testing.T) {
	for _, tc := range goldenCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RunSeason(tc.seed, tc.picks)
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
