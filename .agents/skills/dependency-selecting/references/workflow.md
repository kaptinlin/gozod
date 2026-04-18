# Workflow & State

## `github.com/agentable/go-schedule` — In-Process Scheduling

- Minimal scheduler with two verbs: `Every` and `Cron`
- Supports hooks, persistence, distributed locking, and graceful shutdown
- Distinct from workflow engines and task queues

**When to use:** Periodic jobs (cleanup tasks, sync jobs, timed dispatch) where you need app-level scheduling but not full workflow orchestration.

## `github.com/agentable/go-cron` — Cron Primitive

- Focused cron expression dependency for parsing and schedule primitives
- Use when you need cron semantics without a full scheduler runtime

**When to use:** Shared cron expression handling, schedule calculation, or lower-level scheduling primitives.

## `github.com/agentable/aster` — Workflow/Pipeline Engine

- DAG scheduling for complex pipelines
- State persistence, resumable execution
- Human-in-the-loop capabilities

**When to use:** Multi-step data pipelines, approval workflows, ETL processes, CI/CD-like orchestration, any DAG-based execution.

## `github.com/agentable/go-flow` — Declarative Flow Engine

- YAML-driven flow execution with default DAG parallelism
- Resume from suspension points without re-running completed steps
- Typed function registration and expression-based wiring

**When to use:** Declarative, resumable flow definitions where business logic should live in a YAML flow rather than imperative orchestration code.

## `github.com/agentable/go-fsm` — Finite State Machine

- Lightweight, generic FSM for Go 1.26+
- Type-safe parameters
- Zero-alloc hot paths
- Two declaration styles

**When to use:** Order status tracking, approval flows, connection state management, game state, any entity lifecycle.

```go
// Example: order lifecycle
fsm := gofsm.New[OrderState, OrderEvent](
    gofsm.Transition(Pending, PaymentReceived, Paid),
    gofsm.Transition(Paid, Shipped, InTransit),
    gofsm.Transition(InTransit, Delivered, Completed),
)
```

## `github.com/dtm-labs/dtm` — Distributed Transactions

- SAGA, TCC, XA, Workflow, Outbox/2-phase messaging patterns
- Multi-language SDKs (Go, Java, PHP, C#, Python, Node.js)
- Storage: MySQL, Redis, MongoDB, PostgreSQL, BoltDB
- Integrations: go-zero, Kratos, Polaris
- High availability, horizontal scaling

**When to use:** Cross-service data consistency, compensation logic (if step fails, roll back), cache synchronization, inventory management, multi-service order systems.

## `github.com/agentable/go-process` — Process Lifecycle Management

- Cross-platform child process lifecycle control with suspend/resume and process group cleanup
- Buffered, streaming, writer, and combined output modes
- Graceful termination and context-aware cancellation

**When to use:** Managing external commands or long-running subprocesses as first-class runtime objects.

## Decision Tree

```
Need workflow/state management?
├── Periodic jobs inside one app?
│   └── agentable/go-schedule
├── Cron parsing / schedule primitive only?
│   └── agentable/go-cron
├── Single entity state transitions?
│   └── agentable/go-fsm
├── Declarative resumable flow in YAML?
│   └── agentable/go-flow
├── Multi-step pipeline / DAG execution?
│   └── agentable/aster
├── Process lifecycle / subprocess orchestration?
│   └── agentable/go-process
└── Cross-service transactional consistency?
    └── dtm-labs/dtm
```

## Do NOT Use

| Library | Reason |
|---------|--------|
| `github.com/robfig/cron/v3` | Standardize scheduler stack on our own scheduler/cron libs |
