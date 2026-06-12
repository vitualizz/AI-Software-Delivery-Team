//go:build !integration

package golden_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/vitualizz/asdt/internal/artifact"
	"gopkg.in/yaml.v3"
)

// --- Local types replacing the deleted requirements/developer packages. ---
// These minimal structs satisfy the golden fixture validation only.

// requirementsSpec mirrors the payload of a requirements-spec artifact.
type requirementsSpec struct {
	UserStories []struct {
		ID     string `yaml:"id"`
		As     string `yaml:"as"`
		Want   string `yaml:"want"`
		SoThat string `yaml:"so_that"`
	} `yaml:"user_stories"`
}

// implementationPlan mirrors the payload of an implementation-plan artifact.
type implementationPlan struct {
	Steps []struct {
		StoryID string `yaml:"story_id"`
	} `yaml:"steps"`
}

// fixtureDir returns the absolute path to the testdata/golden directory.
// It is resolved relative to this test file's location so it works regardless
// of the working directory at test invocation time.
func fixtureDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixtureDir(t), name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

// TestValidRequirementsSpec loads valid-requirements-spec.yaml, validates the
// envelope header, and asserts structural completeness of the payload.
func TestValidRequirementsSpec(t *testing.T) {
	data := readFixture(t, "valid-requirements-spec.yaml")

	var env artifact.Envelope[requirementsSpec]
	if err := yaml.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Envelope header must be valid.
	if err := artifact.Validate(env.EnvelopeHeader); err != nil {
		t.Errorf("Validate: %v", err)
	}

	// Payload must have at least one user story.
	if len(env.Payload.UserStories) == 0 {
		t.Error("UserStories must not be empty")
	}

	// Each story must have all required fields.
	for i, s := range env.Payload.UserStories {
		if s.ID == "" {
			t.Errorf("UserStories[%d].ID is empty", i)
		}
		if s.As == "" {
			t.Errorf("UserStories[%d].As is empty", i)
		}
		if s.Want == "" {
			t.Errorf("UserStories[%d].Want is empty", i)
		}
		if s.SoThat == "" {
			t.Errorf("UserStories[%d].SoThat is empty", i)
		}
	}
}

// TestValidImplementationPlan loads valid-implementation-plan.yaml, validates
// the envelope header, and asserts structural completeness of the payload.
func TestValidImplementationPlan(t *testing.T) {
	data := readFixture(t, "valid-implementation-plan.yaml")

	var env artifact.Envelope[implementationPlan]
	if err := yaml.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Envelope header must be valid.
	if err := artifact.Validate(env.EnvelopeHeader); err != nil {
		t.Errorf("Validate: %v", err)
	}

	// Payload must have at least one step.
	if len(env.Payload.Steps) == 0 {
		t.Error("Steps must not be empty")
	}

	// Each step must reference a story ID.
	for i, step := range env.Payload.Steps {
		if step.StoryID == "" {
			t.Errorf("Steps[%d].StoryID is empty", i)
		}
	}

	// input_refs must chain back to a requirements artifact.
	if len(env.InputRefs) == 0 {
		t.Error("InputRefs must not be empty for an implementation plan")
	}
}

// TestInvalidMissingAgent loads invalid-missing-agent.yaml and verifies that
// artifact.Validate rejects it with an error mentioning "agent".
func TestInvalidMissingAgent(t *testing.T) {
	data := readFixture(t, "invalid-missing-agent.yaml")

	var env artifact.Envelope[requirementsSpec]
	if err := yaml.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	err := artifact.Validate(env.EnvelopeHeader)
	if err == nil {
		t.Fatal("expected Validate to return an error for missing agent, got nil")
	}
	if !containsField(err.Error(), "agent") {
		t.Errorf("error %q should mention field 'agent'", err.Error())
	}
}

// TestInvalidMissingStoryID loads invalid-missing-story-id.yaml and verifies
// that a validation loop catches stories with an empty ID.
func TestInvalidMissingStoryID(t *testing.T) {
	data := readFixture(t, "invalid-missing-story-id.yaml")

	var env artifact.Envelope[requirementsSpec]
	if err := yaml.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Envelope header itself is valid.
	if err := artifact.Validate(env.EnvelopeHeader); err != nil {
		t.Fatalf("header should be valid: %v", err)
	}

	// Custom validation: every story must have a non-empty ID.
	var missingIDs int
	for _, s := range env.Payload.UserStories {
		if s.ID == "" {
			missingIDs++
		}
	}
	if missingIDs == 0 {
		t.Error("expected at least one story with a missing ID, but all IDs are set")
	}
}

// containsField returns true when the string contains the field name.
func containsField(msg, field string) bool {
	return len(msg) > 0 && containsSubstring(msg, field)
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
