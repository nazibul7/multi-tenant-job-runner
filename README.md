# Multi-Tenant Job Runner

A Go-based job execution service that processes jobs for multiple tenants using a fixed-size worker pool.

The system supports per-tenant concurrency limits, round-robin scheduling, DAG-based job dependencies, retries, and budget tracking.

## Features

- **Multi-tenancy:** Multiple tenants can submit and execute jobs independently.
- **Worker pool:** A fixed number of workers execute jobs concurrently.
- **Per-tenant concurrency:** Each tenant has a configurable maximum number of running jobs.
- **Round-robin scheduling:** Jobs are dispatched fairly across tenants.
- **Job dependencies:** Steps can depend on other steps using `depends_on`.
- **Retry mechanism:** Failed steps are retried according to `max_attempts`.
- **Budget management:** Each execution attempt consumes the declared step cost.
- **Idempotency:** Supports duplicate request protection using an idempotency key, if enabled in the implementation.
- **Live usage:** Tenant usage can be monitored through an SSE endpoint.
- **In-memory storage:** Jobs and tenant information are stored in memory.

## Technology

- Go
- Standard library HTTP server
- Goroutines and channels
- Mutexes and WaitGroups
- In-memory storage

## Project Structure

```text
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   ├── job/
│   ├── service/
│   ├── store/
│   ├── tenant/
│   ├── worker/
│   └── sheduler/
├── README.md
├── DECISIONS.md
└── AI_LOG.md
```

## Requirements

- Go installed on your machine.
- Git.

## Run the Application

Clone the repository:

```bash
git clone <repository-url>
cd multi-tenant-job-runner
```

Start the server:

```bash
go run ./cmd/server
```

The server starts on:

```text
http://localhost:8080
```

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/tenants` | Create a tenant |
| POST | `/jobs` | Submit a job |
| GET | `/jobs/{id}` | Get job details |
| GET | `/tenants/{id}/usage` | Get tenant usage |
| GET | `/tenants/{id}/live` | Stream live tenant usage using SSE |

Job submission requires the `X-Tenant-ID` header.

## Tenant Configuration

Example:

```json
{
  "id": "tenant-a",
  "budget": 1000,
  "max_concurrency": 2
}
```

- `id`: Unique tenant identifier.
- `budget`: Maximum allowed spending.
- `max_concurrency`: Maximum number of jobs the tenant can run simultaneously.

## Job Configuration

Example:

```json
{
  "name": "example-job",
  "steps": [
    {
      "id": "step-a",
      "depends_on": [],
      "duration_ms": 1000,
      "cost": 10,
      "fail_times": 0,
      "max_attempts": 3
    },
    {
      "id": "step-b",
      "depends_on": ["step-a"],
      "duration_ms": 1000,
      "cost": 20,
      "fail_times": 0,
      "max_attempts": 3
    }
  ]
}
```

### Step fields

- `id`: Unique step identifier within the job.
- `depends_on`: Steps that must succeed before this step can execute.
- `duration_ms`: Simulated execution duration.
- `cost`: Cost charged for each execution attempt.
- `fail_times`: Number of initial attempts that should fail for simulation.
- `max_attempts`: Maximum number of attempts, including the first attempt.

## Scheduling

The scheduler uses round-robin tenant selection.

For each dispatch, it selects the next eligible tenant and claims one queued job.

The scheduler respects each tenant's `max_concurrency` setting.

A fixed-size worker pool limits the number of jobs executing simultaneously across the application.

The scheduler communicates with workers through a buffered channel.

## Job Execution

Steps are executed according to their dependencies.

- A step executes only after all its dependencies succeed.
- Failed steps are retried according to `max_attempts`.
- When retries are exhausted, the step is marked failed.
- Steps depending on a failed step are marked blocked.
- Independent steps can continue even when another dependency branch fails.
- A job is marked failed if any required step fails or becomes blocked.

## Budget Management

Each execution attempt consumes the step's declared cost.

This includes failed attempts.

The system checks whether the tenant has sufficient remaining budget before dispatching work.

Remaining budget:

```text
remaining = budget - spent
```

## Usage Monitoring

The usage endpoint provides tenant information, including:

- Total budget.
- Amount spent.
- Remaining budget.
- Queued jobs.
- Running jobs.
- Successful jobs.
- Failed jobs.

The live endpoint provides usage updates through Server-Sent Events (SSE).

## Error Handling

The service validates tenant and job input before accepting requests.

Expected error scenarios include:

- Invalid tenant configuration.
- Duplicate tenant identifier.
- Unknown tenant.
- Invalid job or DAG.
- Duplicate step identifiers.
- Missing dependencies.
- Circular dependencies.
- Insufficient tenant budget.
- Job not found.

Execution errors are reflected in step and job statuses.

The exact HTTP status codes and response formats are determined by the HTTP handlers.

## Limitations

- Storage is in memory.
- Restarting the application loses tenant and job data.
- There is no persistent job recovery after a process crash.
- Job execution is simulated using the configured duration.
- The system currently runs as a single application instance.

## Production Considerations

For production use, the system would need:

- Persistent storage for tenants and jobs.
- Recovery of unfinished jobs after restart.
- Distributed coordination if multiple application instances are used.
- Authentication and authorization.
- Structured logging and metrics.
- Configurable shutdown deadlines.
- Monitoring and alerting.

## Concurrency Safety

The application uses mutexes to protect shared in-memory state.

The race detector can be used to check for data races:

```bash
go run -race ./cmd/server
```
