# Core Infrastructure Backend Prompt

## Mission
Enforce consistent response formats, rate limiting, pagination, and error handling so every backend module presents a uniform experience.

## Entities & Data Relationships
- Not entity-specific, but every resource response needs the shared metadata (`success`, `data`, `meta`, `error`) described in the API docs. Include pagination meta (`page`, `limit`, `total`, `totalPages`, `hasMore`) for list endpoints.

## API Surface & Responsibilities
- All endpoints must request/return `meta` in the standard JSON envelope and handle pagination query params (`page`, `limit`, `sortBy`, `sortOrder`). Validate incoming pagination values and clamp to defined maximums (limit ≤ 100).
- Rate limits per category: Auth (10 req/min), Read (100 req/min), Write (50 req/min), Sync (20 req/min), AI/Insights (10 req/min). Implement per-user or per-device counters with early rejection (`429 RATE_LIMIT_EXCEEDED`).
- Error conventions: respond with structured errors (`code`, `message`, `details`) referencing codes like `VALIDATION_ERROR`, `RESOURCE_NOT_FOUND`, `IDEMPOTENCY_CONFLICT`, or provider-specific ones.

## Real-Time & WebSocket Notes
- When REST rate limiting rejects a request, optionally push a WebSocket event describing the new cooldown so UIs can disable related buttons.

## Operational Considerations
- Log all errors and rate-limit hits for monitoring; include `userId` and `endpoint` in logs for tracing.
- Keep `success` bools consistent; even on errors, pad the `data` field as `null` and fill `error` with contract details.
