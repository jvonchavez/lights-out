package sim

// The roster is the game's content: real Formula 1 team-seasons, each
// carrying the four things a team is made of. A roll lands on one of these
// and the player takes exactly one item from it, so a team-era is a hand of
// five things you can have one of -- and a great season offers a great car
// AND a great driver AND a great principal, which is where the whole
// decision comes from (docs/Game Design.md).
//
// Ratings are 0-99 sports-game integers. Overall is COMPUTED from the
// sub-ratings and never authored, following the rule the card draft
// established for risk pips: what the player is shown is derived from what
// the simulation applies, so the two cannot drift.
//
// Every display string lives in a single Name field per entity. Renaming
// the entire roster is therefore a content edit that touches no simulation
// code.

// CarSpec is a chassis. Power, Cornering and Aero feed the three circuit
// profile weights directly; Reliability drives mechanical failure.
type CarSpec struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Power       int    `json:"power"`       // -> CircuitProfile.Engine
	Cornering   int    `json:"cornering"`   // -> CircuitProfile.Chassis
	Aero        int    `json:"aero"`        // -> CircuitProfile.Aero
	Reliability int    `json:"reliability"` // -> mechanical failure chance
}

// Overall weights the three performance ratings equally and reliability at
// a third of one of them. Reliability matters enormously in a 10-race
// season, but it is a floor rather than a way to win, so it must not let a
// slow, unbreakable car read as a good one.
func (c CarSpec) Overall() int {
	return (30*c.Power + 30*c.Cornering + 30*c.Aero + 10*c.Reliability) / 100
}

// DriverSpec is one driver. Consistency is the interesting one: it narrows
// or widens this driver's own performance sigma, so a fast, erratic driver
// and a slower metronome can carry the same Overall and play completely
// differently across ten races.
type DriverSpec struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Pace        int    `json:"pace"`        // raw speed, added to performance
	Racecraft   int    `json:"racecraft"`   // reduces the grid-position penalty
	Consistency int    `json:"consistency"` // narrows this driver's sigma
	Composure   int    `json:"composure"`   // reduces driver-error DNFs
}

func (d DriverSpec) Overall() int {
	return (40*d.Pace + 25*d.Racecraft + 20*d.Consistency + 15*d.Composure) / 100
}

// EngineerSpec is the team's lead race engineer -- or, for team-eras before
// named race engineers existed in the record, the chief designer or
// technical director who filled that role. Setup applies in qualifying
// only, Strategy in the race only, which is what makes an engineer worth
// more to a car that qualifies badly than to one that does not.
type EngineerSpec struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Setup    int    `json:"setup"`    // qualifying-only performance
	Strategy int    `json:"strategy"` // race-only performance
	Ops      int    `json:"ops"`      // reduces mechanical failure
}

func (e EngineerSpec) Overall() int {
	return (35*e.Setup + 40*e.Strategy + 25*e.Ops) / 100
}

// PrincipalSpec is the team boss. Because the whole team is locked in
// before round one, the principal is the only thing that changes the car
// during the season: Development adds rating every round, so a mid car
// under a great principal out-develops a great car under a weak one by the
// closing rounds.
type PrincipalSpec struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Development int    `json:"development"` // per-round car rating gain
	Leadership  int    `json:"leadership"`  // added driver pace
	Nerve       int    `json:"nerve"`       // extra pace in the final rounds
}

func (p PrincipalSpec) Overall() int {
	return (45*p.Development + 30*p.Leadership + 25*p.Nerve) / 100
}

// TeamEra is one team in one season: the unit a roll lands on.
type TeamEra struct {
	ID        string        `json:"id"`
	Team      string        `json:"team"`
	Year      int           `json:"year"`
	EraID     string        `json:"era_id"`
	Livery    string        `json:"livery"` // hex, for the UI only
	Car       CarSpec       `json:"car"`
	Drivers   [2]DriverSpec `json:"drivers"`
	Engineer  EngineerSpec  `json:"engineer"`
	Principal PrincipalSpec `json:"principal"`
}

// Label is how a roll reads on screen: "1988 McLaren".
func (t TeamEra) Label() string { return itoa(t.Year) + " " + t.Team }

// Era groups team-eras by the technical regulations they raced under. The
// draft does not use eras mechanically -- they exist so the roster can be
// checked for coverage and so the UI can say what it is showing you.
type Era struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Years string `json:"years"`
}

// Eras are in chronological order.
var Eras = []Era{
	{ID: "1950s", Name: "Front-Engine", Years: "1950-1960"},
	{ID: "1960s", Name: "Rear-Engine", Years: "1961-1970"},
	{ID: "1970s", Name: "Ground Effect", Years: "1971-1981"},
	{ID: "1980s", Name: "Turbo", Years: "1982-1988"},
	{ID: "1990s", Name: "Electronics and the Ban", Years: "1989-1999"},
	{ID: "2000s", Name: "V10 Dominance", Years: "2000-2009"},
	{ID: "2010s", Name: "Blown Floors and Hybrids", Years: "2010-2019"},
	{ID: "2020s", Name: "Ground Effect II", Years: "2020-2025"},
	{ID: "2026", Name: "The Current Grid", Years: "2026"},
}

// Roster is every rollable team-era, historical entries followed by the
// current grid. This is a slice, not a map, because it is ranged over in a
// path that affects results -- map iteration order is randomised in Go and
// would destroy determinism.
var Roster = func() []TeamEra {
	all := make([]TeamEra, 0, 64)
	all = append(all, roster1950s...)
	all = append(all, roster1960s...)
	all = append(all, roster1970s...)
	all = append(all, roster1980s...)
	all = append(all, roster1990s...)
	all = append(all, roster2000s...)
	all = append(all, roster2010s...)
	all = append(all, roster2020s...)
	all = append(all, Grid2026...)
	return all
}()

// itoa avoids importing strconv, which would pull unicode tables into a
// package whose whole point is having no dependencies worth speaking of.
// Years are always four positive digits.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [8]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
