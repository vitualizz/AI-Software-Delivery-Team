---
title: Knowledge Base & Memory
description: How ASDT persists artifacts across specialists and sessions — and how the knowledge base keeps your team in sync.
order: 7
locale: en
---

# Knowledge Base & Memory

## Why memory matters

Without persistent memory, every specialist starts from scratch. The Developer can't read the Architect's decision record. The QA specialist doesn't know what the Developer built. You'd have to manually copy context between every step — which defeats the purpose of a team.

ASDT solves this with a knowledge base. Every artifact a specialist produces is saved with a stable key. The next specialist retrieves it automatically by key. Context flows forward without manual intervention, even across sessions separated by days.

## How knowledge flows

Every specialist step produces one artifact. That artifact is saved to the knowledge base with two identifiers:

- A **human-readable title** — e.g. `"add-auth/developer/dev-implementation"`
- A **stable key** for machine retrieval — used by subsequent specialists to fetch exactly the artifact they need

When the next specialist runs, it queries the knowledge base by key. If the artifact exists, it proceeds normally. If it's missing, it notes the gap in `open_items` and continues with available context — no hard failures, no blocked pipelines.

This means run order is flexible. You can start with any specialist. You can pause between steps. The knowledge base holds the state.

## Memory providers

ASDT requires a **memory provider** — a persistent store that survives session boundaries and makes artifacts available across different specialist invocations.

A memory provider gives ASDT:
- **Cross-session persistence** — Monday's Architect output is available to Thursday's Developer
- **Project scoping** — artifacts from different projects don't mix
- **Key-based retrieval** — deterministic lookup without fuzzy matching

Today, **Engram** is the supported memory provider. More providers are planned.

### Engram — current implementation

[Engram](https://github.com/Gentleman-Programming/engram) is an MCP (Model Context Protocol) server that provides persistent, project-scoped memory for AI assistants.

Engram must be running before you invoke any specialist (or whichever memory provider you have configured). Artifacts saved to Engram survive the end of a Claude Code session — that is the property ASDT depends on for cross-session pipeline continuity.

#### Installation

Follow the [Engram setup guide](https://github.com/Gentleman-Programming/engram) to install and start the MCP server. Then add it to your AI assistant's MCP configuration.

#### Verifying it's running

```
/asdt-init
```

The init specialist checks for memory provider connectivity as part of setup and reports if the MCP server is unreachable.

## How artifacts are stored

Each artifact is saved with:

- **`title`** — human-readable, e.g. `"add-auth/developer/dev-implementation"`
- **`topic_key`** — machine-readable key for retrieval, e.g. `"add-auth/developer/dev-implementation"`
- **`type`** — `architecture`, `decision`, `bugfix`, etc.
- **`project`** — the project name, used to scope search results

When a specialist needs a prior artifact, it calls `mem_search` with the topic_key, then `mem_get_observation` to retrieve the full content. This is a one-step lookup — no fuzzy matching, no context scanning.

## Cross-session continuity

Picking up a pipeline after a session ends is the same as continuing it mid-session:

```
/asdt-developer Implement based on the Architect's ADR
```

The Developer searches the knowledge base for the Architect's artifact by key. If found, it proceeds normally. If not found, it notes the missing input and proceeds with available context.

Session boundaries don't matter. Monday's Architect output is as readable to Thursday's Developer as if they'd run back to back.

## Project scoping

Every `mem_save` call includes a `project` field. ASDT derives the project name from `.asdt/config.yaml` or the directory name. Artifacts are searchable within a project scope — running ASDT on two different projects doesn't mix their memory.

## What Engram is not

Engram is not a file system. Artifacts are documents, not code files. When the Developer produces code snippets as part of its implementation artifact, those snippets live inside the Engram document — they're specifications for what to write, not files on disk. The human (or the AI assistant in a later step) applies them to the actual codebase.

This separation is intentional. Code files have git history. Artifacts have the knowledge base. Both live where they belong.
