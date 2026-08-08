// Command simulate runs one season from a seed and prints the result.
// It exists so the simulation can be exercised and eyeballed without a
// server, a database, or a browser.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jvonmikael/lights-out/internal/sim"
)

func main() {
	seed := flag.Int64("seed", 1, "season seed")
	strategy := flag.String("strategy", "adaptive", "player strategy: greedy, cautious, aerofirst, adaptive, first")
	asJSON := flag.Bool("json", false, "emit the full result as JSON")
	flag.Parse()

	season := sim.GenerateSeason(*seed)
	picks := sim.Strategy(*strategy, season)
	if picks == nil {
		fmt.Fprintf(os.Stderr, "unknown strategy %q\n", *strategy)
		os.Exit(2)
	}

	res, err := sim.RunSeason(*seed, picks)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("Seed %d · sim %s · strategy %s\n\n", *seed, res.SimVersion, *strategy)

	fmt.Printf("%-10s %-30s %-24s %4s\n", "SLOT", "TAKEN", "FROM", "OVR")
	fmt.Println(strings.Repeat("-", 72))
	rolls := res.Rolls
	for i, p := range res.Picks {
		te := rolls[i]
		var slot, name string
		var ovr int
		switch sim.ItemKind(p) {
		case sim.ItemCar:
			slot, name, ovr = "CAR", te.Car.Name, te.Car.Overall()
		case sim.ItemDriverA:
			slot, name, ovr = "DRIVER", te.Drivers[0].Name, te.Drivers[0].Overall()
		case sim.ItemDriverB:
			slot, name, ovr = "DRIVER", te.Drivers[1].Name, te.Drivers[1].Overall()
		case sim.ItemEngineer:
			slot, name, ovr = "ENGINEER", te.Engineer.Name, te.Engineer.Overall()
		case sim.ItemPrincipal:
			slot, name, ovr = "PRINCIPAL", te.Principal.Name, te.Principal.Overall()
		}
		fmt.Printf("%-10s %-30s %-24s %4d\n", slot, name, te.Label(), ovr)
	}
	fmt.Println()

	fmt.Printf("%-4s %-18s %-11s %5s %8s %6s\n", "RD", "CIRCUIT", "TYPE", "GRID", "FIN", "PTS")
	fmt.Println(strings.Repeat("-", 56))
	for i, r := range res.Races {
		var mine []sim.EntryResult
		for _, c := range r.Entries {
			if c.TeamID == 0 {
				mine = append(mine, c)
			}
		}
		player := mine[0]
		fin := ""
		pts := 0
		for _, c := range mine {
			if fin != "" {
				fin += "/"
			}
			if c.DNF {
				fin += "DNF"
			} else {
				fin += fmt.Sprintf("P%d", c.Finish)
			}
			pts += c.Points
		}
		sc := ""
		if r.SafetyCar {
			sc = " (SC)"
		}
		fmt.Printf("%-4d %-18s %-11s %5d %8s %6d%s\n",
			r.Round, r.Circuit, season.Calendar[i].Archetype, player.Grid, fin, pts, sc)
	}

	fmt.Printf("\n%-4s %-22s %6s %5s %8s %5s\n", "POS", "TEAM", "PTS", "WINS", "PODIUMS", "DNF")
	fmt.Println(strings.Repeat("-", 56))
	for i, s := range res.Standings {
		marker := " "
		if s.TeamID == 0 {
			marker = ">"
		}
		fmt.Printf("%s%-3d %-22s %6d %5d %8d %5d\n", marker, i+1, s.Name, s.Points, s.Wins, s.Podiums, s.DNFs)
	}

	fmt.Printf("\n%s\n", res.Share)
}
