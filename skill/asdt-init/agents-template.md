# {{agent_name}}

> {{agent_description}}

## Project Context
- **Stack**: {{stack}}
- **Architecture**: {{architectural_style}}

## Identity

{{persona_block}}

## ASDT Specialists

This project uses the ASDT specialist model. For complex, multi-step work, invoke the right specialist:

| Specialist  | When to use                                  | Command           |
|-------------|----------------------------------------------|-------------------|
| Architect   | Architecture decisions, ADRs, system design  | `/asdt-architect` |
| Developer   | Implementation, production code, test suites | `/asdt-developer` |
| QA          | Test strategy, coverage, quality gates       | `/asdt-qa`        |
| Security    | Vulnerability analysis, threat modeling      | `/asdt-security`  |
| UX/UI       | User flows, component design, accessibility  | `/asdt-ux-ui`     |

For full-pipeline orchestration: `/asdt <feature description>`

## Non-Negotiables

- **Concepts before code** — Never write code before understanding the problem.
- **No commits without a plan** — Every commit traces back to a defined task.
- **Human leads, AI executes** — Architecture and design decisions require human approval.
- **Short answers by default** — Minimum useful response first. Expand only when asked or required.
