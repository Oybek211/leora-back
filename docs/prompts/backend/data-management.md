# Data Management Backend Prompt

## Mission
Protect user data by orchestrating backups, exports, and GDPR-friendly deletions while keeping caches clean and long-running jobs transparent.

## Entities & Data Relationships
- `backups` record type (`manual`, `auto`, `export`), status, storage, file metadata, entity counts, expiration, and included entities.
- `exports` track format (`json`, `csv`, `pdf`), scope (`finance`, `planner`, etc.), status, and file delivery.
- All data management operations reference `userId` and, when relevant, `entitiesIncluded` so restore or export jobs can reconstruct state.

## API Surface & Responsibilities
- POST `/data/backup`, GET `/data/backups`, `/data/backups/:id`, POST `/data/backups/:id/restore`, DELETE `/data/backups/:id`, POST `/data/export`, GET `/data/exports`, GET `/data/exports/:id/download`.
- DELETE `/data/account` handles account deletion with confirmation (e.g., `DELETE MY ACCOUNT`), reason, and compliance logging.
- POST `/data/cache/clear` empties application caches affecting offline sync or analytics without touching user data.

## Real-Time & WebSocket Notes
- Emit `entity:updated` when backups complete or exports are ready so dashboard widgets can surface new files.
- For long-running operations, consider dedicated progress events over WebSocket so the client can show status without polling.

## Operational Considerations
- Backups should optionally run asynchronously and produce metadata with entity counts to aid restores.
- Exports may require background workers due to size; store `expiresAt` and delete stale files automatically.
- GDPR delete should cascade through all modules (tasks, transactions, insights) while keeping audit trails for compliance.
