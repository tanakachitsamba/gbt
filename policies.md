# Agent Orchestration Policies

This document describes how the multi-agent orchestration layer is structured, the purpose of each role, and why the workflow is designed the way it is. The goal is to provide human-readable guidance for future contributors rather than embedding executable Go code inside the policy description.

## Overview

The orchestration layer coordinates a small team of specialised agents that collaborate on every user request. Each agent performs a narrow task, passes its output to the next role, and receives constructive feedback that encourages iterative refinement. The result is a conversation loop that balances rapid ideation with systematic validation before a response is returned to the user.

The core principles behind the design are:

- **Separation of concerns** – each agent has a clearly defined objective that limits scope creep and simplifies prompt engineering.
- **Recursive feedback** – the critic role reviews intermediate results and requests revisions when necessary, preventing low-quality answers from propagating.
- **Memory awareness** – conversation history and previous task outputs can be summarised and reintroduced when relevant, allowing the system to handle multi-step engagements without overwhelming any single agent.

## Agent roles

| Role | Responsibility | Typical prompt goals |
| --- | --- | --- |
| **Summariser** | Condenses the current instruction and any surfaced memory into a concise working brief. | Highlight the task, list salient constraints, note prior context the executor will need. |
| **Criticiser** | Reviews drafts from creative or planning agents to spot logical gaps, violations of policy, or poor reasoning. | Provide a verdict (pass/fail), explain flaws, and suggest concrete fixes. |
| **Enquirer** | Collects missing requirements from the user when the summariser indicates gaps. | Ask clarifying questions, capture answers, and update the working brief. |
| **Prioritiser** | Orders subtasks or blockers so downstream agents attack them efficiently. | Rank steps, flag dependencies, and mark items requiring human confirmation. |
| **Planner** | Produces a step-by-step strategy for the executor to follow. | Outline milestones, assign ownership (agent vs. user), and set acceptance criteria. |
| **Lister** | Enumerates artefacts, resources, or references relevant to the plan. | Provide checklists, reference URLs, or file paths for quick lookup. |
| **Decider** | Chooses between competing options produced earlier in the chain. | Compare alternatives, document trade-offs, and record the final selection. |
| **Executor** | Delivers the final answer or artefact to the user once upstream reviews pass. | Incorporate feedback, verify requirements, and format the response for delivery. |

Although the table shows a linear progression, the orchestrator can loop back to earlier roles when the criticiser requests revisions or when the executor flags missing information.

## Orchestration flow

1. **Initial briefing** – The summariser consumes the latest instruction, merges in any stored memories, and generates a concise brief. If key details are missing, the enquirer branch is triggered to gather clarifications before proceeding.
2. **Iterative planning** – The prioritiser, planner, and lister collaborate to produce an actionable roadmap. Their outputs give structure to the task, ensuring the executor receives both the why and the how of the requested work.
3. **Draft creation** – The executor, sometimes assisted by specialised domain agents, crafts a draft response aligned with the plan.
4. **Quality review** – The criticiser inspects the draft. When issues are found, the feedback is appended to the conversation thread and routed back to the appropriate agent (often the executor or planner) for revision. This feedback loop continues until the criticiser is satisfied or the system escalates for human intervention.
5. **Decision gate** – For tasks with multiple possible solutions, the decider evaluates the available options, documents the rationale, and confirms the final direction.
6. **Delivery and memory update** – The approved response is returned to the user. Relevant insights are summarised for long-term memory so future runs can leverage prior knowledge.

## Design rationale

- **Human-like collaboration**: Modelling the workflow after multidisciplinary teams creates natural checkpoints that mirror real-world review processes.
- **Controlled recursion**: By limiting recursive loops to specific phases (typically between executor and criticiser) the system balances thoroughness with predictable latency.
- **Extensibility**: New roles can be introduced with minimal disruption by defining their responsibilities, prompts, and insertion point in the pipeline.
- **Traceability**: Each agent’s output is stored in a structured conversation thread, enabling easy auditing of how a conclusion was reached.

## Implementation notes

Executable examples demonstrating how to wire these roles together now live in `pkg/policies/example.go`. That file defines the supporting types (`Input`, `Agent`, `ConversationThread`, and the `getResponse` helper) and showcases a minimal orchestration loop suitable for experimentation or testing. Contributors should reference that code when implementing or modifying the production orchestrator, while keeping this document focused on conceptual guidance.

## Next steps

- Formalise prompts for each role once product requirements stabilise.
- Integrate persistent memory storage so the summariser can recall prior sessions without manual intervention.
- Expand the criticiser’s capabilities with automated policy and safety checks before hand-off to the executor.

By treating these policies as living documentation, we maintain clarity for both engineers and prompt designers, ensuring the agent collective remains reliable, debuggable, and easy to extend.
