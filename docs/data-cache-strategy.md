# DATA & CACHE STRATEGY  
## PostgreSQL + Redis (Go + Fiber)

You are a **senior backend engineer** building a **production-grade backend**.

This document defines the **ONLY allowed approach** for using PostgreSQL and Redis together.

---

## 1. SOURCE OF TRUTH (ABSOLUTE RULE)

- **PostgreSQL is the ONLY source of truth**
- Redis is **NOT** a database
- The system MUST work correctly even if Redis is fully unavailable

Rule:
> If Redis data is lost, the system must recover fully from PostgreSQL.

---

## 2. ROLE SEPARATION

### PostgreSQL — Persistent & Critical Data
Use PostgreSQL for:
- Users
- Authentication
- Planner (tasks, focus history)
- Finance (transactions, balances, budgets)
- Subscriptions / plans
- Audit & history

Rules:
- Use DB transactions where consistency matters
- Enforce relations with foreign keys
- Never skip a PostgreSQL write in favor of Redis

---

### Redis — Ephemeral / Performance Layer
Use Redis ONLY for:
- Cache (home widgets, summaries, stats)
- Session / token blacklist
- OTP & verification
- Notification queues
- WebSocket presence
- Rate limiting
- Distributed locks

Rules:
- Redis data MUST be rebuildable from PostgreSQL
- Redis keys MUST have TTL
- Redis data loss is acceptable

---

## 3. WRITE FLOW (MANDATORY)

❌ Forbidden:
```
Write to Redis → Write to PostgreSQL
```

✅ Correct:
```
Write to PostgreSQL → Invalidate / update Redis
```

Redis must NEVER be the primary write target.

---

## 4. CACHE INVALIDATION (REQUIRED)

Any of the following actions MUST invalidate related cache:
- CREATE
- UPDATE
- DELETE

Example:
```
transaction created
→ invalidate: finance:summary:user:{user_id}
```

Stale cache is NOT allowed.

---

## 5. MODULE-SPECIFIC RULES

### Planner
- Tasks & focus history → PostgreSQL
- Active focus session → Redis
- Live focus updates → WebSocket + Redis

---

### Finance
- Transactions & balances → PostgreSQL (transactional)
- Exchange rates → Redis (cached)
- Daily / monthly summaries → Redis

---

### Widgets / Home
- Raw data → PostgreSQL
- Aggregated snapshot → Redis
- Live updates → WebSocket (Redis pub/sub)

---

### Notifications
- Stored notifications → PostgreSQL
- Unread counters → Redis
- Delivery queue → Redis
- Push via WebSocket

---

## 6. LOCKING & CONCURRENCY

- Prefer PostgreSQL transactions for consistency
- Use Redis locks ONLY for coordination
- Locks MUST be short-lived (TTL required)

---

## 7. REDIS KEY NAMING STANDARD

Keys MUST follow:
```
{module}:{entity}:{scope}:{id}
```

Examples:
```
planner:focus:user:{user_id}
finance:summary:user:{user_id}
notifications:unread:user:{user_id}
widgets:home:user:{user_id}
```

---

## 8. FAILURE TOLERANCE

- Redis downtime MUST NOT corrupt data
- Redis downtime MUST NOT block critical flows
- Backend must gracefully degrade without Redis

---

## 9. PROHIBITED PRACTICES

❌ Storing money or balances in Redis  
❌ Using Redis as a primary datastore  
❌ Eventual sync from Redis to PostgreSQL  
❌ Long-lived Redis keys without TTL  
❌ Silent cache misses  

---

## 10. ENGINEERING EXPECTATION

- Favor correctness over performance
- Assume Redis can disappear at any time
- Document assumptions clearly
- Write production-grade code only

FOLLOW THIS DOCUMENT STRICTLY.
