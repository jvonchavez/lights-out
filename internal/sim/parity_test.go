//go:build parity

package sim

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"
)

// TestNativeWASMParity is the architectural centrepiece's guard rail.
//
// internal/sim compiles to two targets: natively into the server binary,
// where it computes the authoritative leaderboard score, and to js/wasm,
// where the browser plays the season locally with no network round-trip.
// There is one implementation, so the client and server cannot disagree
// about the rules -- provided the two builds actually produce identical
// output. This asserts that over several thousand seeds.
//
// It lives behind a build tag so `go test ./...` stays fast. Run it with
// `make parity`, which rebuilds sim.wasm first.
//
// If this ever fails, the cause is almost certainly floating point or a
// stdlib RNG that slipped into the package -- Go's assembly implementations
// of log/exp/cos differ from the pure-Go ones used on js/wasm. The fix is
// integer arithmetic in the affected calculation, never a tolerance.
func TestNativeWASMParity(t *testing.T) {
	const seeds = 3000

	type parityCase struct {
		Seed  int64 `json:"seed"`
		Picks []int `json:"picks"`
	}

	// Rotate through every scripted strategy so parity is exercised across
	// the whole space of card choices, not just one shape.
	names := []string{"greedy", "cautious", "aerofirst", "adaptive", "first"}
	cases := make([]parityCase, 0, seeds)
	for i := 0; i < seeds; i++ {
		seed := int64(i) * 7919 // spread seeds across the range
		season := GenerateSeason(seed)
		cases = append(cases, parityCase{
			Seed:  seed,
			Picks: Strategy(names[i%len(names)], season),
		})
	}

	input, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("node", "../../scripts/parity_runner.mjs")
	cmd.Stdin = bytes.NewReader(input)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("wasm runner failed: %v\n%s", err, stderr.String())
	}

	lines := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
	if len(lines) != len(cases) {
		t.Fatalf("got %d wasm results, want %d", len(lines), len(cases))
	}

	for i, c := range cases {
		native, err := RunSeason(c.Seed, c.Picks)
		if err != nil {
			t.Fatalf("case %d (seed %d): native error %v", i, c.Seed, err)
		}
		// Both sides marshal with encoding/json from the same struct
		// definitions, so field order and formatting match and a byte
		// comparison is meaningful.
		want, err := json.Marshal(native)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(want, lines[i]) {
			t.Fatalf("PARITY DIVERGENCE at case %d (seed %d, strategy %s)\nnative: %s\nwasm:   %s",
				i, c.Seed, names[i%len(names)], want, lines[i])
		}
	}

	t.Logf("native and js/wasm agreed byte-for-byte across %d seasons", len(cases))
}
