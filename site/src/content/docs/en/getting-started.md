---
title: Getting Started
description: Install and run your first ASDT pipeline in minutes.
order: 1
locale: en
---

# Getting Started

## Requirements

Before using ASDT, you need:

- **Claude Code** or **OpenCode** — installed and authenticated
- **A memory provider** — required for cross-session persistence (default: [Engram](https://github.com/vitualizz/AI-Software-Delivery-Team))
- A terminal (bash or zsh)

> **Building from source?** Go 1.22+ is required. Running the one-line installer downloads a pre-built binary — no compiler needed.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/vitualizz/AI-Software-Delivery-Team/main/install.sh | bash
```

Downloads the latest pre-built binary for your platform and installs it to `~/.local/bin/`.

## Initialization

Initialize ASDT in your project:

```bash
asdt init
```

Creates `.asdt/config.yaml` with sensible defaults.

## Your first pipeline

```
/asdt Add user authentication with email and password
```

ASDT analyzes the request and recommends a specialist sequence — for example: `/asdt-pm` → `/asdt-architect` → `/asdt-developer`. Confirm the plan, then run each command. Each specialist saves its output to the knowledge base so the next one picks up where the previous left off.

## Running individual specialists

```
/asdt-pm Add dark mode to the settings page
/asdt-architect Design the caching strategy
/asdt-developer Implement the user profile component
```
