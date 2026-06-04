# Examples

This directory contains a complete snapshot of an ASDT run against a fictional
Go web API project. It shows exactly what files are created, what they contain,
and which command produced each one.

## What is here

```
examples/
└── sample-project/               # a fictional Go web API ("myapp")
    ├── go.mod
    ├── README.md
    └── .asdt/                    # everything ASDT wrote
        ├── config.yaml           # active change name
        ├── knowledge/
        │   └── platform.yaml     # produced by: /asdt knowledge
        └── artifacts/add-user-auth/
            ├── requirements-spec.yaml    # produced by: /asdt requirements "..."
            ├── implementation-plan.yaml  # produced by: /asdt develop
            └── pipeline-state.yaml      # updated automatically after each step
```

## Which command produced each file

### `.asdt/knowledge/platform.yaml`

```bash
/asdt knowledge
```

The knowledge detector scans the project root, identifies the tech stack (Go,
PostgreSQL), infers conventions (snake_case file names, REST API, table-driven
tests), and writes `platform.yaml`. This file is read by every subsequent agent
to tailor its output to the project's existing patterns.

### `.asdt/artifacts/add-user-auth/requirements-spec.yaml`

```bash
/asdt requirements "add user authentication with email and password login"
```

The requirements agent reads `platform.yaml`, interprets the free-text idea,
produces three user stories (register, login, logout), adds acceptance criteria
for each, defines scope in/out, lists NFRs (bcrypt, session expiry, rate
limiting), and surfaces one open question about account lockout tracking.

### `.asdt/artifacts/add-user-auth/implementation-plan.yaml`

```bash
/asdt develop
```

The developer agent reads the requirements spec and the platform knowledge.
For each user story it produces a list of files to create and modify, a
rationale, and concrete code snippets matching the project's conventions.
`input_refs` in the envelope header links this artifact back to the
requirements spec — full traceability.

### `.asdt/artifacts/add-user-auth/pipeline-state.yaml`

Updated automatically after each `/asdt requirements` and `/asdt develop` call.
Records the current FSM state (`plan`) and the full transition history with
timestamps. You can always see where you are in the delivery pipeline and when
each step happened.

### `.asdt/config.yaml`

Written by ASDT when a change is first referenced. Stores the active change
name so you can omit `--change` in subsequent commands.

---

## How to use ASDT on a real project

See the root [README.md](../README.md) for installation and quick-start
instructions. The snapshot in `examples/sample-project/.asdt/` is the direct
output of running these three commands in order:

```bash
/asdt knowledge
/asdt requirements "add user authentication with email and password login"
/asdt develop
```

Every file under `.asdt/` is human-readable, committable, and diff-able. You
can review the requirements spec in a PR before running the developer agent,
adjust acceptance criteria, re-run `/asdt requirements`, and the new spec will
be in place for the next `/asdt develop` call.
