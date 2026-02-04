# Examples

## POST /accounts
Request:
```json
{
  "name": "Cash",
  "currency": "UZS",
  "accountType": "cash",
  "initialBalance": 250000
}
```
Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "name": "Cash",
    "accountType": "cash",
    "currency": "UZS",
    "initialBalance": 250000,
    "currentBalance": 250000,
    "linkedGoalId": null,
    "customTypeId": null,
    "isArchived": false,
    "showStatus": "active",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": null
}
```

