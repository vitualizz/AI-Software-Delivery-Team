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
		"explore",
		"spec",
		"design",
		"tasks",
		"implement",
		"test",
		"review",
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

// TestDeveloperDescriptor_InputRefs verifies per-step InputRefs and OutputArtifact values
// for the Developer descriptor's critical steps.
func TestDeveloperDescriptor_InputRefs(t *testing.T) {
	d := specialists.DeveloperDescriptor()

	// Step index 2 = "design" — must read ONLY developer/dev-spec.
	designStep := d.Workflow[2]
	if designStep.ID != "design" {
		t.Fatalf("expected Workflow[2] to be 'design', got %q", designStep.ID)
	}
	if len(designStep.InputRefs) != 1 || designStep.InputRefs[0] != "developer/dev-spec" {
		t.Errorf("design step InputRefs = %v, want [developer/dev-spec]", designStep.InputRefs)
	}
	if designStep.OutputArtifact != "developer/dev-design" {
		t.Errorf("design step OutputArtifact = %q, want %q", designStep.OutputArtifact, "developer/dev-design")
	}

	// All non-last steps must have non-empty OutputArtifact.
	for i, step := range d.Workflow {
		if i < len(d.Workflow)-1 && step.OutputArtifact == "" {
			t.Errorf("non-last step %q (index %d) has empty OutputArtifact", step.ID, i)
		}
	}

	// Last step (review) must have non-empty OutputArtifact (implementation-plan is final).
	lastStep := d.Workflow[len(d.Workflow)-1]
	if lastStep.OutputArtifact == "" {
		t.Errorf("last step %q OutputArtifact must not be empty for developer (uses per-step write)", lastStep.ID)
	}
}

// TestSecurityDescriptor_NoRequiredPredecessor verifies the Security invariant explicitly.
func TestSecurityDescriptor_NoRequiredPredecessor(t *testing.T) {
	d := specialists.SecurityDescriptor()
	if len(d.Artifacts.Reads) != 0 {
		t.Errorf("SecurityDescriptor.Artifacts.Reads must be empty, got %v", d.Artifacts.Reads)
	}
}

// TestSecurityDescriptor_FirstStepIsPlatformAnalysis verifies the Security pipeline starts
// with a platform-analysis step.
func TestSecurityDescriptor_FirstStepIsPlatformAnalysis(t *testing.T) {
	d := specialists.SecurityDescriptor()
	if len(d.Workflow) == 0 {
		t.Fatal("SecurityDescriptor has no workflow steps")
	}
	if d.Workflow[0].ID != "platform-analysis" {
		t.Errorf("SecurityDescriptor first step = %q, want %q", d.Workflow[0].ID, "platform-analysis")
	}
}

// TestDescriptorTable_AllStepsHaveNonEmptyID verifies that every step across all
// descriptors has a non-empty ID (extends the table test with an explicit coverage note).
func TestDescriptorTable_InputRefsPresent(t *testing.T) {
	for _, tc := range allDescriptors() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for i, step := range tc.descriptor.Workflow {
				if step.ID == "" {
					t.Errorf("step[%d] has empty ID", i)
				}
				// InputRefs field must be non-nil (may be empty slice for first steps).
				if step.InputRefs == nil {
					t.Errorf("step[%d] %q: InputRefs must not be nil (use []string{} for empty)", i, step.ID)
				}
			}
		})
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

// TestSkipIfInitialized_ZeroValue verifies that a WorkflowStep with no
// SkipIfInitialized set defaults to false (backward compatible).
func TestSkipIfInitialized_ZeroValue(t *testing.T) {
	step := specialists.WorkflowStep{
		ID:             "step-1",
		Description:    "a step",
		InputRefs:      []string{},
		OutputArtifact: "some/artifact",
	}
	if step.SkipIfInitialized {
		t.Error("SkipIfInitialized zero value must be false")
	}
}

// TestSkipIfInitialized_FlaggedSteps verifies that exactly the three platform-analysis
// steps in UX-UI, Security, and Architect descriptors have SkipIfInitialized: true,
// and that Developer.explore and Architect.load-constraints do NOT.
func TestSkipIfInitialized_FlaggedSteps(t *testing.T) {
	findStep := func(d specialists.SpecialistDescriptor, stepID string) (specialists.WorkflowStep, bool) {
		for _, s := range d.Workflow {
			if s.ID == stepID {
				return s, true
			}
		}
		return specialists.WorkflowStep{}, false
	}

	// UX-UI platform-analysis must be flagged.
	uxui := specialists.UXUIDescriptor()
	if s, ok := findStep(uxui, "platform-analysis"); !ok {
		t.Error("UXUIDescriptor: expected platform-analysis step")
	} else if !s.SkipIfInitialized {
		t.Error("UXUIDescriptor.platform-analysis: SkipIfInitialized must be true")
	}

	// Security platform-analysis must be flagged.
	sec := specialists.SecurityDescriptor()
	if s, ok := findStep(sec, "platform-analysis"); !ok {
		t.Error("SecurityDescriptor: expected platform-analysis step")
	} else if !s.SkipIfInitialized {
		t.Error("SecurityDescriptor.platform-analysis: SkipIfInitialized must be true")
	}

	// Architect platform-analysis must be flagged.
	arch := specialists.ArchitectDescriptor()
	if s, ok := findStep(arch, "platform-analysis"); !ok {
		t.Error("ArchitectDescriptor: expected platform-analysis step")
	} else if !s.SkipIfInitialized {
		t.Error("ArchitectDescriptor.platform-analysis: SkipIfInitialized must be true")
	}

	// Architect load-constraints must NOT be flagged (deliberate design decision).
	if s, ok := findStep(arch, "load-constraints"); !ok {
		t.Error("ArchitectDescriptor: expected load-constraints step")
	} else if s.SkipIfInitialized {
		t.Error("ArchitectDescriptor.load-constraints: SkipIfInitialized must be false")
	}

	// Developer explore must NOT be flagged.
	dev := specialists.DeveloperDescriptor()
	if s, ok := findStep(dev, "explore"); !ok {
		t.Error("DeveloperDescriptor: expected explore step")
	} else if s.SkipIfInitialized {
		t.Error("DeveloperDescriptor.explore: SkipIfInitialized must be false")
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
