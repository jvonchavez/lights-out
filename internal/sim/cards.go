package sim

// Card is one development part offered during a window.
//
// Effect is the budget allocation the part represents, so a card is simply
// a named Decision: every existing carState method applies to it unchanged,
// and the risk it carries falls out of the same convex model the sliders
// used rather than being authored separately. That is deliberate -- the
// risk shown on a card face is the real risk, computed from the effect.
type Card struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Blurb  string   `json:"blurb"`
	Effect Decision `json:"effect"`
}

// Cost is the budget the part consumes. It drives cumulative risk through
// carState.failureChance, which is quadratic in total spend, so an
// expensive card is a bet rather than a free upgrade.
func (c Card) Cost() int { return c.Effect.Total() }

// CardPool is every part in the game. Deals draw from it without
// replacement within a window.
//
// Four shapes, so a deal is rarely a non-choice: a big swing into one area,
// a two-area split, a balanced update, and a cautious one. Aero-weighted
// cards are genuinely safer for their cost, because failureChance credits
// aero AeroEfficiency of the pressure its own share caused -- the player
// discovers that from the risk pips rather than from a rules paragraph.
var CardPool = []Card{
	// Big swings: one area, 240-260 units.
	{"gearbox-ti", "Titanium Gearbox", "Lighter, shorter ratios, and a habit of letting go.",
		Decision{Engine: 250}},
	{"hybrid-rework", "Hybrid Deployment Rework", "Bank more, spend it later, hope the software agrees.",
		Decision{Engine: 240}},
	{"monocoque-carbon", "Carbon Monocoque", "A season's stiffness bought in one go.",
		Decision{Chassis: 250}},
	{"pushrod-susp", "Pushrod Suspension", "Sharper turn-in, unforgiving over kerbs.",
		Decision{Chassis: 240}},
	{"skirt-assembly", "Skirt Assembly", "Seals the floor. Enormous when it works.",
		Decision{Aero: 260}},
	{"blown-diffuser", "Blown Diffuser", "Exhaust gas where the rules did not quite say no.",
		Decision{Aero: 250}},

	// Splits: two areas, 190-210 units.
	{"twin-keel", "Twin-Keel Floor", "Cleaner airflow to the rear, stiffer front end.",
		Decision{Chassis: 100, Aero: 100}},
	{"lowdrag-wing", "Low-Drag Rear Wing", "Trades a little cornering for a lot of straight.",
		Decision{Engine: 80, Aero: 120}},
	{"revised-floor", "Revised Floor", "The unglamorous update that usually works.",
		Decision{Chassis: 110, Aero: 90}},
	{"short-cassette", "Short-Ratio Cassette", "Better out of slow corners, busier on the straight.",
		Decision{Chassis: 90, Engine: 110}},
	{"coanda-exhaust", "Coanda Exhaust", "Bends the plume down onto the floor.",
		Decision{Engine: 100, Aero: 100}},
	{"stiff-bulkhead", "Stiffened Bulkhead", "Lets the suspension do what it was drawn to do.",
		Decision{Chassis: 120, Engine: 80}},
	{"undertray-vanes", "Undertray Vanes", "Small parts, disproportionate downforce.",
		Decision{Chassis: 70, Aero: 130}},
	{"compressor-up", "Compressor Upgrade", "More air, more heat, more of everything.",
		Decision{Engine: 130, Aero: 70}},

	// Balanced updates: 180-200 units.
	{"winter-rebuild", "Winter Rebuild", "Everything a little better, nothing transformed.",
		Decision{Chassis: 70, Engine: 70, Aero: 60}},
	{"windtunnel-prog", "Wind Tunnel Programme", "Hours in the tunnel, mostly on the body.",
		Decision{Chassis: 60, Engine: 60, Aero: 80}},
	{"sim-correlation", "Simulator Correlation", "The model finally matches the car.",
		Decision{Chassis: 65, Engine: 65, Aero: 65}},
	{"lightweight-pkg", "Lightweight Package", "Grams everywhere. They add up.",
		Decision{Chassis: 70, Engine: 60, Aero: 70}},
	{"aggressive-upd", "Aggressive Update", "Every department got what it asked for.",
		Decision{Chassis: 60, Engine: 70, Aero: 70}},

	// Cautious updates: 140-160 units.
	{"conservative-upd", "Conservative Update", "Nobody was promoted for this. Nobody retired either.",
		Decision{Chassis: 50, Engine: 50, Aero: 50}},
	{"reliability-pass", "Reliability Pass", "Fewer clever ideas, more finished races.",
		Decision{Chassis: 40, Engine: 50, Aero: 60}},
	{"minor-refine", "Minor Refinement", "A quiet fortnight in the drawing office.",
		Decision{Chassis: 50, Engine: 40, Aero: 50}},
	{"cooling-revision", "Cooling Revision", "Bigger inlets. The engine stops complaining.",
		Decision{Chassis: 45, Engine: 55, Aero: 50}},
	{"bodywork-tidy", "Bodywork Tidy", "Tidier surfaces, no headlines.",
		Decision{Chassis: 55, Engine: 45, Aero: 55}},
}
