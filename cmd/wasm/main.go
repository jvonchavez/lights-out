//go:build js && wasm

// Command wasm exposes the simulation to the browser. It is a thin binding
// layer and deliberately contains no rules: internal/sim is compiled here
// unchanged, so the browser and the server cannot disagree about the game.
package main

import (
	"encoding/json"
	"strconv"
	"syscall/js"

	"github.com/jvonmikael/lights-out/internal/sim"
)

func errJSON(msg string) string {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return string(b)
}

// runSeason(seedString, decisionsJSON) -> resultJSON, or {"error": ...}.
//
// The seed crosses as a STRING. JavaScript numbers are float64, so an int64
// seed above 2^53 would silently lose precision on the way in and the
// browser would play a different season from the one the server verifies.
func runSeason(this js.Value, args []js.Value) any {
	if len(args) != 2 {
		return errJSON("runSeason expects (seedString, decisionsJSON)")
	}
	seed, err := strconv.ParseInt(args[0].String(), 10, 64)
	if err != nil {
		return errJSON("bad seed: " + err.Error())
	}
	var ds []sim.Decision
	if err := json.Unmarshal([]byte(args[1].String()), &ds); err != nil {
		return errJSON("bad decisions: " + err.Error())
	}
	res, err := sim.RunSeason(seed, ds)
	if err != nil {
		return errJSON(err.Error())
	}
	out, err := json.Marshal(res)
	if err != nil {
		return errJSON(err.Error())
	}
	return string(out)
}

// generateSeason(seedString) -> seasonJSON. Lets the client render the
// calendar and field without a network round-trip.
func generateSeason(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return errJSON("generateSeason expects (seedString)")
	}
	seed, err := strconv.ParseInt(args[0].String(), 10, 64)
	if err != nil {
		return errJSON("bad seed: " + err.Error())
	}
	out, err := json.Marshal(sim.GenerateSeason(seed))
	if err != nil {
		return errJSON(err.Error())
	}
	return string(out)
}

func main() {
	js.Global().Set("lightsOutRunSeason", js.FuncOf(runSeason))
	js.Global().Set("lightsOutGenerateSeason", js.FuncOf(generateSeason))
	js.Global().Set("lightsOutVersion", sim.Version)
	js.Global().Set("lightsOutReady", true)
	select {} // keep the module alive for the page's lifetime
}
