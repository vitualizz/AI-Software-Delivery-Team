// Package specialists implements the data-driven specialist model.
// A SpecialistDescriptor is pure data interpreted by a generic Runner.
// Adding a new specialist requires one descriptor constructor and one
// skill/{id}/ tree — no new Go packages or switch arms.
package specialists

import (
	"errors"
	"fmt"
)

// SpecialistDescriptor is the complete data definition of one professional role.
// It is pure data: no methods that perform I/O. The Runner interprets it.
type SpecialistDescriptor struct {
	// ID is the stable key used for registry lookup and artifact agent fields.
	// Examples: "developer", "ux-ui", "architect", "qa", "security".
	ID string

	// Name is the human-readable label shown in CLI output.
	Name string

	// Description contains trigger keywords for the meta-orchestrator.
	Description string

	// Skills lists shared skill IDs loaded for every workflow step.
	// These are resolved via registry.Skill (from _shared/skills/).
	Skills []string

	// Workflow is the ordered list of discipline-specific steps.
	Workflow []WorkflowStep

	// Artifacts declares the artifact types this specialist reads and writes.
	Artifacts ArtifactContract
}

// WorkflowStep is one ordered step in a specialist's process.
type WorkflowStep struct {
	// ID is the unique key within the descriptor; used as the pipeline step key.
	ID string

	// Description describes what this step accomplishes.
	Description string

	// SkillRefs lists specialist-scoped skill IDs activated at this step.
	// Resolved via registry.ScopedSkill(descriptor.ID, skillRef).
	SkillRefs []string
}

// ArtifactContract declares the artifact types a specialist consumes and produces.
type ArtifactContract struct {
	// Reads lists artifact type IDs read from .asdt/artifacts/{change}/.
	// All reads are soft: a missing input produces an open_items note, not an error.
	Reads []string

	// Writes lists artifact type IDs this specialist writes.
	Writes []string
}

// Validate enforces descriptor well-formedness at startup (fail fast).
// It checks: non-empty ID and Name, unique step IDs within Workflow.
func (d SpecialistDescriptor) Validate() error {
	if d.ID == "" {
		return errors.New("specialist descriptor: ID must not be empty")
	}
	if d.Name == "" {
		return fmt.Errorf("specialist %s: Name must not be empty", d.ID)
	}
	seen := make(map[string]bool, len(d.Workflow))
	for _, step := range d.Workflow {
		if step.ID == "" {
			return fmt.Errorf("specialist %s: workflow step has empty ID", d.ID)
		}
		if seen[step.ID] {
			return fmt.Errorf("specialist %s: duplicate workflow step ID %q", d.ID, step.ID)
		}
		seen[step.ID] = true
	}
	return nil
}

// DeveloperDescriptor returns the full descriptor for the Developer specialist.
// It matches the 7-step workflow defined in skill/developer/SKILL.md exactly.
func DeveloperDescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "developer",
		Name:        "Developer",
		Description: "Trigger: developer, implement, code, build, feature, write tests, ship",
		Skills:      []string{"platform-context", "artifact-envelope"},
		Workflow: []WorkflowStep{
			{
				ID:          "artifact-loading",
				Description: "Load any upstream artifacts in .asdt/artifacts/{change}/ (spec, design, ux). Missing = note in open_items, continue.",
				SkillRefs:   []string{"artifact-loading"},
			},
			{
				ID:          "platform-context",
				Description: "Load platform.yaml via the platform-context shared skill. Inject detected stack, naming conventions, and design fingerprint into working context.",
				SkillRefs:   []string{},
			},
			{
				ID:          "complexity-estimate",
				Description: "Estimate implementation complexity from available context. Output: S | M | L | XL with rationale.",
				SkillRefs:   []string{},
			},
			{
				ID:          "implementation-planning",
				Description: "Produce ordered implementation steps, one per upstream story/requirement where present.",
				SkillRefs:   []string{"code-generation"},
			},
			{
				ID:          "code-generation",
				Description: "Generate code snippets and file create/modify lists for each implementation step.",
				SkillRefs:   []string{"code-generation"},
			},
			{
				ID:          "test-generation",
				Description: "Generate test cases covering each implementation step.",
				SkillRefs:   []string{"test-generation"},
			},
			{
				ID:          "self-review",
				Description: "Verify traceability: every read story appears in a step; unmatched go to open_items.",
				SkillRefs:   []string{},
			},
		},
		Artifacts: ArtifactContract{
			Reads:  []string{"requirements-spec", "system-design", "ux-brief"},
			Writes: []string{"implementation-plan"},
		},
	}
}

