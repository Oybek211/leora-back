# Finance Counterparties Backend Prompt

## Mission
Model personas or entities tied to debts and transactions so users can track relationships, phone numbers, and notes while maintaining fast searchability.

## Entities & Data Relationships
- `counterparties` store `displayName`, optional `phoneNumber`, `comment`, `searchKeywords`, and `userId` reference.
- This entity links to `debts` via `counterpartyId` and to `transactions`. Debt summaries may aggregate per counterparty (overdue counts, net position).

## API Surface & Responsibilities
- GET `/counterparties`, `/counterparties/:id`, POST `/counterparties`, PATCH `/counterparties/:id`, DELETE `/counterparties/:id`.
- Endpoints `/counterparties/:id/debts` and `/counterparties/:id/transactions` return relational data scoped to the tenant.
- Provide search-friendly fields (`searchKeywords`) and support pagination/filters for names.

## Real-Time & WebSocket Notes
- Emit `entity:updated` when counters update to keep debt overviews or transaction filters reactive.

## Operational Considerations
- Soft delete needs to hide related debts/transactions while keeping historical references intact for reporting.
