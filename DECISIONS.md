# Design Decisions

## 1. In-Memory Storage

Used in-memory maps for tenant and job storage to keep the implementation simple.

Trade-off: Data is lost when the application restarts.

## 2. Fixed Worker Pool

Used a fixed number of workers instead of creating a goroutine for every job.

Reason: Limits concurrent execution and controls resource usage.

## 3. Buffered Channel

Used a buffered channel to transfer jobs from the scheduler to workers.

Reason: Allows the scheduler to dispatch jobs without waiting for a worker to receive every job immediately.

## 4. Round-Robin Scheduling

Used round-robin tenant selection.

Reason: Prevents tenants with many queued jobs from continuously taking scheduling opportunities.

## 5. Per-Tenant Concurrency

Tracked running jobs for each tenant.

Reason: Prevents one tenant from exceeding its configured concurrency limit.

## 6. DAG-Based Execution

Used step dependencies to determine execution order.

Reason: Allows jobs to express dependencies between individual steps.

## 7. Budget Charging

Charges the tenant for every execution attempt, including failed attempts.

Reason: Every attempt consumes execution resources.

## 8. Mutex-Based Synchronization

Used mutexes to protect shared in-memory data.

Reason: Multiple HTTP handlers, scheduler operations, and workers access shared state concurrently.
