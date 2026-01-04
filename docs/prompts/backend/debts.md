# Finance Debts Backend Prompt

## Mission
Manage all lending/borrowing obligations so principal, rate, and payment tracking stays auditable while integrating with transactions, reminders, and goals.

## Entities & Data Relationships
- `debts` capture direction (`i_owe`, `they_owe_me`), principal/pricing info, counterparty metadata, linked accounts/budgets/goals, reminders, and settlement statuses.
- Related tables: `debt_payments` (per-payment ledger) and `debt_payment -> transactions` (one-to-one) ensure financial flow is traceable; `counterpartyId` ties to Counterparties module.
- Debts may produce `insights` (e.g., overdue warnings) and contribute to `achievements` (paying off debts).

## API Surface & Responsibilities
- GET `/debts`, `/debts/:id`, POST `/debts`, PATCH `/debts/:id`, DELETE `/debts/:id`.
- POST `/debts/:id/payments` to log payments; each payment updates `debt_payments`, reduces outstanding principal, updates `totalPaidInRepaymentCurrency`, attaches a transaction if `relatedTransactionId` provided, and respects currency conversions (`rateUsedToBase`).
- GET `/debts/:id/payments`, POST `/debts/:id/settle`, and GET `/debts/summary` (with currency breakdowns, overdue counts) round out the surface.

## Real-Time & WebSocket Notes
- Broadcast `entity:updated` for debts and payments to keep dashboards synced and to allow `insight:new` or `reminder` events for approaching due dates or overdue states.
- Overdue and status transitions (paid/overdue) should also emit new insights or notifications to trigger user actions.

## Operational Considerations
- Enforce `interestRateAnnual` calculations and schedule hints when `interestMode` is `simple` or `compound`, possibly via background jobs that update `debt_payments` or `insights`.
- Soft deletes (`showStatus`) must cascade to `debt_payments` and correlated `transactions` to avoid phantom obligations.
