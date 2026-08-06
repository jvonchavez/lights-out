package sim

// The V10 years and the run-out to 2009. Reliability is now close to
// solved, so a car from here is the safe half of the roster -- and the
// F2004 is the only entry that is both the fastest thing in its season and
// close to unbreakable, which is exactly why it should cost you every
// other slot to take it.
var roster2000s = []TeamEra{
	{
		ID: "mclaren-2000", Team: "McLaren", Year: 2000, EraID: "2000s", Livery: "#C0C0C0",
		Car: CarSpec{ID: "mclaren-mp415", Name: "McLaren MP4-15",
			Power: 92, Cornering: 93, Aero: 92, Reliability: 80},
		Drivers: [2]DriverSpec{
			{ID: "hakkinen-2000", Name: "Mika Hakkinen", Pace: 94, Racecraft: 87, Consistency: 88, Composure: 89},
			{ID: "coulthard-2000", Name: "David Coulthard", Pace: 86, Racecraft: 83, Consistency: 85, Composure: 83},
		},
		Engineer:  EngineerSpec{ID: "newey-2000", Name: "Adrian Newey", Setup: 95, Strategy: 88, Ops: 84},
		Principal: PrincipalSpec{ID: "dennis-2000", Name: "Ron Dennis", Development: 89, Leadership: 88, Nerve: 88},
	},
	{
		ID: "ferrari-2004", Team: "Ferrari", Year: 2004, EraID: "2000s", Livery: "#CC0000",
		Car: CarSpec{ID: "ferrari-f2004", Name: "Ferrari F2004",
			Power: 97, Cornering: 95, Aero: 94, Reliability: 93},
		Drivers: [2]DriverSpec{
			{ID: "schumacher-2004", Name: "Michael Schumacher", Pace: 97, Racecraft: 96, Consistency: 96, Composure: 94},
			{ID: "barrichello-2004", Name: "Rubens Barrichello", Pace: 86, Racecraft: 84, Consistency: 86, Composure: 84},
		},
		Engineer:  EngineerSpec{ID: "brawn-2004", Name: "Ross Brawn", Setup: 92, Strategy: 97, Ops: 93},
		Principal: PrincipalSpec{ID: "todt-2004", Name: "Jean Todt", Development: 93, Leadership: 92, Nerve: 94},
	},
	{
		ID: "renault-2005", Team: "Renault", Year: 2005, EraID: "2000s", Livery: "#FFD400",
		Car: CarSpec{ID: "renault-r25", Name: "Renault R25",
			Power: 88, Cornering: 93, Aero: 91, Reliability: 90},
		Drivers: [2]DriverSpec{
			{ID: "alonso-2005", Name: "Fernando Alonso", Pace: 95, Racecraft: 93, Consistency: 93, Composure: 92},
			{ID: "fisichella-2005", Name: "Giancarlo Fisichella", Pace: 82, Racecraft: 78, Consistency: 78, Composure: 76},
		},
		Engineer:  EngineerSpec{ID: "bbell-2005", Name: "Bob Bell", Setup: 88, Strategy: 90, Ops: 89},
		Principal: PrincipalSpec{ID: "briatore-2005", Name: "Flavio Briatore", Development: 87, Leadership: 86, Nerve: 91},
	},
	{
		ID: "honda-2006", Team: "Honda", Year: 2006, EraID: "2000s", Livery: "#E8E8E8",
		Car: CarSpec{ID: "honda-ra106", Name: "Honda RA106",
			Power: 88, Cornering: 78, Aero: 76, Reliability: 74},
		Drivers: [2]DriverSpec{
			{ID: "button-2006", Name: "Jenson Button", Pace: 85, Racecraft: 82, Consistency: 84, Composure: 84},
			{ID: "barrichello-2006", Name: "Rubens Barrichello", Pace: 84, Racecraft: 83, Consistency: 85, Composure: 84},
		},
		Engineer:  EngineerSpec{ID: "gwillis-2006", Name: "Geoff Willis", Setup: 82, Strategy: 80, Ops: 78},
		Principal: PrincipalSpec{ID: "nfry-2006", Name: "Nick Fry", Development: 76, Leadership: 76, Nerve: 74},
	},
	{
		ID: "brawn-2009", Team: "Brawn GP", Year: 2009, EraID: "2000s", Livery: "#D6E600",
		Car: CarSpec{ID: "brawn-bgp001", Name: "Brawn BGP 001",
			Power: 90, Cornering: 92, Aero: 93, Reliability: 88},
		Drivers: [2]DriverSpec{
			{ID: "button-2009", Name: "Jenson Button", Pace: 87, Racecraft: 84, Consistency: 86, Composure: 86},
			{ID: "barrichello-2009", Name: "Rubens Barrichello", Pace: 85, Racecraft: 84, Consistency: 84, Composure: 84},
		},
		Engineer: EngineerSpec{ID: "bigois-2009", Name: "Loic Bigois", Setup: 90, Strategy: 86, Ops: 84},
		// Brawn as a principal rather than an engineer here: in 2009 he was
		// running the team that carried his name, not drawing the car.
		Principal: PrincipalSpec{ID: "brawn-2009", Name: "Ross Brawn", Development: 91, Leadership: 90, Nerve: 93},
	},
}
