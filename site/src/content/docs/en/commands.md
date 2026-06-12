---
title: Commands
description: Complete reference for all ASDT CLI commands and slash commands.
order: 4
locale: en
---

# Commands

## CLI Commands

### `asdt init`

Initialize ASDT in the current project. Creates `.asdt/config.yaml` and sets up the memory provider.

```bash
asdt init
```

### `asdt install`

Install or update the ASDT specialist skills for Claude Code.

```bash
asdt install
asdt install --all        # install all available specialists
asdt install --force      # reinstall even if up to date
```

## Slash Commands (Claude Code / OpenCode)

### `/asdt`

Pipeline routing suggestion. Describe what you want to build — ASDT analyzes the request, recommends which specialists to involve and in what order, and waits for your confirmation before listing the commands to run.

```
/asdt Add email verification to the signup flow
```

### `/asdt-pm`

Runs the Product Manager specialist only.

```
/asdt-pm Redesign the notification settings page
```

### `/asdt-architect`

Runs the Architect specialist only.

```
/asdt-architect Design the real-time sync architecture
```

### `/asdt-developer`

Runs the Developer specialist only.

```
/asdt-developer Implement the sidebar component
```

### `/asdt-qa`

Runs the QA specialist only.

```
/asdt-qa Review the checkout flow for edge cases
```

### `/asdt-security`

Runs the Security specialist only.

```
/asdt-security Review the OAuth integration
```

### `/asdt-ux-ui`

Runs the UX/UI specialist only.

```
/asdt-ux-ui Design the onboarding flow
```
