---
title: User Flows
description: Common patterns and end-to-end workflows with ASDT.
order: 4
locale: en
---

# User Flows

## Getting a pipeline suggestion

Use `/asdt` when you're not sure which specialists a feature needs:

```
/asdt Add passwordless login with magic links
```

ASDT analyzes the request, assesses complexity and risk surface, and presents a routing plan — for example:

```
Recommended specialists:
  PM — define scope and acceptance criteria
  Architect — design token flow and API contracts
  Developer — implement the magic link handler
  Security — review the auth mechanism

Suggested order:
  /asdt-pm → /asdt-architect → /asdt-developer → /asdt-security

Proceed with this plan? (yes / modify / no)
```

Confirm the plan. ASDT then gives you the exact commands to run, with tailored workflow steps for each specialist. **You run them — ASDT does not execute specialists automatically.**

## Running specialists directly

When you already know what you need, skip `/asdt` and invoke the specialist directly:

```
/asdt-architect Design the rate-limiting strategy for the API
/asdt-qa Review the checkout flow for edge cases
/asdt-security Audit the OAuth integration
```

Each specialist runs its full workflow (explore → spec → design → implement, depending on complexity) and saves artifacts to Engram.

## Picking up mid-pipeline

If you ran some specialists and want to continue later, just invoke the next one. It reads prior artifacts from Engram automatically — even from a previous session:

```
/asdt-developer Implement based on the Architect's ADR
```

The Developer reads whatever the Architect produced via Engram. You don't pass context manually.

## Memory and Engram

ASDT uses [Engram](https://github.com/vitualizz/AI-Software-Delivery-Team) as its memory layer. Engram is **required** for the pipeline to function — it is how artifacts persist between specialists and across sessions.

More memory providers are planned. Today, Engram is the only supported option.

## Supported AI assistants

All slash commands work in both **Claude Code** and **OpenCode**. The syntax is identical in both tools.
