---
title: Specialists
description: Learn what each ASDT specialist does and when to use them.
order: 2
locale: en
---

# Specialists

## Pipeline Advisor (`/asdt`)

The starting point when you're not sure which specialist(s) a feature needs. Describe what you want to build — ASDT analyzes the request and recommends which specialists to involve and in what order.

```
/asdt Add email verification to the signup flow
```

ASDT presents a routing plan and waits for your confirmation. Once you confirm, it tells you which commands to run. **You run each specialist — /asdt does not execute them for you.**

Each specialist saves its output to Engram memory. The next specialist reads that output automatically, so context flows from PM all the way to UX/UI — as long as you run them in the suggested order.

---

ASDT brings six AI specialists to your CLI. Each has a focused role and hands structured artifacts to the next step.

## Product Manager (`/asdt-pm`)

Turns raw feature requests into structured backlog entries with user stories, scope boundaries, and acceptance criteria.

**Use it when**: Requirements are ambiguous, you need user stories, or scope needs defining.

## Architect (`/asdt-architect`)

Makes technical decisions, produces Architecture Decision Records (ADRs), system design, and API contracts.

**Use it when**: A choice will shape service boundaries, data models, or scalability.

## Developer (`/asdt-developer`)

Transforms specs and designs into working code — implementation plans, production code, and test suites.

**Use it when**: The shape of the solution is settled and it's time to build it.

## QA Engineer (`/asdt-qa`)

Builds the safety net before code ships — test plans, acceptance criteria validation, edge case analysis, and quality reports.

**Use it when**: "It works on my machine" isn't good enough.

## Security (`/asdt-security`)

Hunts for the gaps an attacker would find first — threat models, OWASP reviews, and hardening checklists.

**Use it when**: Auth, data handling, or external integrations are on the table.

## UX/UI Designer (`/asdt-ux-ui`)

Shapes how people experience the product — user flows, information architecture, component specs, and accessibility strategy.

**Use it when**: A screen needs to be built or a user journey needs mapping.

## Memory and continuity

ASDT uses [Engram](https://github.com/vitualizz/AI-Software-Delivery-Team) as its memory provider. Every specialist reads from and writes to Engram, so:

- You can stop after any step and resume later — context is preserved across sessions.
- Specialists automatically pick up where the previous one left off.
- Work history is searchable and reusable across features.

Engram must be running before you invoke any specialist. See [Getting Started](/docs/getting-started) for setup.

## Supported AI assistants

ASDT slash commands work identically in both [Claude Code](https://claude.ai/code) and [OpenCode](https://opencode.ai). No configuration changes are needed to switch between them.
