# Verification Checklist

## Setup
- Run migrations: `./bin/migrate up` (or `go run ./cmd/server` with migrations enabled)
- Start server: `go run ./cmd/server`

## Auth
- Register: `curl -X POST http://localhost:9090/api/v1/auth/register -H 'Content-Type: application/json' -d '{"email":"test@leora.app","fullName":"Test User","password":"pass1234","confirmPassword":"pass1234","region":"UZ","currency":"UZS"}'`
- Login: `curl -X POST http://localhost:9090/api/v1/auth/login -H 'Content-Type: application/json' -d '{"emailOrUsername":"test@leora.app","password":"pass1234"}'`
- Use `Authorization: Bearer <accessToken>` for all protected routes

## Finance
- Accounts CRUD + subresources:
  - `GET /accounts`
  - `POST /accounts`
  - `GET /accounts/:id`
  - `GET /accounts/:id/transactions`
  - `GET /accounts/:id/balance-history`
- Transactions:
  - `GET /transactions?type=expense&dateFrom=2024-01-01&dateTo=2024-01-31`
  - `POST /transactions`
  - `POST /transactions/transfer`
  - `POST /transactions/bulk`
- Budgets:
  - `GET /budgets?periodType=monthly`
  - `GET /budgets/:id/transactions`
  - `GET /budgets/:id/spending`
  - `POST /budgets/:id/recalculate`
- Debts + payments:
  - `POST /debts`
  - `POST /debts/:id/payments`
  - `POST /debts/:id/settle`
  - `POST /debts/:id/extend`
- Counterparties + FX:
  - `GET /counterparties?search=ali`
  - `GET /fx/rates?from=USD&to=UZS&date=2024-01-15`

## Planner
- Tasks:
  - `GET /tasks?status=planned&goalId=<id>`
  - `POST /tasks/:id/complete`
- Goals:
  - `GET /goals/:id/stats`
  - `GET /goals/:id/tasks`
  - `GET /goals/:id/habits`
- Habits:
  - `POST /habits/evaluate-all-finance`
- Focus:
  - `GET /focus-sessions?status=completed`

## Dashboard / Home
- `GET /dashboard/summary?date=2024-01-15`
- `GET /dashboard/calendar?from=2024-01-01&to=2024-01-31`
- `GET /home`

## Reports
- `GET /reports/finance/summary?from=2024-01-01&to=2024-01-31`
- `GET /reports/finance/categories?from=2024-01-01&to=2024-01-31`
- `GET /reports/finance/cashflow?from=2024-01-01&to=2024-01-31&granularity=day`
- `GET /reports/planner/productivity?from=2024-01-01&to=2024-01-31`

## Insights
- `POST /insights/daily`
- `POST /insights/period`
- `POST /insights/qa`
- `POST /insights/voice`
