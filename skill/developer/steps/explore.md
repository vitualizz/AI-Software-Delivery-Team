# Explore — Developer Specialist

## Purpose
Understand the area of the codebase that will change before writing a single line.

## Inputs
- Request: the feature/change description from the user
- `platform-summary`: stack, naming conventions, key patterns

Note: This step has no prior step artifacts. It operates purely on the request and platform context.

## Context budget
Request text + platform-summary: max 1,000 tokens combined.

## Processing
1. Identify which existing files/modules are likely affected by this change.
2. Identify naming patterns and conventions from platform-summary relevant to this change.
3. List known risks or constraints (e.g. "this touches the auth layer which has rate limiting").
4. List open questions that need answering before speccing (e.g. "does this need migrations?").

Do NOT design the solution. Do NOT write code. Only explore and understand.

## Output
Produces: `developer/dev-exploration`

Schema:
```yaml
payload:
  files_to_understand: []     # existing files/modules relevant to this change
  patterns_to_follow: []      # naming/structural conventions from platform-summary
  risks: []                   # known constraints or risks
  open_questions: []          # questions that will be answered in the spec step
  open_items: []
```
