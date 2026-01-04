# Planner Focus Sessions Backend Prompt

## Mission
Track Pomodoro-style sessions tied to tasks/goals so the backend can report focus time, interruptions, and completion rates without knowing UI specifics.

## Entities & Data Relationships
- `focus_sessions` links to `users`, optional `tasks`, and `goals`. Fields include planned vs actual minutes, status (`in_progress`, `completed`, etc.), timestamps, interruption counts, and notes.
- Focus sessions update `tasks.focusTotalMinutes` when completed, and potentially contribute to `insights` or achievements (e.g., `focus_minutes_total`).

## API Surface & Responsibilities
- Endpoints: GET `/focus-sessions`, GET `/focus-sessions/:id`, POST `/focus-sessions/start`, PATCH `/focus-sessions/:id/pause`, PATCH `/focus-sessions/:id/resume`, POST `/focus-sessions/:id/complete`, POST `/focus-sessions/:id/cancel`, POST `/focus-sessions/:id/interrupt`, GET `/focus-sessions/stats`.
- Starts create new rows with `status = in_progress`, `plannedMinutes`, and `startedAt`. Pause/resume mutate `status` and track `interruptionsCount`. Complete writes `actualMinutes`, updates notes, and sets `endedAt`.
- Stats endpoint aggregates today/this week/this month totals, drawing from completed sessions per user.

## Real-Time & WebSocket Notes
- When a session starts/ends, emit `entity:updated` so UI widgets can display live timers or summary cards.
- Consider sending `reminder` events for upcoming `focus_sessions` or `entity:updated` notifications when interruptions occur.

## Operational Considerations
- Ensure no overlapping in-progress sessions per user; enforce constraints within transactions.
- Respect sync semantics: even pending/in-progress sessions created offline should carry `syncStatus` for later reconciliation.
