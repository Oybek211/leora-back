# Examples

## POST /transactions
Request:
```json
{
  "type": "expense",
  "accountId": "uuid",
  "amount": 120000,
  "currency": "UZS",
  "categoryId": "food",
  "date": "2024-01-15",
  "description": "Lunch"
}
```
Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "type": "expense",
    "accountId": "uuid",
    "fromAccountId": null,
    "toAccountId": null,
    "amount": 120000,
    "currency": "UZS",
    "baseCurrency": "UZS",
    "rateUsedToBase": 1,
    "convertedAmountToBase": 120000,
    "toAmount": 0,
    "toCurrency": null,
    "effectiveRateFromTo": 1,
    "feeAmount": 0,
    "feeCategoryId": null,
    "categoryId": "food",
    "subcategoryId": null,
    "name": null,
    "description": "Lunch",
    "date": "2024-01-15",
    "time": null,
    "goalId": null,
    "budgetId": null,
    "debtId": null,
    "habitId": null,
    "counterpartyId": null,
    "recurringId": null,
    "attachments": [],
    "tags": [],
    "isBalanceAdjustment": false,
    "skipBudgetMatching": false,
    "showStatus": "active",
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-01-15T10:00:00Z"
  },
  "error": null,
  "meta": null
}
```

