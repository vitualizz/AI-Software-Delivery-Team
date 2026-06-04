package specialists_test

import (
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/specialists"
)

// descriptorCase holds the input/expectation for one descriptor.
type descriptorCase struct {
	name       string
	descriptor specialists.SpecialistDescriptor
	// emptyReads asserts that Artifacts.Reads must be empty (security invariant).
	emptyReads bool
}

func allDescriptors() []descriptorCase {
	return []descriptorCase{
		{name: "developer", descriptor: specialists.DeveloperDescriptor()},
		{name: "ux-ui", descriptor: specialists.UXUIDescriptor()},
		{name: "architect", descriptor: specialists.ArchitectDescriptor()},
		{name: "qa", descriptor: specialists.QADescriptor()},
		{name: "security", descriptor: specialists.SecurityDescriptor(), emptyReads: true},
	}
}

// TestDescriptorTable runs structural assertions over all 5 descriptors.
func TestDescriptorTable(t *testing.T) {
	for _, tc := range allDescriptors() {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			d := tc.descriptor

			if d.ID == "" {
				t.Error("ID must not be empty")
			}
			if d.Name == "" {
				t.Error("Name must not be empty")
			}
			if len(d.Workflow) == 0 {
				t.Error("Workflow must not be empty")
			}
			if len(d.Artifacts.Writes) == 0 {
				t.Error("Artifacts.Writes must not be empty")
			}

			// No duplicate step IDs.
			seen := make(map[string]bool)
			for _, step := range d.Workflow {
				if step.ID == "" {
					t.Errorf("workflow step has empty ID")
				}
				if seen[step.ID] {
					t.Errorf("duplicate workflow step ID %q", step.ID)
				}
				seen[step.ID] = true
			}

			// SecurityDescriptor invariant: empty Reads means no required predecessor.
			if tc.emptyReads && len(d.Artifacts.Reads) != 0 {
				t.Errorf("expected empty Reads (no required predecessor), got %v", d.Artifacts.Reads)
			}

			// Validate() must return nil for a well-formed descriptor.
			if err := d.Validate(); err != nil {
				t.Errorf("Validate() returned error: %v", err)
			}
		})
	}
}

// TestDeveloperDescriptor_ExactSteps verifies the Developer descriptor has exactly 7 steps
// with the expected IDs in order.
func TestDeveloperDescriptor_ExactSteps(t *testing.T) {
	d := specialists.DeveloperDescriptor()
	want := []string{
		"artifact-loading",
		"platform-context",
		"complexity-estimate",
		"implementation-planning",
		"code-generation",
		"test-generation",
		"self-review",
	}
	if len(d.Workflow) != len(want) {
		t.Fatalf("Developer: expected %d workflow steps, got %d", len(want), len(d.Workflow))
	}
	for i, wantID := range want {
		if d.Workflow[i].ID != wantID {
			t.Errorf("Workflow[%d]: expected ID %q, got %q", i, wantID, d.Workflow[i].ID)
		}
	}
}

// TestSecurityDescriptor_NoRequiredPredecessor verifies the Security invariant explicitly.
func TestSecurityDescriptor_NoRequiredPredecessor(t *testing.T) {
	d := specialists.SecurityDescriptor()
	if len(d.Artifacts.Reads) != 0 {
		t.Errorf("SecurityDescriptor.Artifacts.Reads must be empty, got %v", d.Artifacts.Reads)
	}
}

// TestDescriptorValidate_RejectsEmptyID verifies Validate rejects a descriptor with empty ID.
func TestDescriptorValidate_RejectsEmptyID(t *testing.T) {
	d := specialists.SpecialistDescriptor{
		ID:   "",
		Name: "Test",
		Workflow: []specialists.WorkflowStep{
			{ID: "step-1", Description: "a step"},
		},
	}
	if err := d.Validate(); err == nil {
		t.Error("Validate() should return error for empty ID")
	}
}

// TestDescriptorValidate_RejectsDuplicateStepIDs verifies Validate catches duplicate step IDs.
func TestDescriptorValidate_RejectsDuplicateStepIDs(t *testing.T) {
	d := specialists.SpecialistDescriptor{
		ID:   "test",
		Name: "Test",
		Workflow: []specialists.WorkflowStep{
			{ID: "step-1", Description: "first"},
			{ID: "step-1", Description: "duplicate"},
		},
	}
	if err := d.Validate(); err == nil {
		t.Error("Validate() should return error for duplicate step IDs")
	}
}
