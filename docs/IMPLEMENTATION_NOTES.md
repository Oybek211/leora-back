# Implementation Notes / Deviations

## Extra Endpoints Not in Backend Doc
These endpoints are currently implemented but not documented in `backend/*`:
- `/admin/*` (role management)
- `/subscriptions/*` and `/subscriptions/plans/*`
- `/widgets/*`
- `/search`
- `/settings/*`
- `/sync/*`
- `/devices/*`
- `/integrations/*`
- `/achievements/*`
- `/ai/*`
- GET `/insights`

These are preserved for backward compatibility with existing clients.
