# Finance Accounts Backend Prompt

## Mission
Represent every financial container (cash, cards, savings, investment, debt) so user balances and currency conversions stay accurate, and downstream modules can rely on canonical balances.

## Entities & Data Relationships
- `accounts` contain `name`, `accountType`, `currency`, `initialBalance`, `currentBalance`, optional color/icon, and ties to `goals` or `debts` through `linkedGoalId`. Every account belongs to a `user`.
- Transactions reference accounts (`accountId`, `fromAccountId`, `toAccountId`, `fundingAccountId`, etc.) to adjust balances. Account deletions must cascade to `transactions` with soft delete semantics.

## API Surface & Responsibilities
- GET `/accounts`, `/accounts/:id`; POST `/accounts`; PATCH `/accounts/:id`; DELETE `/accounts/:id`; PATCH `/accounts/:id/adjust-balance`; GET `/accounts/summary`.
- Summary endpoint aggregates balance per currency and distills per-type counts. `adjust-balance` updates `currentBalance`, writes an audit reason, and creates or tags a balancing transaction with `isBalanceAdjustment = true`.
- When creating or deleting accounts ensure `transactions`, `budgets`, `debts`, and `insights` referencing the account adjust or mark stale.

## Real-Time & WebSocket Notes
- Emit `entity:updated` when balances change (new transactions, adjustments) so dashboards and widgets display instant totals.
- Consider `reminder` or `insight:new` triggers when balance thresholds cross thresholds configured in budgets or alerts.

## Operational Considerations
- Multi-currency: store `baseCurrency`, `rateUsedToBase`, and `convertedAmountToBase` when transactions impact accounts to keep cross-user reporting consistent.
- When summing `currentBalance`, include soft-deleted accounts only if `showStatus` is `active` or if a hidden audit view requires it.
