# Workflow & State

## `github.com/netresearch/go-cron` — Cron Scheduling

- Cron expression based periodic job scheduling
- Lightweight scheduler for recurring tasks
- Focused API for job registration and execution

**When to use:** Periodic jobs (cleanup tasks, sync jobs, timed dispatch) where you need cron-style scheduling but not full workflow orchestration.

## `github.com/agentable/aster` — Workflow/Pipeline Engine

- DAG scheduling for complex pipelines
- State persistence, resumable execution
- Human-in-the-loop capabilities

**When to use:** Multi-step data pipelines, approval workflows, ETL processes, CI/CD-like orchestration, any DAG-based execution.

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

## Decision Tree

```
Need workflow/state management?
├── Periodic cron jobs?
│   └── netresearch/go-cron
├── Single entity state transitions?
│   └── agentable/go-fsm
├── Multi-step pipeline / DAG execution?
│   └── agentable/aster
└── Cross-service transactional consistency?
    └── dtm-labs/dtm
```

## Do NOT Use

| Library | Reason |
|---------|--------|
| `github.com/robfig/cron/v3` | Standardize scheduler stack; use one direct cron dependency |
