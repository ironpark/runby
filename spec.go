package runby

import "strings"

// The CI, terminal, and remote axes detect a product the same way: a marker
// decides, then a fixed set of variables is read by name. Only the names
// differ, so those products are declared as data rather than as one function
// each, and the part of that data every axis shares lives here.
//
// The agent axis is deliberately not spec-driven, and the reason is about this
// file rather than about the agents. specField binds one variable name to one
// string field, which is all the three axes here ever need. Agents need more:
// a preferred source with a fallback (Orca's worktree then its repository,
// Codex's thread then its session), a field set to a constant rather than read
// (OpenCode's "acp", Antigravity's "sidecar", Amp's entrypoint, which depends
// on which variable matched), a bool field (Claude Code's Nested), a
// confidence that depends on which variable matched, an enum derived from a
// boolean, and an Extra value normalized rather than copied (both Codex).
// Only two of the eight agents fit the shape as it stands.
//
// Widening specField to cover them would land that complexity on the
// thirty-odd specs in this package that need none of it, and the irregular
// agents would still reach for an escape hatch. So agents are written as
// functions, and the one thing spec-driving would have guaranteed them — that
// a variable cannot be read without being reported as evidence — is provided
// instead by reader, in reader.go.

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

// apply copies the shared results onto a result's Axis and resolves the
// consulted variables into evidence. It is the last thing every spec-driven
// axis does, so that the three of them cannot drift apart in how they do it.
// Only names reach Evidence; values may be sensitive.
func (v *specValues) apply(env Env, axis *Axis) {
	axis.Confidence = v.confidence
	axis.Extra = v.extra
	axis.Evidence = PresentNames(env, v.names...)
}

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
