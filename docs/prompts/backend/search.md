# Search Backend Prompt

## Mission
Provide a scalable search layer that surfaces tasks, transactions, counterparties, and insights without embedding UI logic; the module should support keyword, filter, and sort combinations while remaining tenant-aware.

## Entities & Data Relationships
- Search touches `tasks` (title), `transactions` (name, description, tags), `counterparties` (displayName, searchKeywords), and `insights` (title/body). Each search query must include `userId` scoping.
- Optionally, maintain lightweight search indexes or leverage PostgreSQL full-text indexes on the relevant columns to keep query latency acceptable.

## API Surface & Responsibilities
- Extend GET `/tasks`, `/transactions`, `/counterparties`, and `/insights` with `search` query parameters; support general `sortBy`, `sortOrder`, `filters`, and pagination meta.
- Provide a dedicated `/search` aggregator endpoint if required, returning grouped results per module for the dashboard widget.
- Ensure `search` respects rate limits and sanitizes inputs to avoid injection attacks.

## Real-Time & WebSocket Notes
- When indexed entities update, emit `entity:updated` (or an `entity:search-index` event if separate) so clients can refresh cached search results, especially on widgets that show live counts.

## Operational Considerations
- Debounce or batch heavy search updates during high traffic; consider caching search results per user and invalidating on `entity:updated` events.
