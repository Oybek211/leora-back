# Backend Doc Summary

## Base URL and Auth
- Base URL: `/api/v1`
- Auth: JWT access token in `Authorization: Bearer <token>`
- Refresh: POST `/auth/refresh`, refresh token rotated
- All responses use `ApiResponse<T>` envelope

## Error Model
```json
{
  "success": false,
  "data": null,
  "error": { "code": 400, "message": "Validation failed", "type": "VALIDATION" },
  "meta": null
}
```
- Types: `VALIDATION`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `RATE_LIMITED`, `INTERNAL`
- IDs are UUIDv4; dates are ISO 8601; numeric fields default to 0

## Endpoints (Doc)
### Auth
- POST `/auth/register`
- POST `/auth/login`
- GET `/auth/me`
- POST `/auth/forgot-password`
- POST `/auth/reset-password`
- POST `/auth/refresh`
- POST `/auth/logout`

### Users
- GET `/users/me`
- PATCH `/users/me`
- GET `/users/:id` (admin)

### Accounts
- GET `/accounts`
- POST `/accounts`
- GET `/accounts/:id`
- PATCH `/accounts/:id`
- DELETE `/accounts/:id`
- GET `/accounts/:id/transactions`
- GET `/accounts/:id/balance-history`

### Transactions
- GET `/transactions`
- POST `/transactions`
- GET `/transactions/:id`
- PATCH `/transactions/:id`
- DELETE `/transactions/:id`
- POST `/transactions/transfer`
- POST `/transactions/bulk`

### Budgets
- GET `/budgets`
- POST `/budgets`
- GET `/budgets/:id`
- PATCH `/budgets/:id`
- DELETE `/budgets/:id`
- GET `/budgets/:id/transactions`
- GET `/budgets/:id/spending`
- POST `/budgets/:id/recalculate`

### Debts + Payments
- GET `/debts`
- POST `/debts`
- GET `/debts/:id`
- PATCH `/debts/:id`
- DELETE `/debts/:id`
- GET `/debts/:id/payments`
- POST `/debts/:id/payments`
- PATCH `/debts/:id/payments/:paymentId`
- DELETE `/debts/:id/payments/:paymentId`
- POST `/debts/:id/settle`
- POST `/debts/:id/extend`

### Counterparties
- GET `/counterparties`
- POST `/counterparties`
- GET `/counterparties/:id`
- PATCH `/counterparties/:id`
- DELETE `/counterparties/:id`
- GET `/counterparties/:id/debts`
- GET `/counterparties/:id/transactions`

### FX
- GET `/fx/rates`
- POST `/fx/rates/manual`
- GET `/fx/supported-currencies`

### Planner
- GET `/tasks`
- POST `/tasks`
- GET `/tasks/:id`
- PATCH `/tasks/:id`
- PUT `/tasks/:id`
- DELETE `/tasks/:id`
- POST `/tasks/:id/complete`
- POST `/tasks/:id/reopen`
- PATCH `/tasks/:taskId/checklist/:itemId`
- POST `/tasks/check-finance-trigger`

- GET `/goals`
- POST `/goals`
- GET `/goals/:id`
- PATCH `/goals/:id`
- DELETE `/goals/:id`
- GET `/goals/:id/stats`
- GET `/goals/:id/tasks`
- GET `/goals/:id/habits`
- POST `/goals/:id/link-budget`
- DELETE `/goals/:id/unlink-budget`
- POST `/goals/:id/link-debt`
- DELETE `/goals/:id/unlink-debt`
- GET `/goals/:id/finance-progress`

- GET `/habits`
- POST `/habits`
- GET `/habits/:id`
- PATCH `/habits/:id`
- DELETE `/habits/:id`
- POST `/habits/:id/complete`
- GET `/habits/:id/history`
- GET `/habits/:id/stats`
- POST `/habits/:id/evaluate-finance`
- POST `/habits/evaluate-all-finance`

- GET `/focus-sessions`
- POST `/focus-sessions/start`
- GET `/focus-sessions/:id`
- PATCH `/focus-sessions/:id/pause`
- PATCH `/focus-sessions/:id/resume`
- POST `/focus-sessions/:id/complete`
- POST `/focus-sessions/:id/cancel`
- POST `/focus-sessions/:id/interrupt`
- GET `/focus-sessions/stats`
- DELETE `/focus-sessions/:id`

### Dashboard / Home
- GET `/dashboard/summary`
- GET `/dashboard/widgets`
- GET `/dashboard/calendar`
- GET `/home`
- GET `/home/widgets`
- GET `/home/calendar`

### Reports
- GET `/reports/finance/summary`
- GET `/reports/finance/categories`
- GET `/reports/finance/cashflow`
- GET `/reports/finance/debts`
- GET `/reports/planner/productivity`
- GET `/reports/insights/daily`
- GET `/reports/insights/period`

### Insights
- POST `/insights/daily`
- POST `/insights/period`
- POST `/insights/qa`
- POST `/insights/voice`

## Data Contracts (Sources)
- Accounts: `backend/accounts/data-model.md`
- Transactions: `backend/transactions/data-model.md`
- Budgets: `backend/budgets/data-model.md`
- Debts + Payments: `backend/debts/data-model.md`, `backend/debt-transactions/data-model.md`
- Counterparties: `backend/counterparties/data-model.md`
- FX Rates: `backend/fx/data-model.md`
- Planner Tasks/Goals/Habits/Focus: `backend/tasks/data-model.md`, `backend/goals/data-model.md`, `backend/habits/data-model.md`, `backend/focus-sessions/data-model.md`
- Dashboard/Home/Reports: `backend/dashboard/data-model.md`, `backend/home/data-model.md`, `backend/reports/data-model.md`
- Users/Auth: `backend/users/data-model.md`, `backend/auth.md`
