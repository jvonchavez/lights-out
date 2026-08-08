package sim

// CircuitProfile weights the three performance areas. Chassis, Engine and
// Aero sum to One. A car's Cornering feeds Chassis, its Power feeds Engine,
// and its Aero feeds Aero.
type CircuitProfile struct {
	Chassis            Milli `json:"chassis"`
	Engine             Milli `json:"engine"`
	Aero               Milli `json:"aero"`
	OvertakeDifficulty Milli `json:"overtake_difficulty"`
}

// Circuit is one round of the calendar.
type Circuit struct {
	Name      string         `json:"name"`
	Archetype string         `json:"archetype"`
	Profile   CircuitProfile `json:"profile"`
}

// Lineup is a complete team: one car, two drivers, one race engineer and
// one team principal. The player assembles theirs one item at a time across
// five rolls; the 2026 grid arrives with theirs already built.
type Lineup struct {
	Car       CarSpec       `json:"car"`
	Drivers   [2]DriverSpec `json:"drivers"`
	Engineer  EngineerSpec  `json:"engineer"`
	Principal PrincipalSpec `json:"principal"`
}

// Team is one entrant. Team 0 is always the player; 1..TeamCount-1 are the
// 2026 grid in Grid2026 order.
type Team struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Livery string `json:"livery"`
	Lineup Lineup `json:"lineup"`
}

// Season is the public descriptor the client needs to render the game. It
// is derived entirely from the seed.
//
// There are no rival ratings to draw any more: the field is Grid2026 and is
// identical on every seed, so the only thing generation randomises is the
// calendar order and the five team-eras the player is offered.
type Season struct {
	Seed       int64              `json:"seed"`
	SimVersion string             `json:"sim_version"`
	Calendar   []Circuit          `json:"calendar"`
	Rolls      [RollCount]TeamEra `json:"rolls"`
	Rivals     []Team             `json:"rivals"`
}

// DNF reasons. A car that finishes carries the empty string.
const (
	DNFMechanical = "mechanical"
	DNFDriver     = "driver"
)

// EntryResult is one CAR's outcome in one race -- there are two per team.
// Finish is 1-based and is 0 when the car did not finish.
type EntryResult struct {
	TeamID    int    `json:"team_id"`
	Entry     int    `json:"entry"` // 0 or 1 within the team
	DriverID  string `json:"driver_id"`
	Driver    string `json:"driver"`
	Grid      int    `json:"grid"`
	Finish    int    `json:"finish"`
	DNF       bool   `json:"dnf"`
	DNFReason string `json:"dnf_reason"`
	Points    int    `json:"points"`
}

// RaceResult is one round. Entries is always sorted by team ID then entry
// index so the JSON is stable regardless of finishing order.
type RaceResult struct {
	Round     int           `json:"round"`
	Circuit   string        `json:"circuit"`
	SafetyCar bool          `json:"safety_car"`
	Entries   []EntryResult `json:"entries"`
}

// Standing is a constructor's season total: both cars' points added
// together, which is the whole reason the second driver is a real decision.
type Standing struct {
	TeamID  int    `json:"team_id"`
	Name    string `json:"name"`
	Points  int    `json:"points"`
	Wins    int    `json:"wins"`
	Podiums int    `json:"podiums"`
	DNFs    int    `json:"dnfs"`
}

// DriverStanding is the secondary table. It decides nothing -- the goal is
// the constructors' championship -- but seeing which of your two drivers
// carried the team is most of the fun of having two.
type DriverStanding struct {
	DriverID string `json:"driver_id"`
	Name     string `json:"name"`
	TeamID   int    `json:"team_id"`
	Points   int    `json:"points"`
	Wins     int    `json:"wins"`
}

// SeasonResult is the output of RunSeason. Standings are sorted by points,
// then wins, then podiums, then team ID, which is a total order.
type SeasonResult struct {
	SimVersion string `json:"sim_version"`
	Seed       int64  `json:"seed"`
	// Rolls and Picks together are the whole record of the draft: what was
	// offered, and what was taken from each. Lineup is what they produced.
	Rolls  [RollCount]TeamEra `json:"rolls"`
	Picks  []int              `json:"picks"`
	Lineup Lineup             `json:"lineup"`

	Races     []RaceResult     `json:"races"`
	Standings []Standing       `json:"standings"`
	Drivers   []DriverStanding `json:"drivers"`
	Player    Standing         `json:"player"`
	PlayerPos int              `json:"player_position"`
	Share     string           `json:"share"`
}
