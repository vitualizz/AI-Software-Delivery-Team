export type SpecialistId = 'researcher' | 'pm' | 'architect' | 'developer' | 'qa' | 'security' | 'ux-ui'

export interface ArtifactRef {
  key: string
  optional?: boolean
  consumedBy?: SpecialistId[]
  sentinel?: boolean
}

export interface SpecialistArtifacts {
  reads: ArtifactRef[]
  writes: ArtifactRef[]
}

export const artifactGraph: Record<SpecialistId, SpecialistArtifacts> = {
  researcher: {
    reads:  [{ key: 'Problem (raw)', sentinel: true }],
    writes: [
      { key: 'researcher/ideation' },
      { key: 'researcher/feasibility' },
      { key: 'researcher/discovery-brief', consumedBy: ['pm'] },
    ],
  },
  pm: {
    reads:  [{ key: 'Request (raw)', sentinel: true }],
    writes: [{ key: 'pm/backlog-entry', consumedBy: ['architect', 'developer', 'qa', 'ux-ui'] }],
  },
  architect: {
    reads:  [{ key: 'pm/backlog-entry' }],
    writes: [
      { key: 'architectural-decision', consumedBy: ['developer', 'qa'] },
      { key: 'system-design',          consumedBy: ['developer', 'qa', 'ux-ui'] },
    ],
  },
  developer: {
    reads:  [{ key: 'architectural-decision' }, { key: 'system-design' }],
    writes: [{ key: 'dev-implementation', consumedBy: ['qa'] }],
  },
  qa: {
    reads:  [{ key: 'dev-implementation' }, { key: 'architectural-decision', optional: true }],
    writes: [{ key: 'test-plan', consumedBy: ['developer'] }],
  },
  security: {
    reads:  [{ key: 'system-design', optional: true }, { key: 'dev-implementation', optional: true }],
    writes: [
      { key: 'security-findings',   consumedBy: ['developer', 'architect'] },
      { key: 'hardening-checklist', consumedBy: ['developer', 'architect'] },
    ],
  },
  'ux-ui': {
    reads:  [{ key: 'pm/backlog-entry' }, { key: 'system-design', optional: true }],
    writes: [
      { key: 'ux-brief',       consumedBy: ['developer'] },
      { key: 'component-spec', consumedBy: ['developer'] },
    ],
  },
}

export const PIPELINE_ORDER: SpecialistId[] = ['researcher', 'pm', 'architect', 'developer', 'qa', 'security', 'ux-ui']

export const SPECIALIST_COLOR: Record<SpecialistId, string> = {
  researcher: '--c-res',
  pm:        '--c-pm',
  architect: '--c-arch',
  developer: '--c-dev',
  qa:        '--c-qa',
  security:  '--c-sec',
  'ux-ui':   '--c-ux',
}

export const SPECIALIST_LABEL: Record<SpecialistId, string> = {
  researcher: 'Researcher',
  pm:        'PM',
  architect: 'Architect',
  developer: 'Developer',
  qa:        'QA',
  security:  'Security',
  'ux-ui':   'UX/UI',
}
