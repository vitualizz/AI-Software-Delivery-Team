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

	// InputRefs lists artifact types this step reads. Empty = fall back to
	// descriptor.Artifacts.Reads (backward compat).
	InputRefs []string

	// OutputArtifact is the artifact type written immediately when this step
	// completes. Empty = no mid-run write (until the last step, which falls
	// back to descriptor.Artifacts.Writes via writeArtifacts).
	OutputArtifact string

	// SkipIfInitialized, when true, causes the runner to skip the LLM call for
	// this step if .asdt/knowledge/platform-summary.yaml exists, emitting the
	// summary as the step's output artifact instead.
	// Zero value (false) preserves prior behavior — the step always executes.
	SkipIfInitialized bool
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
// It matches the 7-step context-isolated pipeline defined in skill/developer/SKILL.md.
// Each step declares exactly which artifacts it reads (InputRefs) and writes (OutputArtifact).
func DeveloperDescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "developer",
		Name:        "Developer",
		Description: "Trigger: developer, implement, code, build, feature, write tests, ship",
		Skills:      []string{"platform-context", "artifact-envelope"},
		Workflow: []WorkflowStep{
			{
				ID:             "explore",
				Description:    "Survey request + platform context; identify scope, unknowns, and relevant patterns.",
				SkillRefs:      []string{"platform-analysis", "artifact-loading"},
				InputRefs:      []string{},
				OutputArtifact: "developer/dev-exploration",
			},
			{
				ID:             "spec",
				Description:    "Define what to build: stories and requirements derived from exploration.",
				SkillRefs:      []string{},
				InputRefs:      []string{"developer/dev-exploration"},
				OutputArtifact: "developer/dev-spec",
			},
			{
				ID:             "design",
				Description:    "Define how: components, file targets, and interfaces — no code yet.",
				SkillRefs:      []string{},
				InputRefs:      []string{"developer/dev-spec"},
				OutputArtifact: "developer/dev-design",
			},
			{
				ID:             "tasks",
				Description:    "Ordered, atomic implementation steps mapped to spec and design.",
				SkillRefs:      []string{"scope-definition"},
				InputRefs:      []string{"developer/dev-spec", "developer/dev-design"},
				OutputArtifact: "developer/dev-tasks",
			},
			{
				ID:             "implement",
				Description:    "Generate code snippets and create/modify file lists per task.",
				SkillRefs:      []string{"code-generation"},
				InputRefs:      []string{"developer/dev-tasks", "developer/dev-design"},
				OutputArtifact: "developer/dev-implementation",
			},
			{
				ID:             "test",
				Description:    "Generate test cases covering each implementation unit.",
				SkillRefs:      []string{"test-generation"},
				InputRefs:      []string{"developer/dev-tasks", "developer/dev-implementation"},
				OutputArtifact: "developer/dev-tests",
			},
			{
				ID:             "review",
				Description:    "Self-review coverage and conventions; assemble final implementation-plan.",
				SkillRefs:      []string{"review", "report"},
				InputRefs:      []string{"developer/dev-implementation", "developer/dev-tests"},
				OutputArtifact: "implementation-plan",
			},
		},
		Artifacts: ArtifactContract{
			Reads:  []string{"requirements-spec", "system-design", "ux-brief"},
			Writes: []string{"implementation-plan"},
		},
	}
}

