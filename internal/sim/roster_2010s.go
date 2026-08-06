package sim

// Blown floors, then the hybrids. Cars from here are fast AND reliable,
// which makes them the least interesting thing to take and the hardest to
// pass up -- the sharpest decision the draft offers is a 2010s car against
// a 1980s driver.
var roster2010s = []TeamEra{
	{
		ID: "ferrari-2010", Team: "Ferrari", Year: 2010, EraID: "2010s", Livery: "#CC0000",
		Car: CarSpec{ID: "ferrari-f10", Name: "Ferrari F10",
			Power: 90, Cornering: 88, Aero: 87, Reliability: 88},
		Drivers: [2]DriverSpec{
			{ID: "alonso-2010", Name: "Fernando Alonso", Pace: 95, Racecraft: 95, Consistency: 93, Composure: 93},
			{ID: "massa-2010", Name: "Felipe Massa", Pace: 84, Racecraft: 81, Consistency: 80, Composure: 78},
		},
		Engineer:  EngineerSpec{ID: "acosta-2010", Name: "Aldo Costa", Setup: 88, Strategy: 84, Ops: 88},
		Principal: PrincipalSpec{ID: "domenicali-2010", Name: "Stefano Domenicali", Development: 84, Leadership: 86, Nerve: 80},
	},
	{
		ID: "redbull-2011", Team: "Red Bull", Year: 2011, EraID: "2010s", Livery: "#1B2E7A",
		Car: CarSpec{ID: "redbull-rb7", Name: "Red Bull RB7",
			Power: 88, Cornering: 97, Aero: 97, Reliability: 88},
		Drivers: [2]DriverSpec{
			{ID: "vettel-2011", Name: "Sebastian Vettel", Pace: 95, Racecraft: 89, Consistency: 94, Composure: 91},
			{ID: "webber-2011", Name: "Mark Webber", Pace: 85, Racecraft: 84, Consistency: 82, Composure: 80},
		},
		Engineer:  EngineerSpec{ID: "newey-2011", Name: "Adrian Newey", Setup: 97, Strategy: 90, Ops: 88},
		Principal: PrincipalSpec{ID: "horner-2011", Name: "Christian Horner", Development: 90, Leadership: 88, Nerve: 90},
	},
	{
		ID: "mclaren-2012", Team: "McLaren", Year: 2012, EraID: "2010s", Livery: "#C0C0C0",
		Car: CarSpec{ID: "mclaren-mp427", Name: "McLaren MP4-27",
			Power: 92, Cornering: 91, Aero: 90, Reliability: 74},
		Drivers: [2]DriverSpec{
			{ID: "hamilton-2012", Name: "Lewis Hamilton", Pace: 95, Racecraft: 94, Consistency: 88, Composure: 87},
			{ID: "button-2012", Name: "Jenson Button", Pace: 86, Racecraft: 84, Consistency: 82, Composure: 85},
		},
		Engineer:  EngineerSpec{ID: "plowe-2012", Name: "Paddy Lowe", Setup: 90, Strategy: 82, Ops: 76},
		Principal: PrincipalSpec{ID: "whitmarsh-2012", Name: "Martin Whitmarsh", Development: 82, Leadership: 84, Nerve: 76},
	},
	{
		ID: "mercedes-2016", Team: "Mercedes", Year: 2016, EraID: "2010s", Livery: "#00A19C",
		Car: CarSpec{ID: "mercedes-w07", Name: "Mercedes F1 W07",
			Power: 98, Cornering: 93, Aero: 93, Reliability: 87},
		Drivers: [2]DriverSpec{
			{ID: "hamilton-2016", Name: "Lewis Hamilton", Pace: 96, Racecraft: 95, Consistency: 90, Composure: 88},
			{ID: "rosberg-2016", Name: "Nico Rosberg", Pace: 89, Racecraft: 86, Consistency: 90, Composure: 87},
		},
		Engineer:  EngineerSpec{ID: "plowe-2016", Name: "Paddy Lowe", Setup: 92, Strategy: 88, Ops: 90},
		Principal: PrincipalSpec{ID: "wolff-2016", Name: "Toto Wolff", Development: 92, Leadership: 90, Nerve: 88},
	},
	{
		ID: "haas-2018", Team: "Haas", Year: 2018, EraID: "2010s", Livery: "#8C8C8C",
		Car: CarSpec{ID: "haas-vf18", Name: "Haas VF-18",
			Power: 88, Cornering: 82, Aero: 80, Reliability: 72},
		Drivers: [2]DriverSpec{
			{ID: "grosjean-2018", Name: "Romain Grosjean", Pace: 82, Racecraft: 76, Consistency: 66, Composure: 64},
			{ID: "magnussen-2018", Name: "Kevin Magnussen", Pace: 82, Racecraft: 80, Consistency: 78, Composure: 74},
		},
		// The pit-stop team that unlapped itself in Australia. Ops is the
		// lowest engineer rating in the roster and it is not a joke: this is
		// what a good car with a bad race operation is worth.
		Engineer:  EngineerSpec{ID: "rtaylor-2018", Name: "Rob Taylor", Setup: 80, Strategy: 70, Ops: 62},
		Principal: PrincipalSpec{ID: "steiner-2018", Name: "Guenther Steiner", Development: 74, Leadership: 84, Nerve: 78},
	},
}
