package sim

// The circuit pool is real Formula 1 tracks: the whole 2026 calendar, plus
// a set of circuits the championship has raced on before. The historical
// entries are there for the same reason the roster spans eras -- a game
// about drafting the 1967 Lotus should be able to send it to Kyalami -- and
// because the pool needs depth in every archetype for the calendar to vary
// from run to run.
//
// Each circuit is classified by the CHARACTER that decides lap time, which
// is what the profile weights model:
//
//   power      long straights, low downforce, engine dominant
//   technical  slow corners, high downforce, very hard to overtake
//   balanced   no single dominant demand
//   highspeed  fast sweeping corners, aero-load dominant
//
// These are judgment calls on real circuits, made on downforce level and
// overtaking difficulty rather than on lap length or top speed alone.
// Zandvoort is technical despite its banking because it is a high-downforce
// circuit where nobody can pass; Spa is power despite its famous corners
// because teams take low-downforce wings there.
//
// Every list below is a slice, not a map, because they are ranged over in a
// path that affects results.

var circuitsPower = []Circuit{
	{Name: "Monza", Archetype: "power"},
	{Name: "Spa-Francorchamps", Archetype: "power"},
	{Name: "Baku City", Archetype: "power"},
	{Name: "Las Vegas Strip", Archetype: "power"},
	{Name: "Mexico City", Archetype: "power"},
	{Name: "Red Bull Ring", Archetype: "power"},
	{Name: "Gilles Villeneuve", Archetype: "power"},
	{Name: "Hockenheimring", Archetype: "power"},
	{Name: "Paul Ricard", Archetype: "power"},
}

var circuitsTechnical = []Circuit{
	{Name: "Monaco", Archetype: "technical"},
	{Name: "Hungaroring", Archetype: "technical"},
	{Name: "Zandvoort", Archetype: "technical"},
	{Name: "Marina Bay", Archetype: "technical"},
	{Name: "Imola", Archetype: "technical"},
	{Name: "Magny-Cours", Archetype: "technical"},
	{Name: "Adelaide", Archetype: "technical"},
	{Name: "Valencia Street", Archetype: "technical"},
}

var circuitsBalanced = []Circuit{
	{Name: "Shanghai", Archetype: "balanced"},
	{Name: "Miami", Archetype: "balanced"},
	{Name: "Barcelona-Catalunya", Archetype: "balanced"},
	{Name: "Madring", Archetype: "balanced"},
	{Name: "Bahrain", Archetype: "balanced"},
	{Name: "Circuit of the Americas", Archetype: "balanced"},
	{Name: "Interlagos", Archetype: "balanced"},
	{Name: "Yas Marina", Archetype: "balanced"},
	{Name: "Sepang", Archetype: "balanced"},
	{Name: "Nurburgring", Archetype: "balanced"},
}

var circuitsHighSpeed = []Circuit{
	{Name: "Silverstone", Archetype: "highspeed"},
	{Name: "Suzuka", Archetype: "highspeed"},
	{Name: "Albert Park", Archetype: "highspeed"},
	{Name: "Lusail", Archetype: "highspeed"},
	{Name: "Istanbul Park", Archetype: "highspeed"},
	{Name: "Kyalami", Archetype: "highspeed"},
	{Name: "Buddh International", Archetype: "highspeed"},
}

// CircuitQuota fixes the SHAPE of every calendar: three power circuits,
// three technical, two balanced, two high-speed, in some order.
//
// The membership varies from run to run and the shape never does, which is
// the point. A run that drew ten power circuits would make one rating the
// whole game; a fixed ten-circuit calendar made every run's calendar
// identical. Drawing to a quota from a deep pool gives variety without
// giving up the balance the profile weights were tuned against.
var CircuitQuota = []struct {
	Archetype string
	Count     int
	pool      []Circuit
}{
	{"power", 3, circuitsPower},
	{"technical", 3, circuitsTechnical},
	{"balanced", 2, circuitsBalanced},
	{"highspeed", 2, circuitsHighSpeed},
}

// circuitPool is every circuit, for validation and for the UI.
var circuitPool = func() []Circuit {
	all := make([]Circuit, 0, 34)
	all = append(all, circuitsPower...)
	all = append(all, circuitsTechnical...)
	all = append(all, circuitsBalanced...)
	all = append(all, circuitsHighSpeed...)
	return all
}()
