package runby

// EnvReader is an Env paired with the record of what has been asked of it. It
// is how a driver reads an environment without having to maintain its evidence
// list by hand.
//
// Every driver must report the variables it consulted as Axis.Evidence, which
// is what lets a caller see why a detection fired without exposing any value.
// Keeping that list beside the lookups is the one part of writing a driver that
// can silently rot: add a lookup, forget to widen the list, and the detection
// under-reports its own reasoning while every test still passes.
//
// An EnvReader closes that gap. Asking it for a value records the name, so a
// driver's evidence is what it actually read rather than a second list
// maintained alongside:
//
//	Detect: func(env runby.Env) (runby.Agent, bool) {
//		r := runby.NewEnvReader(env)
//		id, ok := r.Value("ACME_RUN_ID")
//		if !ok {
//			return runby.Agent{}, false
//		}
//		return runby.Agent{
//			SessionID: id,
//			Axis:      runby.Axis{Evidence: r.Evidence()},
//		}, true
//	}
//
// The built-in agent drivers are written this way, so a driver supplied through
// Register or WithDrivers reports its evidence exactly as they do.
//
// Names are recorded whether or not they are set, and whether or not the driver
// goes on to match; Evidence resolves to the subset that is set, so an absent
// variable costs nothing and a driver that bails early reports nothing.
//
// An EnvReader is not safe for concurrent use. Create one per Detect call, as
// the example above does, and do not retain it past the call.
type EnvReader struct {
	env   Env
	names []string
}

// NewEnvReader returns an EnvReader over env.
func NewEnvReader(env Env) *EnvReader { return &EnvReader{env: env} }

// Record marks names as consulted without reading them. It is for a variable
// whose value has already been examined some other way, such as one that only
// counts as evidence when it holds a particular value.
func (r *EnvReader) Record(names ...string) { r.names = append(r.names, names...) }

// Value returns the space-trimmed value of name, and whether it is set to a
// non-empty value. A variable set to the empty string is not evidence.
func (r *EnvReader) Value(name string) (string, bool) {
	r.Record(name)
	return envValue(r.env, name)
}

// Peek reads name without recording it, for a variable that is evidence only
// when its value says so. Follow it with Record once the value has decided.
func (r *EnvReader) Peek(name string) (string, bool) { return envValue(r.env, name) }

// Bool returns the boolean held by name, and whether name holds a value that
// strconv.ParseBool accepts.
func (r *EnvReader) Bool(name string) (value, ok bool) {
	r.Record(name)
	return envBool(r.env, name)
}

// IsTrue reports whether name holds a parsable boolean that is true.
func (r *EnvReader) IsTrue(name string) bool {
	value, ok := r.Bool(name)
	return ok && value
}

// EqualsFold reports whether name holds want, ignoring case.
func (r *EnvReader) EqualsFold(name, want string) bool {
	r.Record(name)
	return envEqualsFold(r.env, name, want)
}

// Any reports whether at least one of the names is set. Every name is recorded
// even though the answer may be settled by the first, because a product that
// advertises alternatives is evidenced by whichever ones it happened to set.
func (r *EnvReader) Any(names ...string) bool {
	r.Record(names...)
	return anyPresent(r.env, names...)
}

// First returns the value of the earliest name that is set, or the empty string
// when none is. It is how a product with a preferred source and a fallback is
// read: Orca prefers the worktree it created over the repository it came from,
// and Codex prefers a thread over a session. Every name is recorded, so the
// fallback is reported as evidence even when it lost.
func (r *EnvReader) First(names ...string) string {
	r.Record(names...)
	for _, name := range names {
		if value, ok := envValue(r.env, name); ok {
			return value
		}
	}
	return ""
}

// Extra collects the Axis.Extra map for keys, recording every variable it
// consults. It is the collection helper with the bookkeeping, so a product's context
// variables count as evidence without being listed a second time.
func (r *EnvReader) Extra(keys map[string]string) map[string]string {
	for _, name := range keys {
		r.Record(name)
	}
	return collectExtra(r.env, keys)
}

// Evidence returns the sorted, deduplicated subset of the consulted variables
// that are set. Only names are returned; values may be sensitive.
func (r *EnvReader) Evidence() []string { return presentNames(r.env, r.names...) }
