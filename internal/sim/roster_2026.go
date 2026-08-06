package sim

// Grid2026 is the opposition: the eleven real 2026 teams, exactly as they
// are. They do not draft and they do not vary by seed. Every player, on
// every seed, races this same field, so a season measures the draft and
// nothing else -- which is a stronger property for a shared daily seed than
// the procedurally generated rivals it replaces (docs/Game Design.md).
//
// Team IDs are the player at 0 and this slice at 1..11, in the order below.
//
// Ratings are anchored to the real 2026 constructors' table after twelve
// rounds -- Mercedes 425, Ferrari 338, McLaren 263, Red Bull 186, then a
// cliff to Racing Bulls 66 and Alpine 63, and a long tail down to Cadillac
// on nothing. The regulation reset scrambled the order, and the roster
// should say so: the power unit is the story of 2026, which is why the
// Mercedes-engined cars carry high Power regardless of where they finish
// and Red Bull's first year of its own unit costs it a title it would
// otherwise have been in.
//
// The 2026 entries are rollable like any other team-era. Taking McLaren's
// car means racing a McLaren with a hole in it, which is the best moment
// the mechanic can produce and costs nothing to allow.
var Grid2026 = []TeamEra{
	{
		ID: "mercedes-2026", Team: "Mercedes", Year: 2026, EraID: "2026", Livery: "#00D7B6",
		Car: CarSpec{ID: "mercedes-w17", Name: "Mercedes F1 W17",
			Power: 97, Cornering: 93, Aero: 94, Reliability: 95},
		Drivers: [2]DriverSpec{
			{ID: "russell-2026", Name: "George Russell", Pace: 93, Racecraft: 89, Consistency: 91, Composure: 87},
			{ID: "antonelli-2026", Name: "Kimi Antonelli", Pace: 88, Racecraft: 82, Consistency: 80, Composure: 78},
		},
		Engineer:  EngineerSpec{ID: "bonnington-2026", Name: "Peter Bonnington", Setup: 91, Strategy: 90, Ops: 89},
		Principal: PrincipalSpec{ID: "wolff-2026", Name: "Toto Wolff", Development: 93, Leadership: 90, Nerve: 88},
	},
	{
		ID: "ferrari-2026", Team: "Ferrari", Year: 2026, EraID: "2026", Livery: "#E8002D",
		Car: CarSpec{ID: "ferrari-sf26", Name: "Ferrari SF-26",
			Power: 93, Cornering: 92, Aero: 90, Reliability: 88},
		Drivers: [2]DriverSpec{
			{ID: "leclerc-2026", Name: "Charles Leclerc", Pace: 96, Racecraft: 88, Consistency: 86, Composure: 84},
			{ID: "hamilton-2026", Name: "Lewis Hamilton", Pace: 89, Racecraft: 94, Consistency: 85, Composure: 90},
		},
		Engineer:  EngineerSpec{ID: "bozzi-2026", Name: "Bryan Bozzi", Setup: 85, Strategy: 84, Ops: 84},
		Principal: PrincipalSpec{ID: "vasseur-2026", Name: "Fred Vasseur", Development: 87, Leadership: 86, Nerve: 84},
	},
	{
		ID: "mclaren-2026", Team: "McLaren", Year: 2026, EraID: "2026", Livery: "#FF8000",
		Car: CarSpec{ID: "mclaren-mcl40", Name: "McLaren MCL40",
			Power: 94, Cornering: 87, Aero: 86, Reliability: 86},
		Drivers: [2]DriverSpec{
			{ID: "norris-2026", Name: "Lando Norris", Pace: 94, Racecraft: 90, Consistency: 89, Composure: 85},
			{ID: "piastri-2026", Name: "Oscar Piastri", Pace: 93, Racecraft: 89, Consistency: 92, Composure: 88},
		},
		Engineer:  EngineerSpec{ID: "will-joseph-2026", Name: "Will Joseph", Setup: 88, Strategy: 89, Ops: 86},
		Principal: PrincipalSpec{ID: "stella-2026", Name: "Andrea Stella", Development: 91, Leadership: 92, Nerve: 88},
	},
	{
		ID: "redbull-2026", Team: "Red Bull", Year: 2026, EraID: "2026", Livery: "#3671C6",
		Car: CarSpec{ID: "redbull-rb22", Name: "Red Bull RB22",
			Power: 78, Cornering: 90, Aero: 89, Reliability: 80},
		Drivers: [2]DriverSpec{
			{ID: "verstappen-2026", Name: "Max Verstappen", Pace: 99, Racecraft: 97, Consistency: 96, Composure: 92},
			{ID: "hadjar-2026", Name: "Isack Hadjar", Pace: 85, Racecraft: 80, Consistency: 79, Composure: 77},
		},
		Engineer:  EngineerSpec{ID: "lambiase-2026", Name: "Gianpiero Lambiase", Setup: 90, Strategy: 92, Ops: 88},
		Principal: PrincipalSpec{ID: "mekies-2026", Name: "Laurent Mekies", Development: 84, Leadership: 83, Nerve: 80},
	},
	{
		ID: "racingbulls-2026", Team: "Racing Bulls", Year: 2026, EraID: "2026", Livery: "#6692FF",
		Car: CarSpec{ID: "racingbulls-vcarb03", Name: "Racing Bulls VCARB 03",
			Power: 76, Cornering: 79, Aero: 77, Reliability: 78},
		Drivers: [2]DriverSpec{
			{ID: "lawson-2026", Name: "Liam Lawson", Pace: 80, Racecraft: 78, Consistency: 76, Composure: 74},
			{ID: "lindblad-2026", Name: "Arvid Lindblad", Pace: 79, Racecraft: 74, Consistency: 71, Composure: 72},
		},
		Engineer:  EngineerSpec{ID: "iliopoulos-2026", Name: "Alexandre Iliopoulos", Setup: 79, Strategy: 80, Ops: 78},
		Principal: PrincipalSpec{ID: "permane-2026", Name: "Alan Permane", Development: 80, Leadership: 79, Nerve: 78},
	},
	{
		ID: "alpine-2026", Team: "Alpine", Year: 2026, EraID: "2026", Livery: "#00A1E8",
		Car: CarSpec{ID: "alpine-a526", Name: "Alpine A526",
			Power: 90, Cornering: 70, Aero: 69, Reliability: 76},
		Drivers: [2]DriverSpec{
			{ID: "gasly-2026", Name: "Pierre Gasly", Pace: 84, Racecraft: 83, Consistency: 82, Composure: 78},
			{ID: "colapinto-2026", Name: "Franco Colapinto", Pace: 79, Racecraft: 76, Consistency: 72, Composure: 72},
		},
		Engineer:  EngineerSpec{ID: "peckett-2026", Name: "Josh Peckett", Setup: 78, Strategy: 78, Ops: 76},
		Principal: PrincipalSpec{ID: "nielsen-2026", Name: "Steve Nielsen", Development: 76, Leadership: 78, Nerve: 76},
	},
	{
		ID: "haas-2026", Team: "Haas", Year: 2026, EraID: "2026", Livery: "#B6BABD",
		Car: CarSpec{ID: "haas-vf26", Name: "Haas VF-26",
			Power: 88, Cornering: 65, Aero: 63, Reliability: 72},
		Drivers: [2]DriverSpec{
			{ID: "ocon-2026", Name: "Esteban Ocon", Pace: 82, Racecraft: 82, Consistency: 83, Composure: 76},
			{ID: "bearman-2026", Name: "Oliver Bearman", Pace: 82, Racecraft: 79, Consistency: 76, Composure: 76},
		},
		Engineer:  EngineerSpec{ID: "muller-2026", Name: "Laura Mueller", Setup: 78, Strategy: 77, Ops: 79},
		Principal: PrincipalSpec{ID: "komatsu-2026", Name: "Ayao Komatsu", Development: 78, Leadership: 82, Nerve: 78},
	},
	{
		ID: "audi-2026", Team: "Audi", Year: 2026, EraID: "2026", Livery: "#8E9A9A",
		Car: CarSpec{ID: "audi-r26", Name: "Audi R26",
			Power: 72, Cornering: 70, Aero: 69, Reliability: 66},
		Drivers: [2]DriverSpec{
			{ID: "hulkenberg-2026", Name: "Nico Hulkenberg", Pace: 83, Racecraft: 85, Consistency: 84, Composure: 85},
			{ID: "bortoleto-2026", Name: "Gabriel Bortoleto", Pace: 80, Racecraft: 77, Consistency: 78, Composure: 78},
		},
		Engineer:  EngineerSpec{ID: "petrik-2026", Name: "Steven Petrik", Setup: 76, Strategy: 76, Ops: 75},
		Principal: PrincipalSpec{ID: "wheatley-2026", Name: "Jonathan Wheatley", Development: 79, Leadership: 80, Nerve: 84},
	},
	{
		ID: "williams-2026", Team: "Williams", Year: 2026, EraID: "2026", Livery: "#1868DB",
		Car: CarSpec{ID: "williams-fw48", Name: "Williams FW48",
			Power: 90, Cornering: 60, Aero: 58, Reliability: 62},
		Drivers: [2]DriverSpec{
			{ID: "albon-2026", Name: "Alex Albon", Pace: 84, Racecraft: 84, Consistency: 86, Composure: 84},
			{ID: "sainz-2026", Name: "Carlos Sainz", Pace: 88, Racecraft: 86, Consistency: 85, Composure: 83},
		},
		Engineer:  EngineerSpec{ID: "urwin-2026", Name: "James Urwin", Setup: 82, Strategy: 83, Ops: 80},
		Principal: PrincipalSpec{ID: "vowles-2026", Name: "James Vowles", Development: 86, Leadership: 85, Nerve: 82},
	},
	{
		ID: "astonmartin-2026", Team: "Aston Martin", Year: 2026, EraID: "2026", Livery: "#229971",
		Car: CarSpec{ID: "astonmartin-amr26", Name: "Aston Martin AMR26",
			Power: 74, Cornering: 64, Aero: 66, Reliability: 58},
		Drivers: [2]DriverSpec{
			{ID: "alonso-2026", Name: "Fernando Alonso", Pace: 85, Racecraft: 96, Consistency: 88, Composure: 93},
			{ID: "stroll-2026", Name: "Lance Stroll", Pace: 74, Racecraft: 72, Consistency: 72, Composure: 70},
		},
		Engineer: EngineerSpec{ID: "cronin-2026", Name: "Chris Cronin", Setup: 80, Strategy: 79, Ops: 77},
		// Newey's Development is the highest number on the current grid and
		// the team is tenth on three points. That gap is the argument for
		// making Development a season-long slope rather than a flat bonus.
		Principal: PrincipalSpec{ID: "newey-2026", Name: "Adrian Newey", Development: 95, Leadership: 80, Nerve: 82},
	},
	{
		ID: "cadillac-2026", Team: "Cadillac", Year: 2026, EraID: "2026", Livery: "#D4AF37",
		Car: CarSpec{ID: "cadillac-mac26", Name: "Cadillac MAC-26",
			Power: 86, Cornering: 52, Aero: 50, Reliability: 55},
		Drivers: [2]DriverSpec{
			{ID: "bottas-2026", Name: "Valtteri Bottas", Pace: 80, Racecraft: 82, Consistency: 85, Composure: 86},
			{ID: "perez-2026", Name: "Sergio Perez", Pace: 80, Racecraft: 83, Consistency: 74, Composure: 75},
		},
		Engineer:  EngineerSpec{ID: "howard-2026", Name: "John Howard", Setup: 74, Strategy: 73, Ops: 72},
		Principal: PrincipalSpec{ID: "lowdon-2026", Name: "Graeme Lowdon", Development: 74, Leadership: 79, Nerve: 76},
	},
}
