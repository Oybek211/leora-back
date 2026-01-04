# Planner Goals Backend Prompt

## Mission
Control the lifecycle of goals so they act as progress anchors across planner, habits, finance, and achievements. Maintain progress metrics, milestones, linked budgets/debts, and the instructions needed for other modules to consume goal state.

## Entities & Data Relationships
- `goals` keeps descriptive metadata (title, type, metric, direction, dates) plus denormalized pointers (`linkedBudgetId`, `linkedDebtId`, optional `financeMode`). Each `goal` belongs to one `user`.
- Related tables: `goal_milestones`, `goal_check_ins`, `goal_tasks`, `goal_habits`, and `goal_finance_contributions` enforce many-to-many relationships and capture progress snapshots. `goal_stats` (embedded JSON) surfaces aggregated percentages and focus minutes.
- Progress updates propagate to `habits` (via `goal_habits`), `tasks`, `transactions`, and `insights` (AI or manual) to keep the entire dashboard cohesive.

## API Surface & Responsibilities
- CRUD: GET `/goals`, `/goals/:id`, POST `/goals`, PATCH `/goals/:id`, DELETE (soft). Validation must ensure target metrics make sense for `metricType`, and finance goals align with `unit`/`currency` fields.
- Actions: `/goals/:id/check-in` for `value`, `link`, `note`; `/goals/:id/complete` and `/goals/:i/reactivate` toggle `status`. Check-ins also create `goal_check_ins` rows and update `progressPercent`.
- When building new goals with `milestones`, insert ordered `goal_milestones` and compute percent boundaries; tie financial goals to budgets/debts when `financeMode` or `linkedBudgetId` is provided.

## Real-Time & WebSocket Notes
- Broadcast `entity:updated` events when progress, check-ins, milestones, or status changes to keep dashboards, widgets, and insights aware of new thresholds.
- Consider emitting targeted `insight:new` events when a milestone is reached or a check-in pushes `progressPercent` above configured thresholds.

## Operational Considerations
- Aggregate goal progress across tasks, habits, and finance contributions before responding to GET requests to avoid stale denormalized values.
- Keep `syncStatus` aligned with the front-end to avoid duplicates; if goal data conflicts during sync, rely on `/sync/resolve-conflict` with merge strategies that respect `updatedAt`.
