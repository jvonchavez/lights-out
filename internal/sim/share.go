package sim

import (
	"strconv"
	"strings"
)

// shareString renders a season as three lines: a header, a summary, and an
// emoji row conveying the shape of the season at a glance -- including the
// one race that went wrong. It deliberately reveals nothing about which
// strategy produced it, so sharing a result never spoils the puzzle.
//
//	Lights Out · Season 142
//	P2 of 12 · 287 pts · 1 DNF
//	🏁🥈🥇🥉🏁🥇✖️🥈🥇🥈
//
// With two cars per team the row reports the team's BEST result each round,
// because the goal is the constructors' championship and that is the number
// the round actually contributed. A cross means both cars retired.
func shareString(seed int64, player Standing, pos int, races []RaceResult) string {
	// A short human-readable label, not the seed itself.
	label := seed % 1000
	if label < 0 {
		label = -label
	}

	var b strings.Builder
	b.WriteString("Lights Out · Season ")
	b.WriteString(strconv.FormatInt(label, 10))
	b.WriteByte('\n')

	// "P3 of 11", not "P3". A share string has to carry its own
	// denominator to be legible to someone who has never played -- which
	// is what makes "71-11" work and a bare position not.
	b.WriteByte('P')
	b.WriteString(strconv.Itoa(pos))
	b.WriteString(" of ")
	b.WriteString(strconv.Itoa(TeamCount))
	b.WriteString(" · ")
	b.WriteString(strconv.Itoa(player.Points))
	b.WriteString(" pts")
	if player.DNFs > 0 {
		b.WriteString(" · ")
		b.WriteString(strconv.Itoa(player.DNFs))
		b.WriteString(" DNF")
		if player.DNFs > 1 {
			b.WriteByte('s')
		}
	}
	b.WriteByte('\n')

	for _, r := range races {
		b.WriteString(raceEmoji(r))
	}
	return b.String()
}

func raceEmoji(r RaceResult) string {
	best, retired, ran := 0, 0, 0
	for _, c := range r.Entries {
		if c.TeamID != 0 {
			continue
		}
		ran++
		if c.DNF {
			retired++
			continue
		}
		if best == 0 || c.Finish < best {
			best = c.Finish
		}
	}
	if ran > 0 && retired == ran {
		return "✖️"
	}
	switch {
	case best == 0:
		return "⬜"
	case best == 1:
		return "🥇"
	case best == 2:
		return "🥈"
	case best == 3:
		return "🥉"
	case best <= len(PointsTable):
		return "🏁"
	default:
		return "⬜"
	}
}
