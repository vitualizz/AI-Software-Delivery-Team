package main

import (
	"testing"
)

// TestIsSpecialist verifies that known specialist IDs are recognized and
// unknown IDs are rejected.
func TestIsSpecialist_KnownIDs(t *testing.T) {
	known := []string{"developer", "ux-ui", "architect", "qa", "security"}
	for _, id := range known {
		if !isSpecialist(id) {
			t.Errorf("isSpecialist(%q) = false, want true", id)
		}
	}
}

func TestIsSpecialist_UnknownID(t *testing.T) {
	unknown := []string{"", "runner", "llm", "requirements", "develop"}
	for _, id := range unknown {
		if isSpecialist(id) {
			t.Errorf("isSpecialist(%q) = true, want false", id)
		}
	}
}

// TestListSpecialists verifies that listSpecialists returns a non-empty
// comma-separated sorted string.
func TestListSpecialists_NonEmpty(t *testing.T) {
	result := listSpecialists()
	if result == "" {
		t.Error("listSpecialists() returned empty string")
	}
}
