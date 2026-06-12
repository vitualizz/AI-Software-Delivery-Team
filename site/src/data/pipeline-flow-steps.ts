import type { SpecialistId } from './artifact-graph'

export interface FlowStep {
  command: string
  specialistId: SpecialistId | 'asdt'
  descriptionEn: string
  descriptionEs: string
}

export interface PipelineFlow {
  id: string
  titleEn: string
  titleEs: string
  steps: FlowStep[]
}

export const pipelineFlows: PipelineFlow[] = [
  {
    id: 'full-feature',
    titleEn: 'Getting a pipeline suggestion',
    titleEs: 'Obtener una sugerencia de pipeline',
    steps: [
      {
        command: '/asdt Add passwordless login with magic links',
        specialistId: 'asdt',
        descriptionEn: 'Ask the orchestrator — ASDT analyzes the request and recommends which specialists to involve and in what order.',
        descriptionEs: 'Pedile al orquestador — ASDT analiza el pedido y recomienda qué especialistas involucrar y en qué orden.',
      },
      {
        command: '/asdt-pm',
        specialistId: 'pm',
        descriptionEn: 'PM defines scope, writes user stories with acceptance criteria, saves pm/backlog-entry to the knowledge base.',
        descriptionEs: 'PM define el alcance, escribe historias de usuario con criterios de aceptación y guarda pm/backlog-entry en la base de conocimientos.',
      },
      {
        command: '/asdt-architect',
        specialistId: 'architect',
        descriptionEn: 'Architect reads the backlog entry, designs the token flow and API contracts, saves architectural-decision + system-design.',
        descriptionEs: 'Architect lee el backlog entry, diseña el flujo de tokens y los contratos de API, guarda architectural-decision + system-design.',
      },
      {
        command: '/asdt-developer',
        specialistId: 'developer',
        descriptionEn: 'Developer reads the ADR and system design, implements the magic link handler, saves dev-implementation.',
        descriptionEs: 'Developer lee el ADR y el system design, implementa el magic link handler y guarda dev-implementation.',
      },
      {
        command: '/asdt-security',
        specialistId: 'security',
        descriptionEn: 'Security reviews the auth mechanism, runs STRIDE and OWASP analysis, saves security-findings + hardening-checklist.',
        descriptionEs: 'Security revisa el mecanismo de autenticación, ejecuta análisis STRIDE y OWASP, guarda security-findings + hardening-checklist.',
      },
    ],
  },
  {
    id: 'mid-pipeline',
    titleEn: 'Picking up mid-pipeline',
    titleEs: 'Continuar a mitad del pipeline',
    steps: [
      {
        command: "/asdt-developer Implement based on the Architect's ADR",
        specialistId: 'developer',
        descriptionEn: 'Developer reads prior artifacts from the knowledge base automatically — even from a previous session. No manual context passing.',
        descriptionEs: 'Developer lee los artefactos previos de la base de conocimientos automáticamente, incluso de sesiones anteriores. Sin pasar contexto manualmente.',
      },
      {
        command: '/asdt-qa',
        specialistId: 'qa',
        descriptionEn: 'QA loads dev-implementation and runs its full workflow: AC validation, edge-case analysis, and test case generation.',
        descriptionEs: 'QA carga dev-implementation y ejecuta su flujo completo: validación de ACs, análisis de edge cases y generación de tests.',
      },
    ],
  },
]
