# Device & Session Management Backend Prompt

## Mission
Track user devices and session tokens so multi-device security, token revocation, and presence reporting work reliably across auth flows.

## Entities & Data Relationships
- `user_devices` record device metadata (type, OS/app version, push token, last activity) and trust status.
- `user_sessions` link to devices and store token hashes, expiration, `isActive`, and revocation timestamps.
- Both tables reference `userId`, enabling queries for all active tokens and devices per tenant.

## API Surface & Responsibilities
- GET `/devices`, DELETE `/devices/:id`, POST `/devices/:id/trust` (mark device as trusted).
- GET `/sessions`, DELETE `/sessions/:id`, DELETE `/sessions/all` (except current) to manage token revocation.
- Device changes should cascade to session invalidation when needed (e.g., deleting a device revokes its session).

## Real-Time & WebSocket Notes
- Emit `entity:updated` for devices and sessions when trust state or last-used timestamps change so clients can surface the device list.
- Tie session creation to `presence:active` events via WebSocket to know when tokens are used live.

## Operational Considerations
- Hash tokens before storing and enforce expiration/refresh boundaries; track `lastActiveAt` for anomaly detection.
- Provide audit logs so admins can review which devices have been revoked or marked untrusted.
