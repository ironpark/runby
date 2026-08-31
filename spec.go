package runby

import "strings"

// The CI, terminal, and remote axes detect a product the same way: a marker
// decides, then a fixed set of variables is read by name. Only the names
// differ, so those products are declared as data rather than as one function
// each, and the part of that data every axis shares lives here.
//
// The agent axis is deliberately not spec-driven. Its rules are irregular —
// confidence that depends on which variable matched, values derived rather
// than copied, evidence that counts only when a value matches — so most agents
// would need an escape hatch anyway. They are written as functions.

// specCore is the part of a spec that every axis shares: how to recognize the
// product, and the context and evidence that recognition carries.
type specCore struct {
	// marker reports whether the environment shows this product.
	marker Marker
	// markerNames lists the variables marker consults, so that they are
	// reported as evidence alongside the fields read below.
	markerNames []string
	// confidence defaults to ConfidenceDefinite when empty.
	confidence Confidence
	// trimBraces strips the surrounding curly braces that some products wrap
	// GUID values in, so the values are usable as-is.
	trimBraces bool
	// extra maps an Extra key to the variable supplying it.
	extra map[string]string
	// evidence lists further variables that count as evidence when set. The
	// marker and every field read through read are added automatically.
	evidence []string
}

// specField binds an environment variable name to the result field it fills.
// A field whose name is empty is skipped, so a spec simply leaves out what its
// product does not advertise.
type specField struct {
	name string
	into *string
}

// specValues is what reading a spec produces besides the fields it filled in
// place: the shared results that each axis copies onto its own struct.
type specValues struct {
	confidence Confidence
	extra      map[string]string
	// names accumulates every variable the spec consulted, set or not.
	// evidence turns it into the subset that is actually set.
	names []string
}

// add records further variables as consulted. Empty names are ignored so that
// an optional spec field can be passed without a guard at the call site.
func (v *specValues) add(names ...string) {
	for _, name := range names {
		if name != "" {
			v.names = append(v.names, name)
		}
	}
}

// evidence returns the sorted, deduplicated subset of the consulted variables
// that are set. Only names are returned; values may be sensitive.
func (v *specValues) evidence(env Env) []string { return PresentNames(env, v.names...) }

// read checks the marker and, when it matches, fills fields from env and
// collects the context and confidence shared by every axis. Callers record any
// axis-specific variables with add before asking for the evidence.
func (core specCore) read(env Env, fields ...specField) (specValues, bool) {
	if !core.marker(env) {
		return specValues{}, false
	}

	values := specValues{confidence: core.confidence}
	if values.confidence == "" {
		values.confidence = ConfidenceDefinite
	}
	values.add(core.markerNames...)
	values.add(core.evidence...)

	for _, field := range fields {
		if field.name == "" {
			continue
		}
		values.add(field.name)
		value, _ := Value(env, field.name)
		*field.into = core.trim(value)
	}

	for key, name := range core.extra {
		values.add(name)
		value, ok := Value(env, name)
		if !ok {
			continue
		}
		if values.extra == nil {
			values.extra = make(map[string]string, len(core.extra))
		}
		values.extra[key] = core.trim(value)
	}
	return values, true
}

func (core specCore) trim(value string) string {
	if !core.trimBraces {
		return value
	}
	return strings.TrimSuffix(strings.TrimPrefix(value, "{"), "}")
}
