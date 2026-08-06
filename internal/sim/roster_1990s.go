package sim

// The 1990s. Active suspension arrives, wins everything, and is banned;
// reliability finally becomes something a team can rely on rather than
// hope for. This is the era where taking a car stops being a gamble and
// starts being a choice.
var roster1990s = []TeamEra{
	{
		ID: "jordan-1991", Team: "Jordan", Year: 1991, EraID: "1990s", Livery: "#2E7D32",
		Car: CarSpec{ID: "jordan-191", Name: "Jordan 191",
			Power: 74, Cornering: 84, Aero: 82, Reliability: 62},
		Drivers: [2]DriverSpec{
			{ID: "gachot-1991", Name: "Bertrand Gachot", Pace: 74, Racecraft: 74, Consistency: 74, Composure: 70},
			{ID: "decesaris-1991", Name: "Andrea de Cesaris", Pace: 78, Racecraft: 74, Consistency: 64, Composure: 62},
		},
		Engineer:  EngineerSpec{ID: "ganderson-1991", Name: "Gary Anderson", Setup: 86, Strategy: 80, Ops: 74},
		Principal: PrincipalSpec{ID: "ejordan-1991", Name: "Eddie Jordan", Development: 76, Leadership: 88, Nerve: 80},
	},
	{
		ID: "mclaren-1991", Team: "McLaren", Year: 1991, EraID: "1990s", Livery: "#D01020",
		Car: CarSpec{ID: "mclaren-mp46", Name: "McLaren MP4/6",
			Power: 92, Cornering: 88, Aero: 84, Reliability: 78},
		Drivers: [2]DriverSpec{
			{ID: "senna-1991", Name: "Ayrton Senna", Pace: 99, Racecraft: 96, Consistency: 90, Composure: 91},
			{ID: "berger-1991", Name: "Gerhard Berger", Pace: 84, Racecraft: 82, Consistency: 80, Composure: 82},
		},
		Engineer:  EngineerSpec{ID: "oatley-1991", Name: "Neil Oatley", Setup: 86, Strategy: 84, Ops: 84},
		Principal: PrincipalSpec{ID: "dennis-1991", Name: "Ron Dennis", Development: 88, Leadership: 89, Nerve: 91},
	},
	{
		ID: "williams-1992", Team: "Williams", Year: 1992, EraID: "1990s", Livery: "#1B4F9C",
		Car: CarSpec{ID: "williams-fw14b", Name: "Williams FW14B",
			Power: 94, Cornering: 97, Aero: 96, Reliability: 76},
		Drivers: [2]DriverSpec{
			{ID: "mansell-1992", Name: "Nigel Mansell", Pace: 94, Racecraft: 91, Consistency: 82, Composure: 80},
			{ID: "patrese-1992", Name: "Riccardo Patrese", Pace: 84, Racecraft: 82, Consistency: 80, Composure: 80},
		},
		Engineer:  EngineerSpec{ID: "newey-1992", Name: "Adrian Newey", Setup: 96, Strategy: 88, Ops: 82},
		Principal: PrincipalSpec{ID: "fwilliams-1992", Name: "Frank Williams", Development: 92, Leadership: 87, Nerve: 88},
	},
	{
		ID: "benetton-1994", Team: "Benetton", Year: 1994, EraID: "1990s", Livery: "#1F9C5A",
		Car: CarSpec{ID: "benetton-b194", Name: "Benetton B194",
			Power: 84, Cornering: 92, Aero: 90, Reliability: 80},
		Drivers: [2]DriverSpec{
			{ID: "schumacher-1994", Name: "Michael Schumacher", Pace: 96, Racecraft: 95, Consistency: 93, Composure: 90},
			{ID: "jverstappen-1994", Name: "Jos Verstappen", Pace: 74, Racecraft: 72, Consistency: 68, Composure: 68},
		},
		Engineer:  EngineerSpec{ID: "byrne-1994", Name: "Rory Byrne", Setup: 90, Strategy: 90, Ops: 86},
		Principal: PrincipalSpec{ID: "briatore-1994", Name: "Flavio Briatore", Development: 86, Leadership: 84, Nerve: 92},
	},
	{
		ID: "mclaren-1998", Team: "McLaren", Year: 1998, EraID: "1990s", Livery: "#C0C0C0",
		Car: CarSpec{ID: "mclaren-mp413", Name: "McLaren MP4/13",
			Power: 92, Cornering: 95, Aero: 94, Reliability: 82},
		Drivers: [2]DriverSpec{
			{ID: "hakkinen-1998", Name: "Mika Hakkinen", Pace: 93, Racecraft: 86, Consistency: 88, Composure: 88},
			{ID: "coulthard-1998", Name: "David Coulthard", Pace: 85, Racecraft: 82, Consistency: 84, Composure: 82},
		},
		Engineer:  EngineerSpec{ID: "newey-1998", Name: "Adrian Newey", Setup: 96, Strategy: 89, Ops: 84},
		Principal: PrincipalSpec{ID: "dennis-1998", Name: "Ron Dennis", Development: 89, Leadership: 89, Nerve: 90},
	},
	{
		ID: "williams-1996", Team: "Williams", Year: 1996, EraID: "1990s", Livery: "#1B4F9C",
		Car: CarSpec{ID: "williams-fw18", Name: "Williams FW18",
			Power: 93, Cornering: 95, Aero: 94, Reliability: 84},
		Drivers: [2]DriverSpec{
			{ID: "dhill-1996", Name: "Damon Hill", Pace: 86, Racecraft: 82, Consistency: 86, Composure: 82},
			{ID: "jvilleneuve-1996", Name: "Jacques Villeneuve", Pace: 88, Racecraft: 84, Consistency: 82, Composure: 84},
		},
		Engineer:  EngineerSpec{ID: "newey-1996", Name: "Adrian Newey", Setup: 96, Strategy: 88, Ops: 86},
		Principal: PrincipalSpec{ID: "fwilliams-1996", Name: "Frank Williams", Development: 90, Leadership: 86, Nerve: 86},
	},
}
