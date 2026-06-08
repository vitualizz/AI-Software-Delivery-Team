# Examples

This directory contains a complete snapshot of an ASDT run against a fictional
Go web API project. It shows exactly what files are created, what they contain,
and which specialist produced each one.

## What is here

```
examples/
└── sample-project/                       # a fictional Go web API ("myapp")
    ├── go.mod
    ├── README.md
    └── .asdt/                            # everything ASDT wrote
        ├── config.yaml                   # active change name
        ├── knowledge/
        │   └── platform.yaml            # loaded by every specialist
        └── artifacts/add-user-auth/
            ├── ux-brief.yaml            # produced by: /asdt-ux-ui
            ├── implementation-plan.yaml  # produced by: /asdt-developer
            └── pipeline-state.yaml      # updated automatically after each step
```

## Which specialist produced each file

### `.asdt/knowledge/platform.yaml`

Read by every specialist before doing any work. Contains the detected tech
stack (Go, PostgreSQL), naming conventions, file structure, and design
fingerprint. ASDT reads `platform.yaml` to extend your system in a way that
feels consistent with what already exists.

### `.asdt/artifacts/add-user-auth/ux-brief.yaml`

```bash
/asdt-ux-ui "add user authentication"
```

The UX/UI specialist analyzed the platform context and the feature request,
produced an information architecture (routes, redirects), three user flows
(register, login, logout), a component mapping (reuse vs. new), and a
responsive strategy. The developer specialist loaded this artifact in Step 1
before generating any code.

### `.asdt/artifacts/add-user-auth/implementation-plan.yaml`

```bash
/asdt-developer "add user authentication"
```

The developer specialist loaded `ux-brief.yaml` and `platform.yaml`, estimated
complexity (M), and produced an ordered implementation plan with file-level
steps, rationale, code snippets, and test snippets — all matching the project's
Go conventions. `input_refs` in the envelope header links this artifact back to
`ux-brief.yaml` for full traceability.

### `.asdt/artifacts/add-user-auth/pipeline-state.yaml`

Updated automatically by each specialist after completing every step. The v2
format tracks each specialist independently — the UX/UI and Developer
specialists have separate step histories and artifact lists. You can see
exactly which steps completed, in what order, and when.

### `.asdt/config.yaml`

Written when a change is first referenced. Stores the active change name so
subsequent specialist commands know which artifact directory to read and write.

---

## The pipeline-state.yaml v2 format

The `schema_version: "2"` format tracks state per specialist rather than per
linear phase. Each specialist entry records:

- `current_step` — the last step that completed
- `steps_completed` — ordered list of step IDs and timestamps
- `artifacts_written` — artifact types this specialist wrote

Specialists that have not yet run are simply absent from the map. No specialist
waits for another — they each advance their own state independently.

Compare this to the v1 format, which recorded a single `current_state` and a
flat `transitions` list: that model only worked for a single linear FSM and
could not represent multiple specialists running in any order.

---

## How to use ASDT on a real project

See the root [README.md](../README.md) for installation and quick-start
instructions. The snapshot in `examples/sample-project/.asdt/` is the direct
output of running these commands:

```bash
/asdt-ux-ui "add user authentication"
/asdt-developer "add user authentication"
```

Or let the meta-orchestrator suggest the plan first:

```bash
/asdt "add user authentication with email and password"
# → confirms: /asdt-ux-ui → /asdt-developer
```

Every file under `.asdt/` is human-readable, committable, and diff-able. You
can review the UX brief in a PR before running the developer specialist, adjust
user flows, re-run `/asdt-ux-ui`, and the updated brief will be in place for
the next `/asdt-developer` run.
