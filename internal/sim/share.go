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
//	P2 · 287 pts · 1 DNF
//	🏁🥈🥇🥉🏁🥇✖️🥈🥇🥈
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

	b.WriteByte('P')
	b.WriteString(strconv.Itoa(pos))
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
	for _, c := range r.Cars {
		if c.TeamID != 0 {
			continue
		}
		switch {
		case c.DNF:
			return "✖️"
		case c.Finish == 1:
			return "🥇"
		case c.Finish == 2:
			return "🥈"
		case c.Finish == 3:
			return "🥉"
		case c.Finish <= len(PointsTable):
			return "🏁"
		default:
			return "⬜"
		}
	}
	return "⬜"
}
