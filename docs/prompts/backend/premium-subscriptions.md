# Premium & Subscriptions Backend Prompt

## Mission
Run the subscription lifecycle (tiers, plans, payments, quotas) so tiered feature access, billing integrations, and premium insight quotas stay synchronized.

## Entities & Data Relationships
- `subscriptions` hold tier, status, provider info, current period dates, cancelation flags, and currency/amount metadata. Each subscription links to a `user`.
- `subscription_history` logs events (`created`, `renewed`, `payment_failed`, etc.) for auditing and debugging webhook payloads.
- `premium_features` map features to tiers and limits, which feeds into usage enforcement (e.g., AI quotas, account limits).
- `plans` describe available offerings (products, intervals, feature lists); each plan can surface `features` to be enforced elsewhere.

## API Surface & Responsibilities
- GET `/subscriptions/me` to surface the current subscription with active features and plan meta (name, interval). Use this to calculate limits client-side.
- GET `/subscriptions/plans`, POST `/subscriptions/checkout`, `/subscriptions/verify-receipt`, `/subscriptions/cancel`, `/subscriptions/restore`, `/subscriptions/history`.
- Webhooks: POST `/subscriptions/webhook/stripe`, `/subscriptions/webhook/apple`, `/subscriptions/webhook/google` to react to provider events and update `subscription_history`.
- `/subscriptions/checkout` should branch per provider (Stripe returns `checkoutUrl`, mobile returns `productId` data) while normalizing `sessionId`/`offerId` for reconciliation.

## Real-Time & WebSocket Notes
- When tier or status changes, broadcast `entity:updated` so widgets (home, finance) can adapt limits immediately and `permissions` can expire premium-only flows.

## Operational Considerations
- Gate premium APIs (AI insights, quotas, multi-currency) using `premium_features` limits and `subscription` status; default to `free` tier when no active subscription exists.
- Rate limit AI-centric endpoints (10 req/min) and update `AIUsage` records to prevent quota overages.
