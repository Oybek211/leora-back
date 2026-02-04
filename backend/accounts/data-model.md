# Account Data Model

```json
{
  "id": "uuid",
  "userId": "uuid",
  "name": "string",
  "accountType": "cash|card|savings|investment|credit|debt|other",
  "currency": "string",
  "initialBalance": 0,
  "currentBalance": 0,
  "linkedGoalId": "uuid|null",
  "customTypeId": "string|null",
  "isArchived": false,
  "showStatus": "active|archived|deleted",
  "createdAt": "ISO8601",
  "updatedAt": "ISO8601"
}
```

Notes:
- `currentBalance` is calculated server-side; never null.

