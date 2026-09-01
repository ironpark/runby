package runby

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

// Env is a read-only view of an environment. A driver's Detect receives an Env
// rather than a map so that it cannot mutate the environment being inspected,
// and so that the process-backed implementation can look values up lazily.
type Env interface {
	// Lookup returns the value bound to name and whether name is set.
	Lookup(name string) (value string, ok bool)
}

// processEnv reads the live process environment. It holds no state, so it
// never goes stale relative to os.Setenv.
type processEnv struct{}

func (processEnv) Lookup(name string) (string, bool) { return os.LookupEnv(name) }

// mapEnv is an Env backed by a parsed environ slice.
type mapEnv map[string]string

func (m mapEnv) Lookup(name string) (string, bool) {
	value, ok := m[name]
	return value, ok
}

// LookupFunc adapts a lookup function, such as os.LookupEnv, to Env:
//
//	runby.Detect(runby.WithEnv(runby.LookupFunc(recorded.Get)))
type LookupFunc func(name string) (value string, ok bool)

func (fn LookupFunc) Lookup(name string) (string, bool) { return fn(name) }

// EnvironEnv builds an Env from a slice of "NAME=value" entries as returned by
// os.Environ. If a name occurs more than once its last value wins, matching
// common environment lookup behavior. Entries without "=" or with an empty name
// are ignored.
func EnvironEnv(environ []string) Env {
	env := make(mapEnv, len(environ))
	for _, entry := range environ {
		name, value, ok := strings.Cut(entry, "=")
		if !ok || name == "" {
			continue
		}
		env[name] = value
	}
	return env
}

// The helpers below are exported so that a driver supplied through
// Register or WithOnlyDrivers parses the environment exactly like the
// built-in ones.

// envValue returns the space-trimmed value of name, and whether it is set to a
// non-empty value. A variable set to the empty string is not evidence.
func envValue(env Env, name string) (string, bool) {
	value, ok := env.Lookup(name)
	value = strings.TrimSpace(value)
	return value, ok && value != ""
}

// envBool returns the boolean held by name, and whether name holds a value that
// strconv.ParseBool accepts.
func envBool(env Env, name string) (value, ok bool) {
	raw, ok := envValue(env, name)
	if !ok {
		return false, false
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, false
	}
	return parsed, true
}

// envIsTrue reports whether name holds a parsable boolean that is true.
func envIsTrue(env Env, name string) bool {
	value, ok := envBool(env, name)
	return ok && value
}

// envEqualsFold reports whether name holds want, ignoring case.
func envEqualsFold(env Env, name, want string) bool {
	value, ok := envValue(env, name)
	return ok && strings.EqualFold(value, want)
}

// collectExtra returns the values of the given variables keyed by the stable
// Extra key each maps to, skipping the ones that are not set. It returns nil
// when none are present, so a detection carrying no context carries no map.
//
// It builds the Extra map of an Agent, CI, Terminal, or Remote the same way
// the built-in drivers do.
func collectExtra(env Env, keys map[string]string) map[string]string {
	var extra map[string]string
	for key, name := range keys {
		value, ok := envValue(env, name)
		if !ok {
			continue
		}
		if extra == nil {
			extra = make(map[string]string, len(keys))
		}
		extra[key] = value
	}
	return extra
}

// presentNames returns the sorted, deduplicated subset of names that are set
// to a non-empty value. Drivers use it to build Evidence, which holds variable
// names only; values may be sensitive and are never copied into it.
//
// Duplicates are dropped because Evidence is a set: a caller often assembles
// the candidate list from several overlapping sources, such as a marker that
// is also read as the session identifier.
func presentNames(env Env, names ...string) []string {
	present := make([]string, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if seen[name] {
			continue
		}
		if _, ok := envValue(env, name); ok {
			seen[name] = true
			present = append(present, name)
		}
	}
	sort.Strings(present)
	return present
}

// marker reports whether an environment shows a product is present. It is the
// first question every driver answers, and it is kept separate from reading
// the product's variables so that recognition and extraction cannot disagree.
type marker func(env Env) bool

// markerSet matches when every name is set to a non-empty value.
func markerSet(names ...string) marker {
	return func(env Env) bool {
		for _, name := range names {
			if _, ok := envValue(env, name); !ok {
				return false
			}
		}
		return true
	}
}

// markerTrue matches when every name holds a parsable boolean that is true.
func markerTrue(names ...string) marker {
	return func(env Env) bool {
		for _, name := range names {
			if !envIsTrue(env, name) {
				return false
			}
		}
		return true
	}
}

// anyPresent reports whether at least one of the names is set to a non-empty
// value. Use it inside a marker for a product that advertises alternatives
// rather than one fixed variable.
func anyPresent(env Env, names ...string) bool {
	for _, name := range names {
		if _, ok := envValue(env, name); ok {
			return true
		}
	}
	return false
}
