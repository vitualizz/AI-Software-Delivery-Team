export type ComplexityTier = 'trivial' | 'simple' | 'moderate' | 'complex'
export type RiskTier = 'none' | 'moderate' | 'high'

export interface Step {
  id: string
  purpose: string
  produces: string
}

export interface SpecialistConfig {
  color: string
  tierType: 'complexity' | 'risk-surface'
  tiers: Partial<Record<ComplexityTier | RiskTier, string[]>>
  special?: Partial<Record<ComplexityTier | RiskTier, 'not-called' | 'not-eligible' | 'not-auto-invoked'>>
  steps: Record<string, Step>
}

// Source of truth: skill/asdt-{id}/SKILL.md and steps/*.md files
export const specialistSteps: Record<string, SpecialistConfig> = {
  researcher: {
    color: '--c-res',
    tierType: 'complexity',
    tiers: {
      trivial:  ['context-recall', 'divergent-ideation', 'decision-preservation'],
      simple:   ['context-recall', 'divergent-ideation', 'feasibility-scan', 'discovery-brief', 'decision-preservation'],
      moderate: ['context-recall', 'divergent-ideation', 'feasibility-scan', 'discovery-brief', 'decision-preservation'],
      complex:  ['context-recall', 'divergent-ideation', 'feasibility-scan', 'discovery-brief', 'decision-preservation'],
    },
    steps: {
      'context-recall': {
        id: 'context-recall',
        purpose: 'Search organizational memory for prior discovery, related decisions, and known constraints before ideating',
        produces: 'context (inline)',
      },
      'divergent-ideation': {
        id: 'divergent-ideation',
        purpose: 'Frame the problem and generate divergent candidate directions — deliberately generative, never selective',
        produces: 'researcher/ideation',
      },
      'feasibility-scan': {
        id: 'feasibility-scan',
        purpose: 'Assess each idea with a green/yellow/red feasibility verdict, supporting evidence, and effort estimate',
        produces: 'researcher/feasibility',
      },
      'discovery-brief': {
        id: 'discovery-brief',
        purpose: "Converge to ONE recommended direction with rationale; won't-do candidates seed PM's out-of-scope list",
        produces: 'researcher/discovery-brief',
      },
      'decision-preservation': {
        id: 'decision-preservation',
        purpose: "Preserve the chosen direction as permanent organizational knowledge via the brief's summary field",
        produces: 'summary (inline)',
      },
    },
  },

  pm: {
    color: '--c-pm',
    tierType: 'complexity',
    tiers: {
      trivial:  ['feature-intake'],
      simple:   ['feature-intake', 'user-stories', 'backlog-entry'],
      moderate: ['feature-intake', 'user-stories', 'scope-analysis', 'backlog-entry'],
      complex:  ['feature-intake', 'user-stories', 'scope-analysis', 'prioritization', 'backlog-entry'],
    },
    steps: {
      'feature-intake': {
        id: 'feature-intake',
        purpose: 'Parse raw request into structured problem statement — extracts problem, goal, stakeholders, flags ambiguities',
        produces: 'pm/feature-intake',
      },
      'user-stories': {
        id: 'user-stories',
        purpose: 'Write user stories with MoSCoW priorities and 1–3 high-level acceptance criteria per story',
        produces: 'pm/user-stories',
      },
      'scope-analysis': {
        id: 'scope-analysis',
        purpose: 'Define explicit in/out of scope boundaries, integration points, and scope risk flags',
        produces: 'pm/scope-analysis',
      },
      'prioritization': {
        id: 'prioritization',
        purpose: 'Order stories by dependency and risk; move Won\'t stories to deferred with explicit reasons',
        produces: 'pm/prioritization',
      },
      'backlog-entry': {
        id: 'backlog-entry',
        purpose: 'Consolidate all PM artifacts into the final backlog entry with executive summary and ordered story list',
        produces: 'pm/backlog-entry',
      },
    },
  },

  architect: {
    color: '--c-arch',
    tierType: 'complexity',
    tiers: {
      trivial:  ['load-constraints'],
      simple:   [],
      moderate: ['load-constraints', 'evaluate-approaches', 'decision-record'],
      complex:  ['load-constraints', 'evaluate-approaches', 'decision-record', 'system-design', 'risk-analysis', 'technical-handoff'],
    },
    special: {
      simple: 'not-called',
    },
    steps: {
      'load-constraints': {
        id: 'load-constraints',
        purpose: 'Read platform context and classify constraints as HARD (non-negotiable), SOFT (preferences), or OPPORTUNITIES',
        produces: 'architect/constraints-analysis',
      },
      'evaluate-approaches': {
        id: 'evaluate-approaches',
        purpose: 'Compare 2–3 viable approaches; choose one with explicit rationale; document why alternatives were rejected',
        produces: 'architect/approaches',
      },
      'decision-record': {
        id: 'decision-record',
        purpose: 'Write the ADR: context, decision, alternatives, and consequences — including negatives. Positives-only = incomplete',
        produces: 'architect/adr',
      },
      'system-design': {
        id: 'system-design',
        purpose: 'Define data model, API surface (method/inputs/errors), service boundaries, and the happy-path sequence',
        produces: 'architect/system-design',
      },
      'risk-analysis': {
        id: 'risk-analysis',
        purpose: 'Identify top 3–5 risks (performance, security, reliability, coupling, migration) with concrete mitigations',
        produces: 'architect/risks',
      },
      'technical-handoff': {
        id: 'technical-handoff',
        purpose: 'Consolidate all architectural work into two final artifacts for Developer and QA',
        produces: 'architectural-decision + system-design',
      },
    },
  },

  developer: {
    color: '--c-dev',
    tierType: 'complexity',
    tiers: {
      trivial:  ['explore'],
      simple:   ['explore', 'spec', 'implement'],
      moderate: ['explore', 'spec', 'design', 'implement'],
      complex:  ['explore', 'spec', 'design', 'tasks', 'implement'],
    },
    steps: {
      'explore': {
        id: 'explore',
        purpose: 'Read affected files and modules; map naming patterns and constraints before designing anything',
        produces: 'developer/dev-exploration',
      },
      'spec': {
        id: 'spec',
        purpose: 'Define scope boundary, answer open questions, write acceptance criteria in Given/When/Then format',
        produces: 'developer/dev-spec',
      },
      'design': {
        id: 'design',
        purpose: 'Choose technical approach, define data model and API shape, list key implementation constraints',
        produces: 'developer/dev-design',
      },
      'tasks': {
        id: 'tasks',
        purpose: 'Break implementation into atomic tasks (<2h each) ordered by dependency, with S/M/L estimates',
        produces: 'developer/dev-tasks',
      },
      'implement': {
        id: 'implement',
        purpose: 'Write code for each task — plan-only mode (snippets) or writing mode (real files to declared targets)',
        produces: 'developer/dev-implementation',
      },
    },
  },

  qa: {
    color: '--c-qa',
    tierType: 'complexity',
    tiers: {
      trivial:  [],
      simple:   ['load-requirements', 'ac-validation', 'test-case-generation', 'quality-report'],
      moderate: ['load-requirements', 'ac-validation', 'edge-case-analysis', 'test-strategy', 'test-case-generation', 'quality-report'],
      complex:  ['load-requirements', 'ac-validation', 'edge-case-analysis', 'test-strategy', 'test-case-generation', 'quality-report'],
    },
    special: {
      trivial: 'not-eligible',
    },
    steps: {
      'load-requirements': {
        id: 'load-requirements',
        purpose: 'Extract and normalize acceptance criteria from upstream artifacts into Given/When/Then format',
        produces: 'qa/ac-list',
      },
      'ac-validation': {
        id: 'ac-validation',
        purpose: 'Review each AC for atomicity, measurability, independence, completeness, unambiguity — rewrite failing ones',
        produces: 'qa/ac-gaps',
      },
      'edge-case-analysis': {
        id: 'edge-case-analysis',
        purpose: 'Discover edge cases via boundary values, equivalence partitioning, state transitions, concurrent access, permission boundaries',
        produces: 'qa/edge-cases',
      },
      'test-strategy': {
        id: 'test-strategy',
        purpose: 'Define test pyramid: what each level covers, what it does not, test data strategy, and flakiness tolerance',
        produces: 'qa/test-strategy',
      },
      'test-case-generation': {
        id: 'test-case-generation',
        purpose: 'Write structured test specs (Given/When/Then) for happy path, validated ACs, and critical/high edge cases',
        produces: 'qa/test-cases',
      },
      'quality-report': {
        id: 'quality-report',
        purpose: 'Verify AC coverage, compute percentage, render READY / READY WITH CAVEATS / BLOCKED verdict',
        produces: 'test-plan',
      },
    },
  },

  security: {
    color: '--c-sec',
    tierType: 'risk-surface',
    tiers: {
      none:     [],
      moderate: ['threat-modeling', 'hardening-checklist'],
      high:     ['threat-modeling', 'attack-surface', 'owasp-analysis', 'hardening-checklist'],
    },
    special: {
      none: 'not-auto-invoked',
    },
    steps: {
      'threat-modeling': {
        id: 'threat-modeling',
        purpose: 'Apply STRIDE: Spoofing, Tampering, Repudiation, Information Disclosure, DoS, Elevation of Privilege',
        produces: 'security/stride-threats',
      },
      'attack-surface': {
        id: 'attack-surface',
        purpose: 'Map entry points, trust boundaries, data flows — verify validation/sanitization/encoding at each step',
        produces: 'security/attack-surface',
      },
      'owasp-analysis': {
        id: 'owasp-analysis',
        purpose: 'Check all 10 OWASP Top 10 categories (A01–A10) as APPLICABLE/NOT APPLICABLE and MITIGATED/AT RISK',
        produces: 'security/owasp-findings',
      },
      'hardening-checklist': {
        id: 'hardening-checklist',
        purpose: 'Deduplicate findings, prioritize by severity, group by effort: quick wins / medium / significant',
        produces: 'security-findings + hardening-checklist',
      },
    },
  },

  'ux-ui': {
    color: '--c-ux',
    tierType: 'complexity',
    tiers: {
      trivial:  ['feature-brief'],
      simple:   ['feature-brief', 'information-architecture', 'user-flows', 'component-mapping', 'ux-handoff'],
      moderate: ['feature-brief', 'information-architecture', 'user-flows', 'component-mapping', 'ux-handoff'],
      complex:  ['feature-brief', 'information-architecture', 'user-flows', 'component-mapping', 'responsive-strategy', 'ux-handoff'],
    },
    steps: {
      'feature-brief': {
        id: 'feature-brief',
        purpose: 'Identify primary actor, define core problem (not solution), establish 3–5 observable success criteria',
        produces: 'ux-ui/feature-brief',
      },
      'information-architecture': {
        id: 'information-architecture',
        purpose: 'Organize content into sections, prioritize immediate vs. progressive disclosure, define navigation path',
        produces: 'ux-ui/ia',
      },
      'user-flows': {
        id: 'user-flows',
        purpose: 'Map happy path, error path, and 2–3 edge case flows as numbered steps from the actor\'s perspective',
        produces: 'ux-ui/flows',
      },
      'component-mapping': {
        id: 'component-mapping',
        purpose: 'Classify each UI state as reuse / extend / new — quality gate: >2:1 reuse ratio required',
        produces: 'ux-ui/components',
      },
      'responsive-strategy': {
        id: 'responsive-strategy',
        purpose: 'Mobile-first breakpoint specs for new/extended components; confirm 44×44px minimum touch targets',
        produces: 'ux-ui/responsive',
      },
      'ux-handoff': {
        id: 'ux-handoff',
        purpose: 'Consolidate all UX work into ux-brief (flows + IA) and component-spec (inventory + props + events)',
        produces: 'ux-brief + component-spec',
      },
    },
  },
}
