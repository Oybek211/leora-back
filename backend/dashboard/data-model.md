# Dashboard Data Models

## Summary Response
```json
{
  "date": "YYYY-MM-DD",
  "progress": {
    "tasks": 0,
    "budget": 0,
    "focus": 0
  },
  "counts": {
    "tasksDue": 0,
    "habitsDue": 0,
    "goalsActive": 0,
    "transactions": 0
  },
  "finance": {
    "income": 0,
    "expense": 0,
    "net": 0,
    "currency": "string"
  }
}
```

## Calendar Indicators
```json
{
  "2024-01-01": {
    "progress": { "tasks": 0, "budget": 0, "focus": 0 },
    "events": { "tasks": 0, "habits": 0, "goals": 0, "finance": 0 }
  }
}
```

