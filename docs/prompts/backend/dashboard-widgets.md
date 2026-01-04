# Widgets / Home Dashboard Backend Prompt

## Mission
Supply the curated data feed for home/dashboard widgets (tasks, insights, finance summaries, reminders) so the UI can render a cohesive overview without requerying each module separately.

## Entities & Data Relationships
- Aggregations pull from `tasks`, `habits`, `budgets`, `transactions`, `insights`, `focus_sessions`, and `achievements`. Each widget request is scoped by `userId` and, where relevant, filtered by `syncStatus`, `showStatus`, and `status` enums.
- Derived models may store snapshots of counts (e.g., active tasks, overdue debts, focus minutes) to reduce repeated work.

## API Surface & Responsibilities
- Provide a dashboard endpoint (e.g., GET `/dashboard/widgets`) that composes:
  - Task summary (TODAY/INBOX counts, next due)
  - Finance highlights (net balance, over/under budgets, upcoming debt payments)
  - Habit streaks and focus session summaries
  - Active insights and notification triggers
- Ensure each widget obeys caching headers and returns meta (lastUpdated) to help clients manage refresh cadence.

## Real-Time & WebSocket Notes
- When underlying data changes (task due status, budget overspend, new insights), emit relevant WebSocket events (`entity:updated`, `insight:new`, `reminder`) so widgets can update reactively.
- Allow widgets to subscribe to entity groups (`subscribe` event) rather than all updates to reduce noise.

## Operational Considerations
- Avoid heavy joins by either precomputing snapshot rows or deferring expensive calculations to background jobs that populate the dashboard cache.
- Respect feature gating (premium plans) when surfacing widgets—for example, hide AI-generated insights if `premium` is not active.
