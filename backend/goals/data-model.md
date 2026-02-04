# Goal Data Model

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string|null",
  "goalType": "financial|health|education|productivity|personal",
  "status": "active|paused|completed|archived",
  "showStatus": "active|archived|deleted",
  "metricType": "none|amount|weight|count|duration|custom",
  "direction": "increase|decrease|neutral",
  "unit": "string|null",
  "initialValue": 0,
  "targetValue": 0,
  "progressTargetValue": 0,
  "currentValue": 0,
  "financeMode": "save|spend|debt_close|null",
  "currency": "string|null",
  "linkedBudgetId": "uuid|null",
  "linkedDebtId": "uuid|null",
  "startDate": "YYYY-MM-DD|null",
  "targetDate": "YYYY-MM-DD|null",
  "completedDate": "YYYY-MM-DD|null",
  "progressPercent": 0,
  "milestones": [
    { "id": "uuid", "title": "string", "targetPercent": 0, "isCompleted": false, "completedAt": "ISO8601|null" }
  ],
  "stats": {
    "totalTasks": 0,
    "completedTasks": 0,
    "totalHabits": 0,
    "focusMinutes": 0
  },
  "createdAt": "ISO8601",
  "updatedAt": "ISO8601"
}
```

