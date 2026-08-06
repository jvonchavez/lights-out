package sim

// Ground effect, part two -- the immediate prehistory of the grid you are
// racing. These are the strongest all-round entries in the roster and the
// most expensive thing to pass up, because a 2020s car costs you nothing in
// reliability at all.
var roster2020s = []TeamEra{
	{
		ID: "mercedes-2020", Team: "Mercedes", Year: 2020, EraID: "2020s", Livery: "#00A19C",
		Car: CarSpec{ID: "mercedes-w11", Name: "Mercedes F1 W11",
			Power: 98, Cornering: 96, Aero: 96, Reliability: 92},
		Drivers: [2]DriverSpec{
			{ID: "hamilton-2020", Name: "Lewis Hamilton", Pace: 97, Racecraft: 96, Consistency: 94, Composure: 92},
			{ID: "bottas-2020", Name: "Valtteri Bottas", Pace: 85, Racecraft: 82, Consistency: 86, Composure: 85},
		},
		Engineer:  EngineerSpec{ID: "allison-2020", Name: "James Allison", Setup: 94, Strategy: 91, Ops: 92},
		Principal: PrincipalSpec{ID: "wolff-2020", Name: "Toto Wolff", Development: 94, Leadership: 92, Nerve: 90},
	},
	{
		ID: "williams-2021", Team: "Williams", Year: 2021, EraID: "2020s", Livery: "#00A3E0",
		Car: CarSpec{ID: "williams-fw43b", Name: "Williams FW43B",
			Power: 84, Cornering: 62, Aero: 60, Reliability: 76},
		Drivers: [2]DriverSpec{
			{ID: "russell-2021", Name: "George Russell", Pace: 88, Racecraft: 85, Consistency: 86, Composure: 84},
			{ID: "latifi-2021", Name: "Nicholas Latifi", Pace: 66, Racecraft: 64, Consistency: 66, Composure: 62},
		},
		Engineer:  EngineerSpec{ID: "demaison-2021", Name: "Francois-Xavier Demaison", Setup: 76, Strategy: 78, Ops: 76},
		Principal: PrincipalSpec{ID: "capito-2021", Name: "Jost Capito", Development: 74, Leadership: 80, Nerve: 78},
	},
	{
		ID: "ferrari-2022", Team: "Ferrari", Year: 2022, EraID: "2020s", Livery: "#CC0000",
		Car: CarSpec{ID: "ferrari-f175", Name: "Ferrari F1-75",
			Power: 94, Cornering: 93, Aero: 91, Reliability: 70},
		Drivers: [2]DriverSpec{
			{ID: "leclerc-2022", Name: "Charles Leclerc", Pace: 95, Racecraft: 87, Consistency: 84, Composure: 82},
			{ID: "sainz-2022", Name: "Carlos Sainz", Pace: 87, Racecraft: 85, Consistency: 84, Composure: 82},
		},
		Engineer: EngineerSpec{ID: "cardile-2022", Name: "Enrico Cardile", Setup: 90, Strategy: 86, Ops: 72},
		// The fastest car of the first half of 2022 and a strategy department
		// that gave the title away. Strategy is the rating this entry exists
		// to demonstrate.
		Principal: PrincipalSpec{ID: "binotto-2022", Name: "Mattia Binotto", Development: 84, Leadership: 78, Nerve: 68},
	},
	{
		ID: "redbull-2023", Team: "Red Bull", Year: 2023, EraID: "2020s", Livery: "#1B2E7A",
		Car: CarSpec{ID: "redbull-rb19", Name: "Red Bull RB19",
			Power: 96, Cornering: 98, Aero: 98, Reliability: 94},
		Drivers: [2]DriverSpec{
			{ID: "verstappen-2023", Name: "Max Verstappen", Pace: 99, Racecraft: 97, Consistency: 97, Composure: 94},
			{ID: "perez-2023", Name: "Sergio Perez", Pace: 83, Racecraft: 84, Consistency: 76, Composure: 76},
		},
		Engineer:  EngineerSpec{ID: "wache-2023", Name: "Pierre Wache", Setup: 95, Strategy: 93, Ops: 93},
		Principal: PrincipalSpec{ID: "horner-2023", Name: "Christian Horner", Development: 92, Leadership: 88, Nerve: 92},
	},
	{
		ID: "mclaren-2024", Team: "McLaren", Year: 2024, EraID: "2020s", Livery: "#FF8000",
		Car: CarSpec{ID: "mclaren-mcl38", Name: "McLaren MCL38",
			Power: 93, Cornering: 95, Aero: 95, Reliability: 90},
		Drivers: [2]DriverSpec{
			{ID: "norris-2024", Name: "Lando Norris", Pace: 94, Racecraft: 86, Consistency: 86, Composure: 82},
			{ID: "piastri-2024", Name: "Oscar Piastri", Pace: 90, Racecraft: 87, Consistency: 88, Composure: 86},
		},
		Engineer:  EngineerSpec{ID: "rmarshall-2024", Name: "Rob Marshall", Setup: 94, Strategy: 88, Ops: 90},
		Principal: PrincipalSpec{ID: "stella-2024", Name: "Andrea Stella", Development: 93, Leadership: 92, Nerve: 86},
	},
}
