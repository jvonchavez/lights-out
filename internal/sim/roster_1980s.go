package sim

// The turbo era. Qualifying engines running 1400bhp for four laps, and
// cars that either won by a minute or stopped on lap three. The MP4/4
// carries the highest car Overall in the roster because it won fifteen of
// sixteen races, and its reliability is still only fair, which is the
// bargain the whole roster is built on.
var roster1980s = []TeamEra{
	{
		ID: "ferrari-1982", Team: "Ferrari", Year: 1982, EraID: "1980s", Livery: "#CC0000",
		Car: CarSpec{ID: "ferrari-126c2", Name: "Ferrari 126C2",
			Power: 94, Cornering: 82, Aero: 84, Reliability: 52},
		Drivers: [2]DriverSpec{
			{ID: "gvilleneuve-1982", Name: "Gilles Villeneuve", Pace: 96, Racecraft: 95, Consistency: 70, Composure: 72},
			{ID: "pironi-1982", Name: "Didier Pironi", Pace: 85, Racecraft: 83, Consistency: 80, Composure: 76},
		},
		Engineer:  EngineerSpec{ID: "postlethwaite-1982", Name: "Harvey Postlethwaite", Setup: 86, Strategy: 82, Ops: 76},
		Principal: PrincipalSpec{ID: "enzo-1982", Name: "Enzo Ferrari", Development: 85, Leadership: 80, Nerve: 88},
	},
	{
		ID: "brabham-1983", Team: "Brabham", Year: 1983, EraID: "1980s", Livery: "#1B4F9C",
		Car: CarSpec{ID: "brabham-bt52", Name: "Brabham BT52",
			Power: 95, Cornering: 84, Aero: 86, Reliability: 54},
		Drivers: [2]DriverSpec{
			{ID: "piquet-1983", Name: "Nelson Piquet", Pace: 90, Racecraft: 88, Consistency: 86, Composure: 88},
			{ID: "patrese-1983", Name: "Riccardo Patrese", Pace: 84, Racecraft: 82, Consistency: 76, Composure: 76},
		},
		Engineer:  EngineerSpec{ID: "murray-1983", Name: "Gordon Murray", Setup: 92, Strategy: 90, Ops: 78},
		Principal: PrincipalSpec{ID: "ecclestone-1983", Name: "Bernie Ecclestone", Development: 82, Leadership: 82, Nerve: 90},
	},
	{
		ID: "mclaren-1984", Team: "McLaren", Year: 1984, EraID: "1980s", Livery: "#D01020",
		Car: CarSpec{ID: "mclaren-mp42", Name: "McLaren MP4/2",
			Power: 92, Cornering: 90, Aero: 88, Reliability: 74},
		Drivers: [2]DriverSpec{
			{ID: "lauda-1984", Name: "Niki Lauda", Pace: 88, Racecraft: 92, Consistency: 92, Composure: 94},
			{ID: "prost-1984", Name: "Alain Prost", Pace: 94, Racecraft: 93, Consistency: 94, Composure: 92},
		},
		Engineer:  EngineerSpec{ID: "barnard-1984", Name: "John Barnard", Setup: 92, Strategy: 86, Ops: 86},
		Principal: PrincipalSpec{ID: "dennis-1984", Name: "Ron Dennis", Development: 90, Leadership: 88, Nerve: 90},
	},
	{
		ID: "williams-1986", Team: "Williams", Year: 1986, EraID: "1980s", Livery: "#1B4F9C",
		Car: CarSpec{ID: "williams-fw11", Name: "Williams FW11",
			Power: 96, Cornering: 88, Aero: 86, Reliability: 68},
		Drivers: [2]DriverSpec{
			{ID: "mansell-1986", Name: "Nigel Mansell", Pace: 93, Racecraft: 90, Consistency: 78, Composure: 76},
			{ID: "piquet-1986", Name: "Nelson Piquet", Pace: 90, Racecraft: 88, Consistency: 87, Composure: 88},
		},
		Engineer:  EngineerSpec{ID: "head-1986", Name: "Patrick Head", Setup: 90, Strategy: 86, Ops: 84},
		Principal: PrincipalSpec{ID: "fwilliams-1986", Name: "Frank Williams", Development: 88, Leadership: 86, Nerve: 88},
	},
	{
		ID: "mclaren-1988", Team: "McLaren", Year: 1988, EraID: "1980s", Livery: "#D01020",
		Car: CarSpec{ID: "mclaren-mp44", Name: "McLaren MP4/4",
			Power: 97, Cornering: 96, Aero: 95, Reliability: 76},
		Drivers: [2]DriverSpec{
			{ID: "senna-1988", Name: "Ayrton Senna", Pace: 99, Racecraft: 95, Consistency: 88, Composure: 90},
			{ID: "prost-1988", Name: "Alain Prost", Pace: 95, Racecraft: 94, Consistency: 95, Composure: 93},
		},
		Engineer:  EngineerSpec{ID: "murray-1988", Name: "Gordon Murray", Setup: 93, Strategy: 91, Ops: 84},
		Principal: PrincipalSpec{ID: "dennis-1988", Name: "Ron Dennis", Development: 92, Leadership: 90, Nerve: 92},
	},
}
