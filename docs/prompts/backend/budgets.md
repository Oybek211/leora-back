# Finance Budgets Backend Prompt

## Mission
Enforce budgeting discipline by tracking limits, periods, category scopes, and contributions so overspend alerts and carryover logic remain consistent across finance and planner modules.

## Entities & Data Relationships
- `budgets` define `name`, `budgetType`, category scope (`categoryIds`), transaction type, currency, limit, period, and rollover behavior. Budget rows belong to a `user` and can link to `goals` or `accounts`.
- `budget_entries` connect budgets to specific `transactions`; fields track applied amount (budget currency) and the conversion rate used when the transaction currency differs from the budget currency.

## API Surface & Responsibilities
- GET `/budgets`, `/budgets/:id`, POST `/budgets`, PATCH `/budgets/:id`, DELETE `/budgets/:id`.
- GET `/budgets/:id/transactions` for drill-down, GET `/budgets/:id/entries` for applied entries, POST `/budgets/:id/recalculate` to re-evaluate spent/remaining totals, especially after backfilled transactions.
- When transactions reference the budget, update `spentAmount`, `remainingAmount`, `percentUsed`, `isOverspent`, and optionally send signals to `insights` or `notifications`.

## Real-Time & WebSocket Notes
- Emit `entity:updated` on spend updates so widgets and home dashboards show real-time usage.
- If `notifyOnExceed = true`, trigger `reminder` or `insight:new` events when `percentUsed` crosses critical thresholds (e.g., 90%, 100%).

## Operational Considerations
- Support multi-currency budgets by storing `rateUsedTxnToBudget` and `appliedAmountBudgetCurrency`; recalc endpoint should consider historical FX rates.
- Carryover logic (`rolloverMode`) may run as background jobs; mark budgets appropriately so `sync/pull` can deliver new state quickly.
