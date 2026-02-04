# Examples

## GET /dashboard/summary?date=2024-01-15
```json
{
  "success": true,
  "data": {
    "date": "2024-01-15",
    "progress": { "tasks": 60, "budget": 45, "focus": 30 },
    "counts": { "tasksDue": 5, "habitsDue": 3, "goalsActive": 2, "transactions": 4 },
    "finance": { "income": 500000, "expense": 320000, "net": 180000, "currency": "UZS" }
  },
  "error": null,
  "meta": null
}
```

