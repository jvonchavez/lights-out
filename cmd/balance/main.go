// Command balance runs many seasons across a worker pool and reports how
// each scripted player strategy performs. It is the M1 gate: no single
// strategy should win more than roughly 35% of seasons, and the qualifying
// and race sigma values are chosen from this data rather than guessed.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/jvonmikael/lights-out/internal/sim"
)

var strategies = []string{"even", "aggressive", "specialist", "adaptive", "aerofirst", "idle"}

type row struct {
	seed     int64
	strategy string
	points   int
	position int
	wins     int
	podiums  int
	dnfs     int
}

func main() {
	n := flag.Int("n", 100000, "number of seasons to simulate")
	out := flag.String("out", "balance.csv", "CSV output path")
	workers := flag.Int("workers", runtime.NumCPU(), "worker count")
	flag.Parse()

	start := time.Now()

	// Every worker runs ALL strategies against the same seed, so the
	// comparison between strategies is paired: they face an identical
	// season and an identical rival field, and the luck cancels out.
	seeds := make(chan int64, *workers*8)
	results := make(chan row, *workers*8)

	var g errgroup.Group
	for w := 0; w < *workers; w++ {
		g.Go(func() error {
			for seed := range seeds {
				season := sim.GenerateSeason(seed)
				for _, name := range strategies {
					res, err := sim.RunSeason(seed, sim.Strategy(name, season))
					if err != nil {
						return fmt.Errorf("seed %d strategy %s: %w", seed, name, err)
					}
					results <- row{
						seed: seed, strategy: name,
						points: res.Player.Points, position: res.PlayerPos,
						wins: res.Player.Wins, podiums: res.Player.Podiums,
						dnfs: res.Player.DNFs,
					}
				}
			}
			return nil
		})
	}

	go func() {
		for i := 0; i < *n; i++ {
			seeds <- int64(i)
		}
		close(seeds)
	}()

	done := make(chan struct{})
	var rows []row
	go func() {
		for r := range results {
			rows = append(rows, r)
		}
		close(done)
	}()

	if err := g.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	close(results)
	<-done

	elapsed := time.Since(start)

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	cw := csv.NewWriter(f)
	cw.Write([]string{"seed", "strategy", "points", "position", "wins", "podiums", "dnfs"})
	for _, r := range rows {
		cw.Write([]string{
			strconv.FormatInt(r.seed, 10), r.strategy,
			strconv.Itoa(r.points), strconv.Itoa(r.position),
			strconv.Itoa(r.wins), strconv.Itoa(r.podiums), strconv.Itoa(r.dnfs),
		})
	}
	cw.Flush()

	summarise(rows, *n, elapsed, *workers)
}

// summarise reports, per strategy, how often it won its season outright and
// how it performed on average. "Championship wins" counts seasons where the
// player finished P1 in the standings.
func summarise(rows []row, n int, elapsed time.Duration, workers int) {
	type agg struct {
		seasons, titles, points, positions, wins, podiums, dnfs int
	}
	byStrategy := map[string]*agg{}
	for _, r := range rows {
		a := byStrategy[r.strategy]
		if a == nil {
			a = &agg{}
			byStrategy[r.strategy] = a
		}
		a.seasons++
		if r.position == 1 {
			a.titles++
		}
		a.points += r.points
		a.positions += r.position
		a.wins += r.wins
		a.podiums += r.podiums
		a.dnfs += r.dnfs
	}

	names := make([]string, 0, len(byStrategy))
	for k := range byStrategy {
		names = append(names, k)
	}
	sort.Strings(names)

	fmt.Printf("%d seasons x %d strategies in %s across %d workers (%.0f seasons/sec)\n\n",
		n, len(strategies), elapsed.Round(time.Millisecond), workers,
		float64(n*len(strategies))/elapsed.Seconds())

	fmt.Printf("%-13s %8s %9s %9s %9s %9s\n", "STRATEGY", "TITLE%", "AVG PTS", "AVG POS", "AVG WINS", "AVG DNF")
	fmt.Println("---------------------------------------------------------------")
	worst := ""
	var maxTitle float64
	for _, name := range names {
		a := byStrategy[name]
		titlePct := 100 * float64(a.titles) / float64(a.seasons)
		if titlePct > maxTitle {
			maxTitle, worst = titlePct, name
		}
		fmt.Printf("%-13s %7.1f%% %9.1f %9.2f %9.2f %9.2f\n",
			name, titlePct,
			float64(a.points)/float64(a.seasons),
			float64(a.positions)/float64(a.seasons),
			float64(a.wins)/float64(a.seasons),
			float64(a.dnfs)/float64(a.seasons))
	}

	fmt.Printf("\nM1 GATE: best strategy is %q at %.1f%% titles (target: at most ~35%%) -- %s\n",
		worst, maxTitle, map[bool]string{true: "PASS", false: "FAIL"}[maxTitle <= 35.0])
}
