# Planner Habits Backend Prompt

## Mission
Keep habits as structured routines with reminder rules, completion modes, and optional finance automation so that planning, gamification, and notifications can rely on deterministic streak and completion data.

## Entities & Data Relationships
- `habits` captures metadata (type, frequency, reminders, difficulty, links to `goals`, `showStatus`). Numeric habits store `targetPerDay` and `unit`, while boolean habits only require `completionMode = boolean`.
- `habit_completion_entries` track day-level status (`done`, `miss`) plus optional `value`. Each entry references `habitId` and writes to `streakCurrent`/`streakBest` on inserts.
- `habit_goals` is a pivot table supporting the many-to-many relationship between habits and goals; `linkedGoalIds` in POST payloads materialize into this table.
- Habit finance automation may reference `transactions`, `budgets`, or `accounts` when running rules such as `no_spend_in_categories` or `daily_spend_under`.

## API Surface & Responsibilities
- Endpoints: GET `/habits`, `/habits/:id`, POST `/habits`, PATCH `/habits/:id`, DELETE; POST `/habits/:id/complete` to capture today's completion, GET `/habits/:id/history` and `/habits/:id/stats` for reporting.
- The completion endpoint must respect reminder windows and update streaks/`completionRate30d`. Recording completions should also trigger habit-linked goals and achievements updates.
- Provide habit-level search/sort filters when returning lists; default to tenant-level data and support pagination meta.

## Real-Time & WebSocket Notes
- Emit `entity:updated` for habits and completions so widgets reflecting streaks can refresh. Habits with reminders may also send `reminder` events if `reminderEnabled` and `REMINDER` scheduler identifies a trigger.
- Hook into `insight:new` when a streak milestone is reached or numeric target surpasses a threshold.

## Operational Considerations
- Habit finance rules may exercise the transactions cache for spend detection; guard on performance by delegating heavy calculations to background jobs when scanning multiple accounts or categories.
- Complete entries should be idempotent per `dateKey`; duplicates detected in server should either update existing row or respond with `IDEMPOTENCY_CONFLICT`.
