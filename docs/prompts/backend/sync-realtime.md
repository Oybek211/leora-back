# Sync & Realtime Backend Prompt

## Mission
Enable offline-first clients by handling bidirectional delta sync and realtime broadcasts so every module can stay consistent across devices.

## Entities & Data Relationships
- All entities share sync metadata (`syncStatus`, `localUpdatedAt`, `serverUpdatedAt`, `idempotencyKey`). Maintain per-entity `sync_status` state and ensure conflict logs capture `userId`, `entity`, and operations.
- Sync pulls/pushes and WebSocket broadcasts must respect `userId` tenancy and `showStatus` values.

## API Surface & Responsibilities
- POST `/sync/push`: accept batched change sets (entity, operation, data). Validate idempotency, resolve references (tasks → goals, transactions → accounts), and respond with updated timestamps.
- GET `/sync/pull`: return server-side changes since a provided timestamp including `serverTime`. Provide optional `lastSyncedAt` to help clients align.
- POST `/sync/resolve-conflict`: accept resolution directives (`use_server`, `use_local`, `merge`) and apply atomic fixes.

## Real-Time & WebSocket Notes
- WebSocket connection (`ws://.../ws?token=JWT`) pushes server events (`entity:updated`, `insight:new`, `reminder`) when relevant modules change. Clients can subscribe via `subscribe` event to limit the stream.
- Accept client events like `presence:active` and `subscribe` to track online status and deliver targeted updates only when needed.
- When the server broadcasts an entity change, clients can skip `/sync/pull` for that entity, reducing contention.

## Operational Considerations
- Persist `syncStatus` (`local`, `pending`, `synced`, `conflict`) per record to support offline editing and background reconciliation.
- Rate limit sync endpoints (20 req/min) and batch operations to reduce load; rely on `syncStatus` to identify un-synced records for background workers.
