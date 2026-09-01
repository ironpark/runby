package runby

import (
	"reflect"
	"testing"
)

// The env helpers are unexported — drivers read the environment through an
// EnvReader, which wraps them — so their unit tests live inside the package.

func TestMarkerHelpers(t *testing.T) {
	env := EnvironEnv([]string{"A=1", "B=false", "C=x", "EMPTY=", "TERM_PROGRAM=WezTerm"})
	for _, test := range []struct {
		name   string
		marker marker
		want   bool
	}{
		{"set both", markerSet("A", "C"), true},
		{"set missing", markerSet("A", "MISSING"), false},
		{"set empty is not set", markerSet("EMPTY"), false},
		{"true", markerTrue("A"), true},
		{"true on false", markerTrue("A", "B"), false},
		{"true on non-boolean", markerTrue("C"), false},
		{"term program folds case", markerTermProgram("wezterm"), true},
		{"term program mismatch", markerTermProgram("ghostty"), false},
	} {
		if got := test.marker(env); got != test.want {
			t.Errorf("%s = %v, want %v", test.name, got, test.want)
		}
	}

	if !anyPresent(env, "MISSING", "C") {
		t.Error("anyPresent missed a set name")
	}
	if anyPresent(env, "MISSING", "EMPTY") {
		t.Error("anyPresent matched an unset name")
	}
	if anyPresent(env) {
		t.Error("anyPresent matched with no names")
	}
}

func TestCollectExtra(t *testing.T) {
	env := EnvironEnv([]string{"A=1", "EMPTY="})
	got := collectExtra(env, map[string]string{"acme.a": "A", "acme.missing": "MISSING", "acme.empty": "EMPTY"})
	if want := map[string]string{"acme.a": "1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("collectExtra = %#v, want %#v", got, want)
	}
	// Nothing present carries no map, so a detection without context has none.
	if got := collectExtra(env, map[string]string{"acme.missing": "MISSING"}); got != nil {
		t.Fatalf("collectExtra = %#v, want nil", got)
	}
}
