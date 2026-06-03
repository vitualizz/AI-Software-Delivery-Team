# AI Software Delivery Team (ASDT)

ASDT is an artifact-first AI delivery system: every agent action produces a durable, human-readable YAML artifact under `.asdt/`, making the entire delivery process auditable, reproducible, and independent of any specific AI runtime.

## Overview

ASDT ships as two independent artifacts that share one boundary (`.asdt/`):

1. **`skill/`** — the runtime-agnostic `/asdt` router. Pure prompt + file I/O. Runs identically in Claude Code and OpenCode.
2. **Go binary** — an optional TUI that reads `.asdt/` for visualization. Built with hexagonal architecture.

## Quick Start

```sh
# TODO: add quick start after T-2-4
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for the full design.

## License

MIT
