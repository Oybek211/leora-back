# Notifications & Reminders Backend Prompt

## Mission
Own notification delivery rules (push, email, reminders) derived from user settings, planner/finance events, and AI insights, and ensure real-time delivery when tenants are connected.

## Entities & Data Relationships
- Notification preferences live inside `user_settings.notifications`, detailing categories (finance, tasks, habits, ai), quiet hours, and delivery toggles.
- Notification triggers observe planner entities (`tasks`, `habits`, `goals`), finance modules (`budgets`, `accounts`, `transactions`, `debts`), and `insights`. Backend must evaluate each trigger against user preferences before emitting.

## API Surface & Responsibilities
- Settings endpoints (PATCH `/settings/notifications`) update preferences; use server-side validation to clamp quiet hour ranges and enable/disable categories.
- Notification service consumes domain events (task deadlines, budget overspend, debt reminders, insight creations) and stores metadata (type, `dueIn`, `entityId`) for auditing.
- Provide a delivery queue for push/email; track `lastSentAt`, `retryCount`, and `status` for each notification so retries obey rate and quiet hour constraints.

## Real-Time & WebSocket Notes
- Use WebSocket `reminder` event (with payload `{type, id, title, dueIn}`) to push imminent alerts while the user is connected.
- For non-connected users, fall back to push providers but still honor `notifications` settings (sound, vibration, quiet hours).
- Broadcast `entity:updated` after reminder scheduling so clients can reflect snoozed/handled statuses.

## Operational Considerations
- Integrate with Sync module to keep pending reminders in sync when the client goes offline.
- Keep a configurable threshold (e.g., send reminder 30 minutes before due) and allow adjustments via `/settings/notifications`.
