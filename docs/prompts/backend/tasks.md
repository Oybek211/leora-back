# Planner Tasks Backend Prompt

## Mission
Design the task domain so it drives planning, automation, and Finance linkage without leaking UI details. Tasks anchor user intent, progress, and context while keeping sync-friendly metadata (`syncStatus`, `showStatus`, `idempotencyKey`).

## Entities & Data Relationships
- `tasks` stores title, status, priority, due/start dates, energy/context hints, optional links to `goals`, `habits`, and finance helpers (budget/debt/account). `lastFocusSessionId` ties into `focus_sessions` for reporting.
- Related tables: `task_checklist_items` (ordered subtasks) and `task_dependencies` (graph of prerequisites). Each entry tracks `taskId` and enforces `dependsOnTaskId` constraints within the same tenant.
- Tasks feed `goal_tasks`, `habit_tasks`, and `finance` (via `financeLink` or direct `budgetId`/`debtId`) so updates must cascade to `goal` progress, habit streaks, and budget tallies.

## API Surface & Responsibilities
- Standard CRUD: GET `/tasks`, `/tasks/:id`; POST `/tasks` (including nested `checklist`), PATCH `/tasks/:id`, DELETE soft delete (`showStatus = deleted`). Ensure filtering (status, priority, dueDate range, search, sort) operates within tenant scope and obeys pagination meta structure.
- Action endpoints: `/tasks/:id/complete`, `/tasks/:id/reopen`, and `/tasks/:id/checklist/:itemId` to mutate progress counters and adjust `focusTotalMinutes` when applicable.
- All writes update `syncStatus`, moderate idempotency via `idempotencyKey`, and trigger goal/habit recalculations. Soft deletes must mark associated checklist entries as `deleted` to avoid orphan conflicts.

## Real-Time & WebSocket Notes
- Notify clients through `entity:updated` (and `entity:created`/`entity:deleted` if applicable) when tasks change so dashboards and widgets stay consistent. Long-polling is not sufficient for reminders tied to due dates.
- Emit `reminder` events ahead of deadlines if server-side scheduler identifies pending tasks and `notifications` settings allow alerts.

## Operational Considerations
- Respect rate limits for write endpoints (50 req/min) and read endpoints (100 req/min).
- Support `sync/push` + `sync/pull` interaction; conflicts in tasks should be resolvable via `/sync/resolve-conflict` with options `use_local`, `use_server`, `merge`.
