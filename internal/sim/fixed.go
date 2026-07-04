package sim

// Milli is a fixed-point number scaled by 1000. All sim arithmetic is
// integer so that native and js/wasm builds are bit-identical by
// construction rather than by hope. See docs/Game Design.md, "Determinism
// contract".
//
// Ratings stay under 200_000 and sigmas under 10_000, so the intermediate
// product in Mul and Div stays well under 2e9. There is no overflow risk at
// these magnitudes and no check for one.
type Milli int64

// One is the fixed-point representation of 1.0. As a probability it means
// certainty.
const One Milli = 1000

// Mul multiplies two Milli values and rescales. Truncation is toward zero,
// which is Go's integer division behaviour and is identical on every target.
func (a Milli) Mul(b Milli) Milli { return a * b / One }

// Div divides two Milli values and rescales.
func (a Milli) Div(b Milli) Milli { return a * One / b }

// FromInt converts a whole number to Milli.
func FromInt(n int) Milli { return Milli(n) * One }