// UXUIDescriptor returns the descriptor for the UX/UI Specialist.
// Matches the 7-step context-isolated pipeline in skill/ux-ui/SKILL.md.
// Last step uses OutputArtifact="" so writeArtifacts writes all Artifacts.Writes (ux-brief + component-spec).
func UXUIDescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "ux-ui",
		Name:        "UX/UI Designer",
		Description: "Trigger: ux, ui, design, wireframe, layout, user flow, accessibility, responsive",
		Skills:      []string{"platform-context", "artifact-envelope"},
		Workflow: []WorkflowStep{
			{
				ID:                "platform-analysis",
				Description:       "Establish/refresh platform context. Scan for existing UI patterns.",
				SkillRefs:         []string{},
				InputRefs:         []string{},
				OutputArtifact:    "platform-summary",
				SkipIfInitialized: true,
			},
			{
				ID:             "feature-brief",
				Description:    "Capture the UX intent: core user problem, primary actor, success criteria.",
				SkillRefs:      []string{},
				InputRefs:      []string{"platform-summary"},
				OutputArtifact: "ux-ui/feature-brief",
			},
			{
				ID:             "information-architecture",
				Description:    "Define screen/content hierarchy, navigation path, and data relationships.",
				SkillRefs:      []string{"information-architecture"},
				InputRefs:      []string{"ux-ui/feature-brief", "platform-summary"},
				OutputArtifact: "ux-ui/ia",
			},
			{
				ID:             "user-flows",
				Description:    "Map primary happy-path and 2–3 edge-case flows with IDs (UF-001, UF-002, …).",
				SkillRefs:      []string{"information-architecture"},
				InputRefs:      []string{"ux-ui/ia"},
				OutputArtifact: "ux-ui/flows",
			},
			{
				ID:             "component-mapping",
				Description:    "Map flows to reusable components: reused, extended, or new.",
				SkillRefs:      []string{},
				InputRefs:      []string{"ux-ui/flows", "platform-summary"},
				OutputArtifact: "ux-ui/components",
			},
			{
				ID:             "responsive-strategy",
				Description:    "Define breakpoints and adaptive behavior using responsive-design and accessibility guidelines.",
				SkillRefs:      []string{"responsive-design", "accessibility"},
				InputRefs:      []string{"ux-ui/components"},
				OutputArtifact: "ux-ui/responsive",
			},
			{
				ID:          "ux-handoff",
				Description: "Summarize decisions, open questions, and handoff notes for Developer and Architect.",
				SkillRefs:   []string{"accessibility"},
				InputRefs:   []string{"ux-ui/feature-brief", "ux-ui/ia", "ux-ui/flows", "ux-ui/components", "ux-ui/responsive"},
				// OutputArtifact="" → last-step fallback writes ux-brief + component-spec via Artifacts.Writes.
				OutputArtifact: "",
			},
		},
		Artifacts: ArtifactContract{
			Reads:  []string{"requirements-spec"},
			Writes: []string{"ux-brief", "component-spec"},
		},
	}
}

// ArchitectDescriptor returns the descriptor for the Software Architect specialist.
// Matches the 7-step context-isolated pipeline in skill/architect/SKILL.md.
// Last step uses OutputArtifact="" so writeArtifacts writes all three finals.
func ArchitectDescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "architect",
		Name:        "Software Architect",
		Description: "Trigger: architecture, design system, adr, scalability, api design, tradeoff",
		Skills:      []string{"platform-context", "artifact-envelope", "scope-definition"},
		Workflow: []WorkflowStep{
			{
				ID:                "platform-analysis",
				Description:       "Establish platform context. Extract language, framework, persistence, service boundaries, and architectural constraints.",
				SkillRefs:         []string{"architecture-review"},
				InputRefs:         []string{},
				OutputArtifact:    "architect/constraints",
				SkipIfInitialized: true,
			},
			{
				ID:             "load-constraints",
				Description:    "Read upstream artifacts and pin the change boundary using scope-definition.",
				SkillRefs:      []string{},
				InputRefs:      []string{"architect/constraints"},
				OutputArtifact: "architect/constraints-analysis",
			},
			{
				ID:             "evaluate-approaches",
				Description:    "Compare 2–3 candidate technical approaches with tradeoffs.",
				SkillRefs:      []string{"architecture-review"},
				InputRefs:      []string{"architect/constraints-analysis"},
				OutputArtifact: "architect/approaches",
			},
			{
				ID:             "decision-record",
				Description:    "Record the chosen approach as an Architecture Decision Record.",
				SkillRefs:      []string{"architecture-review"},
				InputRefs:      []string{"architect/approaches"},
				OutputArtifact: "architect/adr",
			},
			{
				ID:             "system-design",
				Description:    "Produce data model, API surface, service boundaries, and sequence diagrams.",
				SkillRefs:      []string{"api-design"},
				InputRefs:      []string{"architect/adr"},
				OutputArtifact: "architect/system-design",
			},
			{
				ID:             "risk-analysis",
				Description:    "Identify top 3–5 risks with likelihood, impact, and concrete mitigation.",
				SkillRefs:      []string{"scalability-analysis"},
				InputRefs:      []string{"architect/system-design"},
				OutputArtifact: "architect/risks",
			},
			{
				ID:          "technical-handoff",
				Description: "Summarize constraints, decisions, suggested implementation order, and open questions.",
				SkillRefs:   []string{},
				InputRefs:   []string{"architect/adr", "architect/system-design", "architect/risks"},
				// OutputArtifact="" → last-step fallback writes architectural-decision + system-design + risk-register.
				OutputArtifact: "",
			},
		},
		Artifacts: ArtifactContract{
			Reads:  []string{"requirements-spec", "ux-brief"},
			Writes: []string{"architectural-decision", "system-design", "risk-register"},
		},
	}
}

