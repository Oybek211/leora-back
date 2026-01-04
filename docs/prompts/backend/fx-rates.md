# Finance FX Rates Backend Prompt

## Mission
Serve authoritative FX data for every currency conversion so transactions, budgets, debts, and insights can rely on consistent base values.

## Entities & Data Relationships
- `fx_rates` holds source (`cbu`, `market_api`, etc.), nominal, spread, and optional `rateBid`/`rateAsk`. Each row references `fromCurrency` and `toCurrency` with valid date windows.
- Multiple modules reference FX data for conversions (`transactions`, `budgets`, `debts`, `accounts`, `insights`), so this module needs to expose both historical lookups and latest snapshots.

## API Surface & Responsibilities
- GET `/fx-rates` (with filters like `fromCurrency`, `toCurrency`, `date`, range, `source`), GET `/fx-rates/latest`, GET `/fx-rates/history`, GET `/fx-rates/convert`, POST `/fx-rates` (admin override).
- Latest endpoint returns base currency and rate map; convert endpoint calculates results using selected rate type (`mid` default).
- When new rates are inserted, optionally trigger recalculations by pushing background jobs to update caches used by `transactions`/`debts`.

## Real-Time & WebSocket Notes
- While FX is mostly REST-based, clients subscribed to finance widgets may benefit from `entity:updated` events when `fx_rates` change, allowing dashboards to refresh totals in base currency.

## Operational Considerations
- Provide TTL for cached rates (e.g., `updatedAt` timestamp) and allow `isOverridden` flag for manual corrections.
- Ensure rate insertion respects uniqueness per date + currency pair to avoid conflicting conversions.
