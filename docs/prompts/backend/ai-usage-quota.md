# AI Quota & Usage Tracking Backend Prompt

## Mission
Track AI channel consumption so quota enforcement and analytics can distinguish between tiers and keep requests accountable.

## Entities & Data Relationships
- `ai_usage` records each request (user, channel, requestType, tokensUsed, success, responseTime, metadata).
- `ai_quota` tracks per-user/channel limits over periods (`periodStart`, `periodEnd`, `limit`, `used`). Link quotas to `subscriptions` to determine available allowances.

## API Surface & Responsibilities
- GET `/ai/quota` returns tier-specific remaining quotas, reset timestamps, and channel breakdowns (e.g., `daily`, `qa`, `voice`).
- GET `/ai/usage/history`, `/ai/usage/stats` deliver usage timelines and aggregated metrics (`totalRequests`, `avgResponseTime`).
- Each AI request (insights generation, Q&A, voice command) must increment `ai_usage`, update `ai_quota.used`, and reject or queue the request if limits (including `-1` for unlimited) are exceeded.

## Real-Time & WebSocket Notes
- Emit `entity:updated` for quota records so clients can disable UI controls when quotas fall to zero. This is especially important during speech or AI chat flows that rely on live counters.

## Operational Considerations
- Default quotas depend on tier (`free`: limited, `premium`: unlimited). Use `ai_quota` to track resets (daily/period).
- Provide stats for both `thisMonth` and `allTime`, ensuring anonymized per-user logs for compliance.
