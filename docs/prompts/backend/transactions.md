# Finance Transactions Backend Prompt

## Mission
Process every financial event (income, expense, transfer) with validation, balance adjustments, and links to budgets, debts, habits, and goals so analytics and automation remain trustworthy.

## Entities & Data Relationships
- `transactions` store type, status, account references, amount, currency, base conversion details, optional `goalId`, `budgetId`, `debtId`, `habitId`, `counterpartyId`, `categoryId`, `subcategoryId`, `attachments`, and tags.
- `transaction_split` lets a single transaction span multiple categories; each row references `transactionId` and defines an `amount` tied to `categoryId`.
- Transactions may spawn or update `budget_entries`, `goal_finance_contributions`, `debt_payments`, `habit_rules`, and `insights`.

## API Surface & Responsibilities
- Standard: GET `/transactions`, `/transactions/:id`, POST `/transactions`, PATCH `/transactions/:id`, DELETE `/transactions/:id`.
- Summary/analytics: GET `/transactions/summary`, `/transactions/analytics` (custom metrics). Summary must roll up totals by type, currency, nets, and category breakdown.
- POST flows: require different payload shapes for `income`, `expense`, `transfer`. Each creation updates associated `accounts` balances, `budgets.spentAmount`, `debts.totalPaidInRepaymentCurrency`, and writes conversion rates (`rateUsedToBase`, `convertedAmountToBase`).
- Filtering must honor tenant scope and include pagination and search by `name`/`description`/`tags`.

## Real-Time & WebSocket Notes
- Emit `entity:updated` or `entity:created` for transactions so dashboards, widgets, and budgets can immediately reflect new amounts.
- Trigger `insight:new` or `reminder` events when unusual spend detected (e.g., overspend) or when a transaction completes a `goal`.

## Operational Considerations
- Keep integrity of `currentBalance` across accounts by wrapping multi-account transfers in transactions; store `idempotencyKey` to prevent double-processing.
- Support conflict resolution via `/sync/resolve-conflict` (preferring latest `updatedAt` while avoiding balance drift).
