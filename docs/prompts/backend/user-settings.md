# User Settings Backend Prompt

## Mission
Centralize theme, language, notifications, focus, security, and AI preferences so that every client can build consistent UX without embedding business rules.

## Entities & Data Relationships
- `user_settings` is uniquely keyed by `userId` and contains `theme`, `language`, `notifications`, `security`, `ai`, `focus`, and `privacy` JSON blobs.
- Notification settings relay into the Notifications module; security settings tie to `devices`/`sessions`; AI settings influence `insights` and `ai_usage`; focus settings inform `focus_sessions` defaults.

## API Surface & Responsibilities
- GET `/settings` returns the complete settings package.
- PATCH `/settings` (and scoped endpoints `/settings/notifications`, `/settings/security`, `/settings/ai`, `/settings/focus`) update the respective JSON structures. Each patch should validate fields (e.g., notification quiet hours, AI feature toggles) before persisting.
- POST `/settings/reset` resets to defaults while respecting `premium` feature access (e.g., only premium tiers get `mentorAdvices`).

## Real-Time & WebSocket Notes
- When settings change, emit `entity:updated` so subscribed clients can rehydrate local caches (especially for notifications/AI features).
- Some scope-specific toggles may trigger immediate recalculations (e.g., turning on `smartReminders` should inform the reminders scheduler to start or stop streaming `reminder` events).

## Operational Considerations
- Keep settings updates idempotent and conflict-safe; use `syncStatus` to propagate offline modifications.
- Settings patches might indirectly affect other modules (e.g., disabling `voiceRecognition` should mute voice input queues). Coordinate via eventing or background jobs.
