# Context Extraction — Shared Skill

## Purpose
Extract only the fields relevant to a specific question or step from a large artifact.
Prevents full YAML dumps from bloating a step's context window.

## When to use
Use this skill in handoff steps (final step of each specialist) that need to summarize
multiple intermediate artifacts into a coherent final output. Do NOT use it in
intermediate steps — those should declare precise InputRefs instead.

## Protocol
Given: an artifact + a specific question or purpose

1. State the question/purpose in one sentence.
2. Identify the top-level keys of the artifact most relevant to that question.
3. For each relevant key: extract the value and summarize it to at most 200 tokens.
4. Discard all other keys entirely.
5. Present as a compact summary block.

## Context budget
Each extracted artifact summary MUST fit within 200 tokens.
If a list has more than 5 items, keep the 5 most important and note "… N more".

## Example
Question: "What are the implementation steps I need to test?"
Artifact: dev-tasks.yaml (contains 20+ fields)
Extracted: tasks[].id, tasks[].title only — discard estimates, rationale, files.
