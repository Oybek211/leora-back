# Authentication & Users Backend Prompt

## Mission
Provide a secure, multi-tenant foundation for every request by owning user identity, credentials, and profile data. Deliver JWT + refresh token flows, enforce region/currency defaults, and keep each entity keyed by `userId` so the rest of the stack can filter by tenant.

## Entities & Data Relationships
- `users`: the core tenant record; exposes profile, preferences, verification flags, region, and currency. Every other module stores `userId` (FK → `users.id`).
- `user_settings`, `subscriptions`, `devices`, `sessions`, `ai_usage`, etc. all cite the same `userId`, so updates must maintain referential integrity, cascading where necessary.
- Profile preferences (JSON blobs) are the only way to customize notifications, AI settings, and focus defaults without branching into frontend logic.

## API Surface & Backend Responsibilities
- Auth arteries: `/auth/register`, `/auth/login`, `/auth/refresh`, `/auth/forgot-password`, `/auth/reset-password`, `/auth/logout`, `/auth/me`. Handle validation, rate limiting, password hashing, OTP emailing, and token lifecycle stamps with `issuedAt`, `expiresAt`, and `idempotencyKey` support.
- User CRUD: `/users/me` for profile updates; enforce unique email/username constraints; guard region/currency enumeration; update `lastLoginAt` and session metadata.
- Security: coordinate `sessions` and `devices` tables (see Device & Session module) whenever login events occur so that `UserSession` tokens can be revoked or rotated.

## Real-Time & WebSocket Notes
- Emit user presence updates via the `presence:active` client event and tie server events to session renewal; the WebSocket handshake verifies JWTs issued by this module.
- User changes (profile, notification preferences, subscription tier) must broadcast via `entity:updated` events so connected clients show consistent state.

## Operational Considerations
- Rate-limit login/forgot-password flows per `auth` bucket (10 req/min).
- Admin or automation flows respect `showStatus` and soft delete semantics, leaving other modules to honor `syncStatus` for offline sync.