// UXUIDescriptor returns the descriptor for the UX/UI Specialist.
// Matches the 7-step workflow in skill/ux-ui/SKILL.md.
func UXUIDescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "ux-ui",
		Name:        "UX/UI Designer",
		Description: "Trigger: ux, ui, design, wireframe, layout, user flow, accessibility, responsive",
		Skills:      []string{"platform-context", "artifact-envelope"},
		Workflow: []WorkflowStep{
			{
				ID:          "platform-analysis",
				Description: "Establish/refresh platform context. Scan for existing UI patterns.",
				SkillRefs:   []string{},
			},
			{
				ID:          "feature-brief",
				Description: "Capture the UX intent: core user problem, primary actor, success criteria.",
				SkillRefs:   []string{},
			},
			{
				ID:          "information-architecture",
				Description: "Define screen/content hierarchy, navigation path, and data relationships.",
				SkillRefs:   []string{"information-architecture"},
			},
			{
				ID:          "user-flows",
				Description: "Map primary happy-path and 2–3 edge-case flows with IDs (UF-001, UF-002, …).",
				SkillRefs:   []string{"information-architecture"},
			},
			{
				ID:          "component-mapping",
				Description: "Map flows to reusable components: reused, extended, or new.",
				SkillRefs:   []string{},
			},
			{
				ID:          "responsive-strategy",
				Description: "Define breakpoints and adaptive behavior using responsive-design and accessibility guidelines.",
				SkillRefs:   []string{"responsive-design", "accessibility"},
			},
			{
				ID:          "ux-handoff",
				Description: "Summarize decisions, open questions, and handoff notes for Developer and Architect.",
				SkillRefs:   []string{"accessibility"},
			},
		},
		Artifacts: ArtifactContract{
			Reads:  []string{"requirements-spec"},
			Writes: []string{"ux-brief", "component-spec"},
		},
	}
}

// ArchitectDescriptor returns the descriptor for the Software Architect specialist.
// Matches the 7-step workflow in skill/architect/SKILL.md.
func ArchitectDescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "architect",
		Name:        "Software Architect",
		Description: "Trigger: architecture, design system, adr, scalability, api design, tradeoff",
		Skills:      []string{"platform-context", "artifact-envelope", "scope-definition"},
		Workflow: []WorkflowStep{
			{
				ID:          "platform-analysis",
				Description: "Establish platform context. Extract language, framework, persistence, service boundaries, and architectural constraints.",
				SkillRefs:   []string{"architecture-review"},
			},
			{
				ID:          "load-constraints",
				Description: "Read upstream artifacts and pin the change boundary using scope-definition.",
				SkillRefs:   []string{},
			},
			{
				ID:          "evaluate-approaches",
				Description: "Compare 2–3 candidate technical approaches with tradeoffs.",
				SkillRefs:   []string{"architecture-review"},
			},
			{
				ID:          "decision-record",
				Description: "Record the chosen approach as an Architecture Decision Record.",
				SkillRefs:   []string{"architecture-review"},
			},
			{
				ID:          "system-design",
				Description: "Produce data model, API surface, service boundaries, and sequence diagrams.",
				SkillRefs:   []string{"api-design"},
			},
			{
				ID:          "risk-analysis",
				Description: "Identify top 3–5 risks with likelihood, impact, and concrete mitigation.",
				SkillRefs:   []string{"scalability-analysis"},
			},
			{
				ID:          "technical-handoff",
				Description: "Summarize constraints, decisions, suggested implementation order, and open questions.",
				SkillRefs:   []string{},
			},
		},
		Artifacts: ArtifactContract{
			Reads:  []string{"requirements-spec", "ux-brief"},
			Writes: []string{"architectural-decision", "system-design", "risk-register"},
		},
	}
}

