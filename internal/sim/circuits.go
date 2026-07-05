package sim

// circuitPool is the fixed 10-circuit roster: 3 power, 3 technical,
// 2 balanced, 2 high-speed. The daily seed shuffles the order, never the
// membership, so every player faces the same ten tracks in a different
// sequence. Names are fictional, which avoids trademark questions entirely
// and costs nothing in fun (docs/Game Design.md).
//
// This is a slice, not a map, because it is ranged over in a path that
// affects results.
var circuitPool = []Circuit{
	{Name: "Vellmar Straight", Archetype: "power"},
	{Name: "Costa Nera", Archetype: "power"},
	{Name: "Lakeshore Mile", Archetype: "power"},
	{Name: "Kestrel Park", Archetype: "technical"},
	{Name: "Innsbach", Archetype: "technical"},
	{Name: "Old Harbour", Archetype: "technical"},
	{Name: "Prairie Ring", Archetype: "balanced"},
	{Name: "Sable Bay", Archetype: "balanced"},
	{Name: "Mont Aubade", Archetype: "highspeed"},
	{Name: "Cape Verrick", Archetype: "highspeed"},
}

// rivalNames are the ten AI teams, in fixed order.
var rivalNames = []string{
	"Ardent Racing", "Bellweather GP", "Cinder Motorsport", "Dalgetty Racing",
	"Eastgate F1", "Fennmoor GP", "Grosvenor Racing", "Halcyon Motorsport",
	"Ironwood GP", "Jorvik Racing",
}

// rivalArchetypes cycles across the ten rivals, giving a 3/3/2/2 mix.
var rivalArchetypes = []string{"aggressive", "conservative", "specialist", "reactive"}
