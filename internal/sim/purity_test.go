package sim

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// TestSimHasNoIOImports enforces the central architectural constraint:
// internal/sim is a pure package. No I/O, no network, no clock, no logging,
// no unseeded randomness. docs/Architecture.md depends on this, because the
// same package compiles to js/wasm for the browser, and anything in this
// list either fails to compile there, behaves differently there, or makes
// the function non-deterministic.
//
// This test uses os/exec, which is fine: _test.go files are not part of the
// package's own build and never reach the WASM target.
func TestSimHasNoIOImports(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Fatalf("go list failed: %v", err)
	}
	deps := strings.Fields(string(out))

	banned := []string{
		"os", "os/exec",
		"net", "net/http",
		"time",
		"log", "log/slog",
		"math/rand", "math/rand/v2",
		"crypto/rand",
		"io/ioutil",
	}
	for _, b := range banned {
		if slices.Contains(deps, b) {
			t.Errorf("internal/sim depends on %q -- the sim must stay pure (docs/Architecture.md)", b)
		}
	}
}
