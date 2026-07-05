package sim

// Area names a development area. Reliability buys risk down without adding
// performance; docs/Game Design.md flags it as an M1 balance question.
type Area int

const (
	AreaChassis Area = iota
	AreaEngine
	AreaAero
	AreaReliability
	areaCount
)

// Decision is one race's budget allocation. Every value must be
// non-negative and the four must sum to at most that race's budget.
// Spends are plain unit counts, not Milli.
type Decision struct {
	Chassis     int `json:"chassis"`
	Engine      int `json:"engine"`
	Aero        int `json:"aero"`
	Reliability int `json:"reliability"`
}

// Total is the budget consumed by this allocation.
func (d Decision) Total() int { return d.Chassis + d.Engine + d.Aero + d.Reliability }

// CircuitProfile weights the three performance areas. Chassis, Engine and
// Aero sum to One.
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

// Ratings are a car's three performance ratings.
type Ratings struct {
	Chassis Milli `json:"chassis"`
	Engine  Milli `json:"engine"`
	Aero    Milli `json:"aero"`
}

// Team is one entrant. Team 0 is always the player and has no archetype.
type Team struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Archetype   string  `json:"archetype"`
	Start       Ratings `json:"start"`
	DriverSkill Milli   `json:"driver_skill"`
}

// Season is the public descriptor the client needs to render the game. It
// is derived entirely from the seed.
type Season struct {
	Seed       int64     `json:"seed"`
	SimVersion string    `json:"sim_version"`
	Calendar   []Circuit `json:"calendar"`
	Teams      []Team    `json:"teams"`
	Budgets    []int     `json:"budgets"`
}

// CarResult is one car's outcome in one race. Finish is 1-based and is 0
// when the car did not finish.
type CarResult struct {
	TeamID int  `json:"team_id"`
	Grid   int  `json:"grid"`
	Finish int  `json:"finish"`
	DNF    bool `json:"dnf"`
	Points int  `json:"points"`
}

// RaceResult is one round. Cars is always sorted by TeamID so the JSON is
// stable regardless of finishing order.
type RaceResult struct {
	Round     int         `json:"round"`
	Circuit   string      `json:"circuit"`
	SafetyCar bool        `json:"safety_car"`
	Cars      []CarResult `json:"cars"`
}

// Standing is a team's season total.
type Standing struct {
	TeamID  int `json:"team_id"`
	Name    string `json:"name"`
	Points  int `json:"points"`
	Wins    int `json:"wins"`
	Podiums int `json:"podiums"`
	DNFs    int `json:"dnfs"`
}

// SeasonResult is the output of RunSeason. Standings are sorted by points,
// then wins, then podiums, then team ID, which is a total order.
type SeasonResult struct {
	SimVersion string       `json:"sim_version"`
	Seed       int64        `json:"seed"`
	Races      []RaceResult `json:"races"`
	Standings  []Standing   `json:"standings"`
	Player     Standing     `json:"player"`
	PlayerPos  int          `json:"player_position"`
	Share      string       `json:"share"`
}