// QADescriptor returns the descriptor for the QA Engineer specialist.
// Matches the 6-step context-isolated pipeline in skill/qa/SKILL.md.
// Last step uses OutputArtifact="" so writeArtifacts writes test-plan + quality-report.
func QADescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "qa",
		Name:        "QA Engineer",
		Description: "Trigger: qa, test, acceptance criteria, edge case, coverage, quality",
		Skills:      []string{"platform-context", "artifact-envelope"},
		Workflow: []WorkflowStep{
			{
				ID:             "load-requirements",
				Description:    "Read spec/AC artifacts: requirements-spec, ux-brief, architectural-decision, system-design.",
				SkillRefs:      []string{},
				InputRefs:      []string{},
				OutputArtifact: "qa/ac-list",
			},
			{
				ID:             "ac-validation",
				Description:    "Validate acceptance criteria completeness, testability, clarity, and measurability.",
				SkillRefs:      []string{"acceptance-criteria"},
				InputRefs:      []string{"qa/ac-list"},
				OutputArtifact: "qa/ac-gaps",
			},
			{
				ID:             "edge-case-analysis",
				Description:    "Enumerate boundary values, error paths, concurrent scenarios, and data edge cases.",
				SkillRefs:      []string{"edge-case-analysis"},
				InputRefs:      []string{"qa/ac-list"},
				OutputArtifact: "qa/edge-cases",
			},
			{
				ID:             "test-strategy",
				Description:    "Define unit/integration/E2E strategy, coverage targets, and test data approach.",
				SkillRefs:      []string{"test-strategy"},
				InputRefs:      []string{"qa/edge-cases"},
				OutputArtifact: "qa/test-strategy",
			},
			{
				ID:             "test-case-generation",
				Description:    "Author concrete test cases in Given/When/Then format covering happy path and edge cases.",
				SkillRefs:      []string{"test-strategy"},
				InputRefs:      []string{"qa/test-strategy", "qa/edge-cases"},
				OutputArtifact: "qa/test-cases",
			},
			{
				ID:          "quality-report",
				Description: "Summarize AC gaps, missing test infrastructure, and open questions for Developer.",
				SkillRefs:   []string{},
				InputRefs:   []string{"qa/test-cases", "qa/ac-gaps"},
				// OutputArtifact="" → last-step fallback writes test-plan + quality-report.
				OutputArtifact: "",
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
// First step is platform-analysis. Last step uses OutputArtifact="" so writeArtifacts
// writes all three finals: threat-model + security-findings + hardening-checklist.
func SecurityDescriptor() SpecialistDescriptor {
	return SpecialistDescriptor{
		ID:          "security",
		Name:        "Security Engineer",
		Description: "Trigger: security, threat model, owasp, vulnerability, hardening, attack surface",
		Skills:      []string{"platform-context", "artifact-envelope"},
		Workflow: []WorkflowStep{
			{
				ID:                "platform-analysis",
				Description:       "Establish platform context for security assessment. Extract stack, trust boundaries, and entry points.",
				SkillRefs:         []string{},
				InputRefs:         []string{},
				OutputArtifact:    "platform-summary",
				SkipIfInitialized: true,
			},
			{
				ID:             "threat-modeling",
				Description:    "STRIDE-style threat model from available platform context.",
				SkillRefs:      []string{"threat-modeling"},
				InputRefs:      []string{"platform-summary"},
				OutputArtifact: "security/stride-threats",
			},
			{
				ID:             "attack-surface",
				Description:    "Enumerate all entry points and map trust boundaries.",
				SkillRefs:      []string{"threat-modeling"},
				InputRefs:      []string{"security/stride-threats"},
				OutputArtifact: "security/attack-surface",
			},
			{
				ID:             "owasp-analysis",
				Description:    "Review the feature against the OWASP Top 10 2021 categories relevant to the stack.",
				SkillRefs:      []string{"owasp-review"},
				InputRefs:      []string{"security/attack-surface"},
				OutputArtifact: "security/owasp-findings",
			},
			{
				ID:          "hardening-checklist",
				Description: "Produce an ordered hardening action list for the Developer specialist.",
				SkillRefs:   []string{},
				InputRefs:   []string{"security/stride-threats", "security/owasp-findings"},
				// OutputArtifact="" → last-step fallback writes threat-model + security-findings + hardening-checklist.
				OutputArtifact: "",
			},
		},
		Artifacts: ArtifactContract{
			Reads:  []string{}, // no required predecessor — can run at any lifecycle stage
			Writes: []string{"threat-model", "security-findings", "hardening-checklist"},
		},
	}
}
