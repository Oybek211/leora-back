# Habit Data Model

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string|null",
  "iconId": "string|null",
  "habitType": "health|finance|productivity|education|personal|custom",
  "status": "active|paused|archived",
  "showStatus": "active|archived|deleted",
  "goalId": "uuid|null",
  "frequency": "daily|weekly|custom",
  "daysOfWeek": [1,2,3,4,5,6,7],
  "timesPerWeek": 0,
  "timeOfDay": "HH:mm|null",
  "completionMode": "boolean|numeric",
  "targetPerDay": 0,
  "unit": "string|null",
  "countingType": "create|quit",
  "difficulty": "easy|medium|hard",
  "priority": "low|medium|high",
  "challengeLengthDays": 0,
  "reminderEnabled": false,
  "reminderTime": "string|null",
  "streakCurrent": 0,
  "streakBest": 0,
  "completionRate30d": 0,
  "financeRule": {
    "type": "no_spend_in_categories|spend_in_categories|has_any_transactions|daily_spend_under",
    "categoryIds": ["string"],
    "accountIds": ["string"],
    "minAmount": 0,
    "amount": 0,
    "currency": "string"
  },
  "linkedGoalIds": ["uuid"],
  "createdAt": "ISO8601",
  "updatedAt": "ISO8601"
}
```

