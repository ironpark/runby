package main

import "testing"

func TestCaptureHashesValuesAndKeepsNames(t *testing.T) {
	got := capture([]string{"A=one", "B=two", "A=latest", "INVALID"})
	if got.Algorithm != "sha256" {
		t.Fatalf("Algorithm = %q", got.Algorithm)
	}
	if len(got.Variables) != 2 {
		t.Fatalf("len(Variables) = %d, want 2", len(got.Variables))
	}
	if got.Variables["A"] == "latest" || got.Variables["B"] == "two" {
		t.Fatal("snapshot contains a raw value")
	}
	if got.Variables["A"] == capture([]string{"A=one"}).Variables["A"] {
		t.Fatal("last duplicate value did not win")
	}
}
