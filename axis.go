package runby

// Axis is what every detection result carries, whatever axis produced it: how
// far the evidence goes, the product-specific context beside it, and the
// variable names it came from.
//
// It is embedded rather than repeated so that the shared defaults are applied
// in one place, and so that a driver on any axis fills the same three fields
// the same way. Embedding keeps the serialized shape flat, so a Layer, CI,
// Terminal, or Remote marshals to the same JSON it would without it.
//
// AncestorPID is deliberately not here. A CI run is a job on a runner rather
// than a process this one descends from, so there is nothing in the ancestor
// chain to corroborate it against and the field would always be zero. The
// three axes that can be corroborated declare it themselves.
type Axis struct {
	Confidence Confidence `json:"confidence"`

	// Extra holds values that only one product on this axis advertises, keyed
	// by "<product-slug>.<name>", so that product-specific metadata does not
	// widen the shared fields. Keys are stable; treat missing keys as unset.
	Extra map[string]string `json:"extra,omitempty"`

	// Evidence lists the environment variable names that produced this
	// result, sorted. Their values may be sensitive and are never copied.
	Evidence []string `json:"evidence"`
}

// applyDefaults fills in what a driver is allowed to leave unset. Every axis
// treats a missing confidence as definite: a driver that matched at all has
// found its product, and says otherwise only when it means to.
func (a *Axis) applyDefaults() {
	if a.Confidence == "" {
		a.Confidence = ConfidenceDefinite
	}
}
