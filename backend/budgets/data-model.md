# Budget Data Model

```json
{
  "id": "uuid",
  "userId": "uuid",
  "name": "string",
  "budgetType": "category|project",
  "categoryIds": ["string"],
  "linkedGoalId": "uuid|null",
  "accountId": "uuid|null",
  "transactionType": "income|expense|null",
  "currency": "string",
  "limitAmount": 0,
  "periodType": "none|weekly|monthly|custom_range",
  "startDate": "YYYY-MM-DD|null",
  "endDate": "YYYY-MM-DD|null",
  "spentAmount": 0,
  "remainingAmount": 0,
  "percentUsed": 0,
  "isOverspent": false,
  "rolloverMode": "none|carryover",
  "notifyOnExceed": false,
  "contributionTotal": 0,
  "currentBalance": 0,
  "isArchived": false,
  "showStatus": "active|archived|deleted",
  "createdAt": "ISO8601",
  "updatedAt": "ISO8601"
}
```

Notes:
- `spentAmount`, `remainingAmount`, `percentUsed`, `isOverspent` are calculated server-side.

