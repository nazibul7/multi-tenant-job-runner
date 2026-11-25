# API Documentation

Base URL:

```text
http://localhost:8080
```

## 1. Create Tenant

**Endpoint**

```http
POST /tenants
```

Creates a tenant with a budget and maximum concurrency.

## 2. Submit Job

**Endpoint**

```http
POST /jobs
```

Submits a job for execution.

## 3. Get Job

**Endpoint**

```http
GET /jobs/{id}
```

Returns job information, including its steps and execution statuses.

## 4. Get Tenant Usage

**Endpoint**

```http
GET /tenants/{id}/usage
```

Returns the current tenant usage information as JSON.

## 5. Get Tenant Live View (SSE)

**Endpoint**

```http
GET /tenants/{id}/live
```

Provides a continuous stream of tenant usage updates using Server-Sent Events (SSE).

**Example:**

```bash
curl -N http://localhost:8080/tenants/A/live
```

**Example response:**

```text
data: {"tenant_id":"A","budget":1000,"spent":30,"remaining":970,"reserved":0,"queued":0,"running":0,"succeeded":3,"failed":0}

data: {"tenant_id":"A","budget":1000,"spent":30,"remaining":970,"reserved":0,"queued":0,"running":0,"succeeded":3,"failed":0}
```

Each `data:` message represents a tenant usage snapshot.

Unlike a normal HTTP request, the connection remains open and the server sends updates periodically.

**Response content type:**

```http
Content-Type: text/event-stream
```

## Job Statuses

| Status | Meaning |
|---|---|
| queued | Waiting for execution |
| running | Currently executing |
| succeeded | Execution completed successfully |
| failed | Execution failed |

## Step Statuses

| Status | Meaning |
|---|---|
| queued | Waiting to execute |
| running | Currently executing |
| succeeded | Execution completed successfully |
| failed | Execution failed |
| blocked | Not executed because of a failure or unmet dependency |

## Validation

Job validation includes:

- Empty job steps
- Empty step IDs
- Duplicate step IDs
- Nil steps
- Unknown dependencies
- Self-dependencies
- Circular dependencies

Invalid requests return HTTP 400.
