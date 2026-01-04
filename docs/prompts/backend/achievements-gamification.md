# Achievements & Gamification Backend Prompt

## Mission
Power XP, achievements, and level tracking so motivational loops stay consistent across planner, finance, and focus actions without relying on UI cues.

## Entities & Data Relationships
- `achievements` define static metadata (`key`, `name`, `description`, `category`, `requirement`, `xpReward`, `tier`, `order`). They are preloaded and reference categories (`finance`, `tasks`, `habits`, etc.).
- `user_achievements` tracks per-user progress, unlocked state, and notification timestamps.
- `user_levels` maintains XP totals, current level, and titles derived from XP rules (e.g., `levelFormula = floor(totalXP / 500) + 1`).
- XP events originate from planner/finance actions (tasks completed, habits streak, transactions) and may feed into achievements and insights.

## API Surface & Responsibilities
- GET `/achievements`, `/achievements/:id`, `/achievements/categories`, `/users/me/achievements` to expose available achievements and user progress.
- GET `/users/me/level` returns `level`, `title`, `currentXP`, `totalXP`, `xpForNextLevel`, `xpProgress`, and recent XP gains.
- POST `/achievements/:id/claim` handles reward capture when achievements require explicit claiming; ensure claimant still meets requirements.

## Real-Time & WebSocket Notes
- Emit `entity:updated` for `user_achievements` and `user_levels` so widgets/leaderboards refresh XP counters. New achievement unlocks can trigger `insight:new` or `reminder` style motivational push.

## Operational Considerations
- XP calculations should be deterministic; store XP events in hashed logs to prevent duplicates and support `sync`.
- Provide batching for awarding XP when multiple triggers occur simultaneously (e.g., completing a task also completes a habit and job). Avoid double counting.
