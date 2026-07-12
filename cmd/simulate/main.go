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
	strategy := flag.String("strategy", "even", "player strategy: even, aggressive, specialist, reliability, idle")
	asJSON := flag.Bool("json", false, "emit the full result as JSON")
	flag.Parse()

	season := sim.GenerateSeason(*seed)
	decisions := sim.Strategy(*strategy, season)
	if decisions == nil {
		fmt.Fprintf(os.Stderr, "unknown strategy %q\n", *strategy)
		os.Exit(2)
	}

	res, err := sim.RunSeason(*seed, decisions)
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

	fmt.Printf("%-4s %-18s %-11s %5s %5s %6s\n", "RD", "CIRCUIT", "TYPE", "GRID", "FIN", "PTS")
	fmt.Println(strings.Repeat("-", 56))
	for i, r := range res.Races {
		var player sim.CarResult
		for _, c := range r.Cars {
			if c.TeamID == 0 {
				player = c
			}
		}
		fin := fmt.Sprintf("P%d", player.Finish)
		if player.DNF {
			fin = "DNF"
		}
		sc := ""
		if r.SafetyCar {
			sc = " (SC)"
		}
		fmt.Printf("%-4d %-18s %-11s %5d %5s %6d%s\n",
			r.Round, r.Circuit, season.Calendar[i].Archetype, player.Grid, fin, player.Points, sc)
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
