# Backend Architecture (Frontend-First)

## Scope
This backend is designed to satisfy all existing frontend screens and data flows in the Leora app. It follows the API client contract in `src/services/api` and domain models in `src/domain`.

## Frontend Screen Map and Data Needs
- Home (`app/(tabs)/index.tsx`, `src/hooks/useHomeDashboard.ts`): greeting, daily tasks, goals progress, finance snapshot (income/expense/budget), focus minutes; requires summary endpoints and per-day aggregates.
- Finance Tabs (`app/(tabs)/(finance)/(tabs)/*`):
  - Accounts: list, detail, balance history, account actions.
  - Transactions: list, filters (type, date range, amount range, category), grouping by date, detail view, cancel/delete, edit.
  - Budgets: list, detail, spending breakdown, budget-linked goal progress.
  - Debts: list, detail, payments, settle/extend actions.
  - Analytics: income/expense trends, category breakdown, top categories, budget overage, upcoming debt due.
- Finance Modals (`app/(modals)/finance/*`): add/edit account, transaction, budget, debt; quick expense; filters; export.
- Planner Tabs (`app/(tabs)/(planner)/(tabs)/*`): tasks, goals, habits; detail modals for each; progress calculations.
- Focus (`app/focus-mode.tsx`, `src/services/api/focusService.ts`): session start/complete/pause/resume; stats.
- Insights (`app/(tabs)/(insights)/*`, `src/services/ai/*`): daily/period summaries and Q&A; uses AiResponse schema.
- Auth (`app/(auth)/*`): register, login, forgot/reset password, session refresh.
- More/Profile (`app/(tabs)/more/*`): user profile, settings, notifications, integrations.

## Core Entities and Relationships
- User 1:N Accounts, Transactions, Budgets, Debts, Counterparties, Tasks, Goals, Habits, FocusSessions, Notifications.
- Account 1:N Transactions (income/expense), Transfer Transactions (from/to accounts).
- Transaction N:1 Budget, Goal, Debt, Habit (optional links).
- Budget N:1 Goal (bidirectional link), N:1 Account (optional).
- Debt N:1 Counterparty, N:1 Goal (bidirectional link), N:1 Budget (optional), 1:N DebtPayments.
- Goal 1:N Tasks, Habits, FocusSessions; optional finance link to Budget or Debt.

## Calculations (No-NaN Guarantees)
All numeric fields are always returned as numbers with safe defaults.
- Account.currentBalance: initialBalance + income - expense - transferOut + transferIn.
- Budget.spentAmount: sum of expense transactions in budget period; remainingAmount = limitAmount - spentAmount; percentUsed = limitAmount > 0 ? spentAmount / limitAmount * 100 : 0.
- Debt.remainingAmount: principalAmount - sum(debtPayments convertedAmountToDebt); totalPaid, percentPaid computed server-side.
- Goal.progressPercent and currentValue: computed from finance link (budget/debt) or manual progress.
- Focus stats: total sessions, total minutes, average minutes, interruptions.
- Dashboard summaries: daily/weekly/monthly totals for tasks, habits, finance, focus.

## API Conventions
- Base URL: `/api/v1`.
- Response envelope matches `ApiResponse<T>` in `src/services/api/apiClient.ts`.
- Pagination with `page`, `limit`, `total`, `totalPages` in `meta`.
- Sorting and filtering via query params; all lists support pagination defaults.
- Soft delete with `showStatus: active|archived|deleted`.
- Timestamps are ISO 8601 strings.

## Currency and FX
- All finance amounts stored with currency code and base currency conversion.
- FX rates are stored per transaction (rateUsedToBase, effectiveRateFromTo) to avoid recalculation drift.
- Debt payments store conversion to debt currency and base currency.

## Events and Cross-Module Sync
- Finance events trigger planner updates (goal progress, task auto-complete, habit finance evaluation).
- Planner events may create finance artifacts (budget/debt link).

