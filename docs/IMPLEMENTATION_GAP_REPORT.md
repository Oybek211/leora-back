# Implementation Gap Report

## Summary
This report compares the backend docs under `backend/*` with the previously implemented API surface. Gaps are grouped by severity.

## BLOCKER
- Missing Finance sub-resources and FX/Counterparty APIs
  - Doc requires `/accounts/:id/transactions`, `/accounts/:id/balance-history`, `/budgets/:id/transactions`, `/budgets/:id/spending`, `/budgets/:id/recalculate`, `/debts/:id/payments`, `/debts/:id/settle`, `/debts/:id/extend`, `/counterparties/*`, `/fx/*`.
  - Code only implemented base CRUD for accounts/transactions/budgets/debts.
- Missing Dashboard, Home, Reports, Insights endpoints
  - Doc requires `/dashboard/*`, `/home/*`, `/reports/*`, `/insights/*`.
  - Code had no such modules (only a sample GET `/insights`).
- Finance DTOs and schema mismatch
  - Doc requires full finance data-models (currency conversions, links, balances, showStatus, etc.).
  - Code used minimal structs and tables (missing columns and computed fields).
- Error envelope mismatch
  - Doc requires `success/data/error/meta` with `VALIDATION` type and `code`=HTTP status.
  - Code omitted `data/meta` on errors and used `BAD_REQUEST` types.

## MAJOR
- Goal endpoints missing
  - Doc requires `/goals/:id/stats`, `/goals/:id/tasks`, `/goals/:id/habits`.
  - Code implemented only core CRUD and finance link endpoints.
- Habit endpoint missing
  - Doc requires `/habits/evaluate-all-finance`.
  - Code only implemented per-habit evaluate.
- List filters not enforced
  - Doc specifies filter params for goals, habits, focus sessions, budgets, debts, transactions.
  - Code ignored several filters.
- Debt payment calculations
  - Doc expects remaining/paid/percent to be computed server-side.
  - Code stored only a `balance` field and did not compute rollups.

## MINOR
- Extra endpoints not in doc
  - `/admin/*`, `/subscriptions/*`, `/widgets/*`, `/search`, `/settings`, `/sync`, `/devices`, `/integrations`, `/achievements`, `/ai`, GET `/insights`.
- Extra HTTP methods
  - PUT routes exist for some resources not mentioned in doc.

## Resolutions Implemented
- Added missing endpoints and handlers for all doc-required routes.
- Expanded finance data models and database schema to match doc fields.
- Added debt payments, counterparties, FX tables and handlers.
- Implemented dashboard/home/reports/insights modules.
- Standardized error envelopes with `VALIDATION` type and HTTP status `code`.
- Added list filtering for goals, habits, focus sessions, transactions, budgets, debts.
- Added computed fields for balances, budgets, debts, and payment rollups.

## Remaining Deviations
- Extra endpoints (admin, subscriptions, widgets, search, settings, sync, devices, integrations, achievements, ai, GET `/insights`) remain available but are not in the Backend Doc.
- Some advanced UI fields (e.g., goal milestone extras) may return additional properties beyond the minimal doc schema.
