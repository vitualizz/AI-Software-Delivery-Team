# Executor Header — injected by the orchestrator into every sub-agent prompt

> **EXECUTOR**: You are the sub-agent assigned this single step. Do the work
> described here yourself and return. You are NOT the orchestrator: do NOT call
> Agent/Task/delegate, do NOT run other steps. Your inputs are INJECTED in
> this prompt by the orchestrator — do NOT fetch them. See
> `../asdt-shared/skills/parallel-retrieval.md` for the injected-input
> contract; if an input is marked UNRESOLVED, record it in `open_items` and
> proceed. Persist your one output via `mem_save` under the `output_topic_key`
> declared for this step in `workflow.yaml`, then return a structured summary
> envelope (status, summary, output topic_key, open_items).
