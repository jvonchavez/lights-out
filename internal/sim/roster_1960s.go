package sim

// The 1960s. The rear-engine revolution has settled and the cars are
// getting genuinely fast without getting any more likely to reach the
// flag -- the Lotus 49 is the sharpest car on this list and the most
// likely to break in your hands.
var roster1960s = []TeamEra{
	{
		ID: "ferrari-1961", Team: "Ferrari", Year: 1961, EraID: "1960s", Livery: "#CC0000",
		Car: CarSpec{ID: "ferrari-156", Name: "Ferrari 156 Sharknose",
			Power: 90, Cornering: 82, Aero: 78, Reliability: 58},
		Drivers: [2]DriverSpec{
			{ID: "phill-1961", Name: "Phil Hill", Pace: 84, Racecraft: 82, Consistency: 84, Composure: 82},
			{ID: "vontrips-1961", Name: "Wolfgang von Trips", Pace: 83, Racecraft: 80, Consistency: 78, Composure: 76},
		},
		Engineer:  EngineerSpec{ID: "chiti-1961", Name: "Carlo Chiti", Setup: 82, Strategy: 80, Ops: 76},
		Principal: PrincipalSpec{ID: "enzo-1961", Name: "Enzo Ferrari", Development: 88, Leadership: 86, Nerve: 90},
	},
	{
		ID: "brabham-1966", Team: "Brabham", Year: 1966, EraID: "1960s", Livery: "#1B6B3A",
		Car: CarSpec{ID: "brabham-bt19", Name: "Brabham BT19",
			Power: 84, Cornering: 86, Aero: 78, Reliability: 76},
		Drivers: [2]DriverSpec{
			{ID: "brabham-1966", Name: "Jack Brabham", Pace: 84, Racecraft: 88, Consistency: 88, Composure: 86},
			{ID: "hulme-1966", Name: "Denny Hulme", Pace: 84, Racecraft: 84, Consistency: 86, Composure: 84},
		},
		Engineer:  EngineerSpec{ID: "tauranac-1966", Name: "Ron Tauranac", Setup: 84, Strategy: 82, Ops: 86},
		Principal: PrincipalSpec{ID: "jbrabham-1966", Name: "Jack Brabham", Development: 82, Leadership: 84, Nerve: 84},
	},
	{
		ID: "lotus-1965", Team: "Lotus", Year: 1965, EraID: "1960s", Livery: "#0B5D2E",
		Car: CarSpec{ID: "lotus-33", Name: "Lotus 33",
			Power: 82, Cornering: 94, Aero: 86, Reliability: 56},
		Drivers: [2]DriverSpec{
			{ID: "clark-1965", Name: "Jim Clark", Pace: 98, Racecraft: 90, Consistency: 92, Composure: 88},
			{ID: "spence-1965", Name: "Mike Spence", Pace: 76, Racecraft: 74, Consistency: 74, Composure: 74},
		},
		Engineer:  EngineerSpec{ID: "terry-1965", Name: "Len Terry", Setup: 86, Strategy: 80, Ops: 76},
		Principal: PrincipalSpec{ID: "chapman-1965", Name: "Colin Chapman", Development: 92, Leadership: 84, Nerve: 82},
	},
	{
		ID: "lotus-1967", Team: "Lotus", Year: 1967, EraID: "1960s", Livery: "#0B5D2E",
		Car: CarSpec{ID: "lotus-49", Name: "Lotus 49",
			Power: 94, Cornering: 90, Aero: 84, Reliability: 48},
		Drivers: [2]DriverSpec{
			{ID: "clark-1967", Name: "Jim Clark", Pace: 98, Racecraft: 91, Consistency: 92, Composure: 89},
			{ID: "ghill-1967", Name: "Graham Hill", Pace: 86, Racecraft: 86, Consistency: 84, Composure: 86},
		},
		Engineer:  EngineerSpec{ID: "philippe-1967", Name: "Maurice Philippe", Setup: 84, Strategy: 80, Ops: 74},
		Principal: PrincipalSpec{ID: "chapman-1967", Name: "Colin Chapman", Development: 94, Leadership: 84, Nerve: 80},
	},
	{
		ID: "matra-1969", Team: "Matra", Year: 1969, EraID: "1960s", Livery: "#1B4F9C",
		Car: CarSpec{ID: "matra-ms80", Name: "Matra MS80",
			Power: 86, Cornering: 88, Aero: 82, Reliability: 70},
		Drivers: [2]DriverSpec{
			{ID: "stewart-1969", Name: "Jackie Stewart", Pace: 94, Racecraft: 94, Consistency: 94, Composure: 92},
			{ID: "beltoise-1969", Name: "Jean-Pierre Beltoise", Pace: 78, Racecraft: 78, Consistency: 76, Composure: 76},
		},
		Engineer:  EngineerSpec{ID: "boyer-1969", Name: "Bernard Boyer", Setup: 82, Strategy: 80, Ops: 80},
		Principal: PrincipalSpec{ID: "ktyrrell-1969", Name: "Ken Tyrrell", Development: 84, Leadership: 88, Nerve: 86},
	},
}
