# Backend Doc Alignment Changelog

## Summary
Aligned API routes, DTOs, error handling, and database schema with the Backend Doc across Finance, Planner, Dashboard/Home, Reports, and Insights.

## Added Endpoints
- Finance: account subresources, budget subresources, debt payments/settle/extend, counterparties, FX rates
- Dashboard/Home: summary, widgets, calendar
- Reports: finance summary/categories/cashflow/debts, planner productivity, insights context
- Insights: daily/period/qa/voice
- Goals: stats/tasks/habits
- Habits: evaluate-all-finance

## DTO Updates
- Expanded finance DTOs to include all documented fields and computed values
- Updated goal stats to match documented fields
- Normalized error envelope and validation error types

## Validation & Error Handling
- Added request validation for finance, debt payments, counterparties
- Standardized error responses with HTTP status codes and `VALIDATION` type

## Data & Migrations
- Added `migrations/009_finance_alignment.sql` to expand finance schema
- Added tables: `counterparties`, `fx_rates`, `debt_payments`
- Added finance columns for accounts, transactions, budgets, debts

## Behavior Fixes
- Implemented budget/debt rollups and account balances
- Implemented list filters for goals, habits, focus sessions, transactions, budgets, debts

## Breaking Changes
- None intended. Existing endpoints remain, but error type for validation is now `VALIDATION` and error `code` matches HTTP status.
