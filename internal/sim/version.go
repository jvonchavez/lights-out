package sim

// Version is recorded on every season. Any change to the RNG draw order or
// to params.go is a breaking change and must bump this. Old seasons are
// verified under the version they were published with and are never
// recomputed. See docs/Architecture.md, "Data model".
//
// 3.0.0 replaced in-season development with a draft: five rolls, one item
// taken from each, and a 24-car field racing the real 2026 grid. Every
// number in the simulation changed.
const Version = "3.0.0"
