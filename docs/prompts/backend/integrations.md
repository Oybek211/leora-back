# Integrations Backend Prompt

## Mission
Manage third-party connections (calendars, banks, apps, devices) so Leora can sync external data, augment insights, and honor user consent and OAuth flows.

## Entities & Data Relationships
- `integrations` store provider, category, tokens (encrypted), scope, account metadata, last sync info, and status. Each row belongs to `userId`.
- `integration_sync_logs` track sync direction, status, item counts, and errors for auditing and retries.
- Connected providers may feed data into planner (calendar events), finance (bank transactions), and habit focus contexts.

## API Surface & Responsibilities
- GET `/integrations`, GET `/integrations/:provider/connect`, POST `/integrations/:provider/callback`, DELETE `/integrations/:id`, POST `/integrations/:id/sync`, GET `/integrations/:id/logs`.
- OAuth flows kick off with `/connect` (returning `authUrl`, `state`) and complete via `/callback` (handling `code`, `state`). Store tokens and schedule periodic syncs.
- Manual `/sync` endpoint triggers immediate data pull/push and writes to `integration_sync_logs`.

## Real-Time & WebSocket Notes
- Send `entity:updated` or dedicated integration events when new data arrives from providers so widgets and dashboards adjust instantly.
- Use WebSocket to signal sync completion (e.g., `integration:sync-success`) if the client is subscribed.

## Operational Considerations
- Handle token refresh and expiry; when refresh fails, emit `entity:updated` so UI can show a reconnect prompt.
- Maintain per-provider rate limits and log errors in `integration_sync_logs` for troubleshooting.
