# AI Development Log

This document records important AI-assisted technical discussions, design alternatives, and implementation decisions made during project development.

## 1. Budget Management: Reserved vs. Immediate Deduction

**Approaches discussed:**

- Without reserved budget: Deduct the step cost before execution.
- With reserved budget: Reserve the estimated job cost before execution, then convert or release the reservation as steps execute.

**Decision:** Use immediate budget deduction before each step attempt.

**Reason:**
- Simpler budget accounting.
- Failed attempts are also charged.
- No additional `Reserved` field is required.
- A refund is needed if a charge succeeds but the corresponding dispatch fails before execution.

## 2. Slice of Values vs. Slice of Pointers

Compared:

```go
[]Slot
```

and:

```go
[]*Slot
```

A slice of values stores the structs directly, while a slice of pointers stores references to structs.

The choice depends on ownership, mutation requirements, and memory considerations. Neither is universally better.

## 3. Mutex Critical Section

Compared checking invalid refund amounts before and after acquiring a mutex.

**Decision:** Check `amount <= 0` before acquiring the mutex.

**Reason:** Invalid refund requests can return immediately without locking. This keeps the critical section shorter.

Both approaches are correct.

## 4. Budget Reservation Example

Consider a tenant with a budget of 100 and a job with three steps costing 10, 20, and 30.

The reserved-budget approach could initially reserve 60.

| Event | Spent | Reserved | Available |
|---|---:|---:|---:|
| Job submitted | 0 | 60 | 40 |
| Step 1 succeeds | 10 | 50 | 40 |
| Step 2 succeeds | 30 | 30 | 40 |
| Step 3 fails | 30 | 0 | 70 |

Unused reserved budget is released when the job finishes.

This was considered as an alternative to immediate deduction.

## 5. Round-Robin Scheduling Fairness

Consider two tenants:

| Tenant | Max concurrency | Running | Queued |
|---|---:|---:|---:|
| A | 3 | 0 | 5 |
| B | 1 | 0 | 5 |

An approach that fills all available slots for one tenant before visiting the next can dispatch three jobs for A before B gets an opportunity.

Example:

```go
for _, t := range tenants {
    for {
        if !s.tenantStore.TryStartJob(t.ID) {
            break
        }

        j := s.jobStore.ClaimNextQueued(t.ID)

        if j == nil {
            s.tenantStore.FinishJob(t.ID)
            break
        }

        s.jobs <- j
    }
}
```

**Problem:** This does not provide per-dispatch round-robin fairness.

**Decision:** Use a `lastTenantID` cursor and dispatch one job per tenant turn.

**Reason:** Each eligible tenant gets an opportunity before the scheduler returns to a tenant that has already received a dispatch.

## 6. General Development Notes

- Prefer simple implementations over unnecessary abstractions.
- Protect shared state when accessed concurrently.
- Document important trade-offs and rejected approaches.
- Keep implementation decisions consistent with assignment requirements.
