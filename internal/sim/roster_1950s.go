package sim

// The 1950s. Cars are rated for how they stood against their own field --
// the Ferrari 500 won every race it entered in 1952 and reads like it --
// but RELIABILITY is on one absolute scale across the whole roster, and it
// is where the eras genuinely differ. A 1950s car in the fifties finished
// about half the races it started, and taking one here means accepting
// that. Raw pace from history, at the price of finishing.
var roster1950s = []TeamEra{
	{
		ID: "ferrari-1952", Team: "Ferrari", Year: 1952, EraID: "1950s", Livery: "#CC0000",
		Car: CarSpec{ID: "ferrari-500", Name: "Ferrari 500",
			Power: 88, Cornering: 94, Aero: 82, Reliability: 60},
		Drivers: [2]DriverSpec{
			{ID: "ascari-1952", Name: "Alberto Ascari", Pace: 95, Racecraft: 90, Consistency: 92, Composure: 88},
			{ID: "farina-1952", Name: "Nino Farina", Pace: 88, Racecraft: 84, Consistency: 80, Composure: 76},
		},
		Engineer:  EngineerSpec{ID: "lampredi-1952", Name: "Aurelio Lampredi", Setup: 84, Strategy: 80, Ops: 78},
		Principal: PrincipalSpec{ID: "enzo-1952", Name: "Enzo Ferrari", Development: 90, Leadership: 88, Nerve: 92},
	},
	{
		ID: "mercedes-1955", Team: "Mercedes", Year: 1955, EraID: "1950s", Livery: "#C8C8C8",
		Car: CarSpec{ID: "mercedes-w196", Name: "Mercedes W196",
			Power: 92, Cornering: 90, Aero: 88, Reliability: 72},
		Drivers: [2]DriverSpec{
			{ID: "fangio-1955", Name: "Juan Manuel Fangio", Pace: 97, Racecraft: 93, Consistency: 95, Composure: 94},
			{ID: "moss-1955", Name: "Stirling Moss", Pace: 94, Racecraft: 91, Consistency: 86, Composure: 85},
		},
		Engineer:  EngineerSpec{ID: "uhlenhaut-1955", Name: "Rudolf Uhlenhaut", Setup: 88, Strategy: 84, Ops: 86},
		Principal: PrincipalSpec{ID: "neubauer-1955", Name: "Alfred Neubauer", Development: 86, Leadership: 90, Nerve: 88},
	},
	{
		ID: "maserati-1957", Team: "Maserati", Year: 1957, EraID: "1950s", Livery: "#B01020",
		Car: CarSpec{ID: "maserati-250f", Name: "Maserati 250F",
			Power: 84, Cornering: 88, Aero: 78, Reliability: 55},
		Drivers: [2]DriverSpec{
			{ID: "fangio-1957", Name: "Juan Manuel Fangio", Pace: 97, Racecraft: 94, Consistency: 94, Composure: 95},
			{ID: "behra-1957", Name: "Jean Behra", Pace: 82, Racecraft: 80, Consistency: 74, Composure: 72},
		},
		Engineer:  EngineerSpec{ID: "colombo-1957", Name: "Gioacchino Colombo", Setup: 80, Strategy: 76, Ops: 74},
		Principal: PrincipalSpec{ID: "ugolini-1957", Name: "Nello Ugolini", Development: 74, Leadership: 80, Nerve: 78},
	},
	{
		ID: "vanwall-1958", Team: "Vanwall", Year: 1958, EraID: "1950s", Livery: "#0B3D2E",
		Car: CarSpec{ID: "vanwall-vw5", Name: "Vanwall VW5",
			Power: 86, Cornering: 84, Aero: 80, Reliability: 50},
		Drivers: [2]DriverSpec{
			{ID: "moss-1958", Name: "Stirling Moss", Pace: 95, Racecraft: 92, Consistency: 87, Composure: 86},
			{ID: "brooks-1958", Name: "Tony Brooks", Pace: 84, Racecraft: 82, Consistency: 80, Composure: 78},
		},
		// Chapman designed the Vanwall spaceframe before Lotus was a Formula
		// One team, which is why he can be drafted twice in one roster.
		Engineer:  EngineerSpec{ID: "chapman-1958", Name: "Colin Chapman", Setup: 86, Strategy: 78, Ops: 72},
		Principal: PrincipalSpec{ID: "vandervell-1958", Name: "Tony Vandervell", Development: 80, Leadership: 76, Nerve: 80},
	},
	{
		ID: "cooper-1959", Team: "Cooper", Year: 1959, EraID: "1950s", Livery: "#0B5D2E",
		Car: CarSpec{ID: "cooper-t51", Name: "Cooper T51",
			Power: 76, Cornering: 90, Aero: 80, Reliability: 64},
		Drivers: [2]DriverSpec{
			{ID: "brabham-1959", Name: "Jack Brabham", Pace: 86, Racecraft: 88, Consistency: 86, Composure: 84},
			{ID: "bmclaren-1959", Name: "Bruce McLaren", Pace: 82, Racecraft: 84, Consistency: 82, Composure: 82},
		},
		Engineer:  EngineerSpec{ID: "maddock-1959", Name: "Owen Maddock", Setup: 80, Strategy: 78, Ops: 80},
		Principal: PrincipalSpec{ID: "jcooper-1959", Name: "John Cooper", Development: 78, Leadership: 82, Nerve: 78},
	},
}
