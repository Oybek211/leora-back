# Task Data Model

```json
{
  "id": "uuid",
  "title": "string",
  "status": "inbox|planned|in_progress|completed|canceled|moved|overdue",
  "showStatus": "active|archived|deleted",
  "priority": "low|medium|high",
  "goalId": "uuid|null",
  "habitId": "uuid|null",
  "financeLink": "record_expenses|pay_debt|review_budget|transfer_money|none|null",
  "progressValue": 0,
  "progressUnit": "string|null",
  "dueDate": "YYYY-MM-DD|null",
  "startDate": "YYYY-MM-DD|null",
  "timeOfDay": "HH:mm|null",
  "estimatedMinutes": 0,
  "energyLevel": 0,
  "context": "string|null",
  "notes": "string|null",
  "lastFocusSessionId": "uuid|null",
  "focusTotalMinutes": 0,
  "checklist": [
    { "id": "uuid", "taskId": "uuid", "title": "string", "completed": false, "order": 0 }
  ],
  "dependencies": [
    { "id": "uuid", "dependsOnTaskId": "uuid" }
  ],
  "createdAt": "ISO8601",
  "updatedAt": "ISO8601"
}
```

