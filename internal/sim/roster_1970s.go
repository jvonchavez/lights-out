package sim

// The 1970s. Wings, then ground effect, and the first cars whose lap time
// came out of the floor rather than the engine. The Lotus 79 is the
// highest Aero rating in the roster before the 1990s and it earned it.
var roster1970s = []TeamEra{
	{
		ID: "lotus-1972", Team: "Lotus", Year: 1972, EraID: "1970s", Livery: "#111111",
		Car: CarSpec{ID: "lotus-72d", Name: "Lotus 72D",
			Power: 86, Cornering: 90, Aero: 88, Reliability: 62},
		Drivers: [2]DriverSpec{
			{ID: "fittipaldi-1972", Name: "Emerson Fittipaldi", Pace: 88, Racecraft: 86, Consistency: 88, Composure: 86},
			{ID: "walker-1972", Name: "Dave Walker", Pace: 64, Racecraft: 62, Consistency: 60, Composure: 60},
		},
		Engineer:  EngineerSpec{ID: "philippe-1972", Name: "Maurice Philippe", Setup: 86, Strategy: 82, Ops: 78},
		Principal: PrincipalSpec{ID: "chapman-1972", Name: "Colin Chapman", Development: 93, Leadership: 84, Nerve: 82},
	},
	{
		ID: "tyrrell-1973", Team: "Tyrrell", Year: 1973, EraID: "1970s", Livery: "#1B4F9C",
		Car: CarSpec{ID: "tyrrell-006", Name: "Tyrrell 006",
			Power: 84, Cornering: 88, Aero: 82, Reliability: 72},
		Drivers: [2]DriverSpec{
			{ID: "stewart-1973", Name: "Jackie Stewart", Pace: 95, Racecraft: 95, Consistency: 95, Composure: 93},
			{ID: "cevert-1973", Name: "Francois Cevert", Pace: 85, Racecraft: 82, Consistency: 82, Composure: 80},
		},
		Engineer:  EngineerSpec{ID: "gardner-1973", Name: "Derek Gardner", Setup: 84, Strategy: 84, Ops: 82},
		Principal: PrincipalSpec{ID: "ktyrrell-1973", Name: "Ken Tyrrell", Development: 84, Leadership: 89, Nerve: 87},
	},
	{
		ID: "mclaren-1976", Team: "McLaren", Year: 1976, EraID: "1970s", Livery: "#E8590C",
		Car: CarSpec{ID: "mclaren-m23", Name: "McLaren M23",
			Power: 86, Cornering: 84, Aero: 82, Reliability: 66},
		Drivers: [2]DriverSpec{
			{ID: "hunt-1976", Name: "James Hunt", Pace: 90, Racecraft: 84, Consistency: 76, Composure: 74},
			{ID: "mass-1976", Name: "Jochen Mass", Pace: 76, Racecraft: 76, Consistency: 78, Composure: 78},
		},
		Engineer:  EngineerSpec{ID: "coppuck-1976", Name: "Gordon Coppuck", Setup: 84, Strategy: 82, Ops: 80},
		Principal: PrincipalSpec{ID: "mayer-1976", Name: "Teddy Mayer", Development: 80, Leadership: 80, Nerve: 84},
	},
	{
		ID: "lotus-1978", Team: "Lotus", Year: 1978, EraID: "1970s", Livery: "#111111",
		Car: CarSpec{ID: "lotus-79", Name: "Lotus 79",
			Power: 82, Cornering: 95, Aero: 96, Reliability: 58},
		Drivers: [2]DriverSpec{
			{ID: "andretti-1978", Name: "Mario Andretti", Pace: 90, Racecraft: 88, Consistency: 88, Composure: 86},
			{ID: "peterson-1978", Name: "Ronnie Peterson", Pace: 93, Racecraft: 88, Consistency: 80, Composure: 80},
		},
		Engineer:  EngineerSpec{ID: "wright-1978", Name: "Peter Wright", Setup: 90, Strategy: 82, Ops: 74},
		Principal: PrincipalSpec{ID: "chapman-1978", Name: "Colin Chapman", Development: 95, Leadership: 84, Nerve: 82},
	},
	{
		ID: "ferrari-1979", Team: "Ferrari", Year: 1979, EraID: "1970s", Livery: "#CC0000",
		Car: CarSpec{ID: "ferrari-312t4", Name: "Ferrari 312T4",
			Power: 92, Cornering: 86, Aero: 80, Reliability: 78},
		Drivers: [2]DriverSpec{
			{ID: "scheckter-1979", Name: "Jody Scheckter", Pace: 85, Racecraft: 86, Consistency: 88, Composure: 86},
			{ID: "gvilleneuve-1979", Name: "Gilles Villeneuve", Pace: 95, Racecraft: 94, Consistency: 72, Composure: 74},
		},
		Engineer:  EngineerSpec{ID: "forghieri-1979", Name: "Mauro Forghieri", Setup: 88, Strategy: 86, Ops: 88},
		Principal: PrincipalSpec{ID: "enzo-1979", Name: "Enzo Ferrari", Development: 86, Leadership: 84, Nerve: 90},
	},
}