// QADescriptor returns the descriptor for the QA Engineer specialist.
// Matches the 6-step workflow in skill/qa/SKILL.md (note: SKILL.md has 6 steps, not 7).
func QADescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "qa",
		Name:        "QA Engineer",
		Description: "Trigger: qa, test, acceptance criteria, edge case, coverage, quality",
		Skills:      []string{"platform-context", "artifact-envelope"},
		Workflow: []WorkflowStep{
			{
				ID:          "load-requirements",
				Description: "Read spec/AC artifacts: requirements-spec, ux-brief, architectural-decision, system-design.",
				SkillRefs:   []string{},
			},
			{
				ID:          "ac-validation",
				Description: "Validate acceptance criteria completeness, testability, clarity, and measurability.",
				SkillRefs:   []string{"acceptance-criteria"},
			},
			{
				ID:          "edge-case-analysis",
				Description: "Enumerate boundary values, error paths, concurrent scenarios, and data edge cases.",
				SkillRefs:   []string{"edge-case-analysis"},
			},
			{
				ID:          "test-strategy",
				Description: "Define unit/integration/E2E strategy, coverage targets, and test data approach.",
				SkillRefs:   []string{"test-strategy"},
			},
			{
				ID:          "test-cases",
				Description: "Author concrete test cases in Given/When/Then format covering happy path and edge cases.",
				SkillRefs:   []string{"test-strategy"},
			},
			{
				ID:          "quality-handoff",
				Description: "Summarize AC gaps, missing test infrastructure, and open questions for Developer.",
				SkillRefs:   []string{},
			},
		},
		Artifacts: ArtifactContract{
			Reads:  []string{"requirements-spec", "implementation-plan"},
			Writes: []string{"test-plan", "quality-report"},
		},
	}
}

// SecurityDescriptor returns the descriptor for the Security Engineer specialist.
// Key invariant: Reads is empty — no required predecessor. Matches skill/security/SKILL.md.
func SecurityDescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "security",
		Name:        "Security Engineer",
		Description: "Trigger: security, threat model, owasp, vulnerability, hardening, attack surface",
		Skills:      []string{"platform-context", "artifact-envelope"},
		Workflow: []WorkflowStep{
			{
				ID:          "threat-modeling",
				Description: "STRIDE-style threat model from any available context.",
				SkillRefs:   []string{"threat-modeling"},
			},
			{
				ID:          "attack-surface-review",
				Description: "Enumerate all entry points and map trust boundaries.",
				SkillRefs:   []string{"threat-modeling"},
			},
			{
				ID:          "owasp-analysis",
				Description: "Review the feature against the OWASP Top 10 2021 categories relevant to the stack.",
				SkillRefs:   []string{"owasp-review"},
			},
			{
				ID:          "security-findings",
				Description: "Record findings with severity, CWE reference, and actionable recommendation.",
				SkillRefs:   []string{"owasp-review"},
			},
			{
				ID:          "hardening-checklist",
				Description: "Produce an ordered hardening action list for the Developer specialist.",
				SkillRefs:   []string{},
			},
		},
		Artifacts: ArtifactContract{
			Reads:  []string{}, // no required predecessor — can run at any lifecycle stage
			Writes: []string{"threat-model", "security-findings", "hardening-checklist"},
		},
	}
}
