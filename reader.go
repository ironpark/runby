package runby

// reader is an Env paired with the record of what has been asked of it.
//
// Every driver must report the variables it consulted as Evidence, and the CI,
// terminal, and remote axes get that for free: a spec declares its variables
// as data, so reading them and reporting them are the same list. The agent
// axis is written as functions instead, because its rules are irregular enough
// that a table would need an escape hatch per agent — see spec.go. That leaves
// its evidence to be kept in step by hand, which is the one thing about the
// arrangement that can silently rot: adding a lookup and forgetting to widen
// the evidence list produces a detection that under-reports why it fired, and
// nothing fails.
//
// A reader closes that gap without a table. Asking it for a value records the
// name, so a driver's evidence is what it actually read rather than a second
// list maintained alongside. Add a lookup and the evidence follows.
//
// Names are recorded whether or not they are set, and whether or not the
// driver goes on to match; evidence resolves to the subset that is set, so an
// absent variable costs nothing and a driver that bails early never reports.
type reader struct {
	env   Env
	names []string
}

func newReader(env Env) *reader { return &reader{env: env} }

// record marks names as consulted without reading them. It is for a variable
// whose value has already been examined some other way, such as one that only
// counts as evidence when it holds a particular value.
func (r *reader) record(names ...string) { r.names = append(r.names, names...) }

// value returns the space-trimmed value of name, and whether it is set to a
// non-empty value, as Value does.
func (r *reader) value(name string) (string, bool) {
	r.record(name)
	return Value(r.env, name)
}

// peek reads name without recording it, for a variable that is evidence only
// when its value says so. Follow it with record once the value has decided.
func (r *reader) peek(name string) (string, bool) { return Value(r.env, name) }

// boolean returns the boolean held by name, and whether name holds a value
// that strconv.ParseBool accepts, as Bool does.
func (r *reader) boolean(name string) (value, ok bool) {
	r.record(name)
	return Bool(r.env, name)
}

// isTrue reports whether name holds a parsable boolean that is true.
func (r *reader) isTrue(name string) bool {
	value, ok := r.boolean(name)
	return ok && value
}

// equalsFold reports whether name holds want, ignoring case.
func (r *reader) equalsFold(name, want string) bool {
	r.record(name)
	return EqualsFold(r.env, name, want)
}

// any reports whether at least one of the names is set. Every name is recorded
// even though the answer may be settled by the first, because a product that
// advertises alternatives is evidenced by whichever ones it happened to set.
func (r *reader) any(names ...string) bool {
	r.record(names...)
	return AnyPresent(r.env, names...)
}

// first returns the value of the earliest name that is set, or the empty
// string when none is. It is how a product with a preferred source and a
// fallback is read: Orca prefers the worktree it created over the repository
// it came from, and Codex prefers a thread over a session. Every name is
// recorded, so the fallback is reported as evidence even when it lost.
func (r *reader) first(names ...string) string {
	r.record(names...)
	for _, name := range names {
		if value, ok := Value(r.env, name); ok {
			return value
		}
	}
	return ""
}

// extra collects the Extra map for keys, recording every variable it consults.
// It is CollectExtra with the bookkeeping, so a product's context variables
// count as evidence without being listed a second time.
func (r *reader) extra(keys map[string]string) map[string]string {
	for _, name := range keys {
		r.record(name)
	}
	return CollectExtra(r.env, keys)
}

// evidence returns the sorted, deduplicated subset of the consulted variables
// that are set. Only names are returned; values may be sensitive.
func (r *reader) evidence() []string { return PresentNames(r.env, r.names...) }
