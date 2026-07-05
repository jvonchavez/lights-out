package sim

// Version is recorded on every season. Any change to the RNG draw order or
// to params.go is a breaking change and must bump this. Old seasons are
// verified under the version they were published with and are never
// recomputed. See docs/Architecture.md, "Data model".
const Version = "1.0.0"
